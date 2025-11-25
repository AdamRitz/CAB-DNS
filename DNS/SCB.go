package main

// CAB-DNS Req、Assign 函数实现
import (
	"encoding/json"
	"fmt"
	"github.com/chain-lab/go-norn/utils"
	"math/big"
	"os/exec"
	"regexp"
	"strings"
)

func Request(context, ip string) ([]byte, string) {
	privateKeyHex := "af44a005ad9e6d4c0873d609de9df16a9f8c8e490597087f4286b8103e0e7149"
	privateKeyBig, _ := new(big.Int).SetString(privateKeyHex, 16)
	mw, proxysk, Kx, Ky, _ := Delegate(privateKeyBig, ip)
	sig, _ := ProxySign(mw, proxysk, Kx, Ky, "test.cab")
	a, _ := json.Marshal(RequestMessage{Context: context, Mw: mw, ProxySK: proxysk, Kx: Kx, Ky: Ky, Sig: sig})
	b, _ := ECCEncrypt(string(a))
	return a, b

}
func Assign(requestmessage string, db *utils.LevelDB) {
	s, _ := ECCDecrypt(requestmessage)
	var msg RequestMessage
	json.Unmarshal([]byte(s), &msg)
	DomainName := DomainNameGen(msg.Context) + ".scb"
	db.Insert([]byte(DomainName), []byte(s))

}
func DomainNameGen(content string) string {
	re := regexp.MustCompile(`[[:punct:]]`) // 匹配所有标点符号
	content = re.ReplaceAllString(content, "")
	//fmt.Println(content)
	cmd := exec.Command("C:/Users/DonQuixote/.conda/envs/py308/python.exe", "E:/Code/MLCode/kg_one2set-master/predict.py", content)
	output, _ := cmd.CombinedOutput()
	//fmt.Println(string(output))
	keyphrases := strings.ReplaceAll(string(output), " ", "")
	a := strings.Split(keyphrases, ";")
	fmt.Println(a)
	return a[0]
}
