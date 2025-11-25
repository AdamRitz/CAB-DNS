package main
//  DNS 实现部分
import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/miekg/dns"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"io"
	"log"
	"math/big"
	"strconv"
)

var db, err = openDBWithoutCache("E:\\Code\\Test\\TestLevelDB3")
var c = 1

type Record struct {
	Sig string `json:"sig"`
	K   string `json:"k"`
}

type RequestMessage struct {
	Context string           `json:"context"`
	Mw      string           `json:"Mw"`
	ProxySK *big.Int         `json:"ProxySK"`
	Kx      *big.Int         `json:"Kx"`
	Ky      *big.Int         `json:"Ky"`
	Sig     SchnorrSignature `json:"sig"`
}

type ResponseMessage struct {
	Mw      string           `json:"Mw"`
	ProxySK *big.Int         `json:"ProxySK"`
	Sig     SchnorrSignature `json:"sig"`
}

func openDBWithoutCache(path string) (*leveldb.DB, error) {
	options := &opt.Options{
		BlockCacheCapacity: 0,       // 禁用块缓存
		WriteBuffer:        1 << 10, // 设置非常小的写缓冲区
	}
	db, err := leveldb.OpenFile(path, options)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	return db, nil
}
func main() {
	StartDNS()

}

func GenKey(i int) string {
	h := sha256.Sum256([]byte("lawwason" + strconv.Itoa(i)))
	return hex.EncodeToString(h[:])
}

func BuildDNSDB() {
	for i := 0; i <= 100000; i++ {
		input := "lawwason" + strconv.Itoa(i)
		Put(input, "{\"context\":\"I love eating fish and pie\",\"Mw\":\"9.9.9.9\",\"ProxySK\":62650150199320531305115128524366440311131353102981565795712617380112966057267,\"Kx\":41158195443767629831988007221955165994994933369172347261641553087616513910773,\"Ky\":114062706648825334784521677113637540754140627679333200572502355619834317149703}")
	}
}

func StartDNS() {
	fmt.Println("StartDNS")
	dns.HandleFunc(".", handleHybridDNSRequest)
	fmt.Println("StarPort: 53")
	server := &dns.Server{Addr: ":5653", Net: "udp"}
	server.ListenAndServe()

}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	name := r.Question[0].Name[0 : len(r.Question[0].Name)-1]
	// 创建一个 goroutine 来处理数据库查询
	go func() {
		m := new(dns.Msg)
		m.SetReply(r)
		m.Authoritative = true

		for _, q := range r.Question {
			switch q.Qtype {
			case dns.TypeTXT:
				// 异步查询数据库
				data := Read(name)

				var segments []string
				segments = append(segments, string(data))
				rr := &dns.TXT{
					Hdr: dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    3600,
					},
					Txt: segments,
				}
				m.Answer = append(m.Answer, rr)

			}
		}
		// 向客户端发送响应
		w.WriteMsg(m)
	}()
}

var rsaPublicPEM = []byte(`
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArQmKREH7t1fHBkYpm2eP
xLiRcoE2W5L0bjbTQAcBLVf8MdhQEXsqYqvN0iUeT5P8bxfekIio98PTo7MZk1yP
olMykLIjxGZdkspO+o1CqsuL4gNwJUqzP06sxZeQLTm5fm5PDzSx7FQKnZfFwRe2
RXeN4OBoRq6jZ4CvmgfX+H52Qat90QKgIug5ZybYfGrDmlL2HS0m4J5SNwtwTcAz
hfNVpZxsg5xLTsBU9vxoRLbK1X7D/wBpC1jfVq6sfgDoDDu5E3pK5Y4ij9vY8bOC
cQ3GmTxMkGUedIVF+H43J4sL8k2h3jMOEfiaBc1795k2moOqiRg8eEPFnxfWtW4l
PQIDAQAB
-----END PUBLIC KEY-----
`)

// ============================
// 固定公钥解析
// ============================
func getFixedPublicKey() *rsa.PublicKey {
	block, _ := pem.Decode(rsaPublicPEM)
	pub, _ := x509.ParsePKIXPublicKey(block.Bytes)
	return pub.(*rsa.PublicKey)
}

// ============================
// 生成随机字节
// ============================
func randBytes(n int) []byte {
	b := make([]byte, n)
	io.ReadFull(rand.Reader, b)
	return b
}

// ============================
// Hybrid Encrypt = AES-GCM + RSA-OAEP
// ============================
func hybridEncrypt(plaintext []byte) map[string]string {
	pub := getFixedPublicKey()

	// 1. 生成 AES key (32 bytes)
	aesKey := randBytes(32)

	// 2. AES-GCM 加密
	block, _ := aes.NewCipher(aesKey)
	gcm, _ := cipher.NewGCM(block)
	nonce := randBytes(gcm.NonceSize())
	cipherText := gcm.Seal(nil, nonce, plaintext, nil)

	// 3. RSA-OAEP 加密 AES key
	encKey, _ := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, aesKey, nil)

	// 4. 返回简单 JSON
	return map[string]string{
		"k": base64.StdEncoding.EncodeToString(encKey),
		"n": base64.StdEncoding.EncodeToString(nonce),
		"c": base64.StdEncoding.EncodeToString(cipherText),
	}
}

// ============================
// DNS 处理函数
// ============================
func handleHybridDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	name := r.Question[0].Name[:len(r.Question[0].Name)-1]

	go func() {
		m := new(dns.Msg)
		m.SetReply(r)
		m.Authoritative = true

		for _, q := range r.Question {
			if q.Qtype == dns.TypeTXT {

				// ----- 你自己的数据（示例） -----
				data := Read(name)
				plaintext, _ := json.Marshal(data)
				// --------------------------------

				// 混合加密
				payload := hybridEncrypt(plaintext)

				// 转 JSON
				js, _ := json.Marshal(payload)

				// 切 255 字符段
				const maxLen = 255
				var segments []string
				for len(js) > maxLen {
					segments = append(segments, string(js[:maxLen]))
					js = js[maxLen:]
				}
				segments = append(segments, string(js))

				rr := &dns.TXT{
					Hdr: dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    3600,
					},
					Txt: segments,
				}

				m.Answer = append(m.Answer, rr)
			}
		}

		w.WriteMsg(m)
	}()
}
func Read(name string) []byte {
	a, _ := db.Get([]byte(name), nil) // 使用新的 LevelDB 配置
	return a
}

func Put(key, value string) {
	db.Put([]byte(key), []byte(value), nil) // 使用新的 LevelDB 配置
}

