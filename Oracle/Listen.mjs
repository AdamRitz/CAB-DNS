import { Web3 } from 'web3';
import * as fs from 'fs'
import * as readline from 'readline'
import * as pythons from './python.mjs'
import * as Sign from './signature.mjs'
import * as Enc from './Encryption.mjs'

//[          初始化区          ]--------------------------------------------------------------------------------------------------------------------------------------------------------
//[          钱包初始化          ]
const web3 = new Web3('YourNodeURL');
const wallet = web3.eth.accounts.wallet.add('YourPrivateKey')
const privateKey = 'YourPrivateKey';
//[          合约初始化          ]
const address = '0x78AbE8C2e6E81b2A2F06d388ca45F05c6cEd5cDD'
const abi = [{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"from","type":"address"},{"indexed":false,"internalType":"bytes32","name":"Tx","type":"bytes32"},{"indexed":true,"internalType":"string","name":"IndexedName","type":"string"},{"indexed":false,"internalType":"string","name":"Name","type":"string"}],"name":"AssignEvent","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"from","type":"address"},{"indexed":false,"internalType":"string","name":"content","type":"string"},{"indexed":false,"internalType":"string","name":"IP","type":"string"},{"indexed":false,"internalType":"string","name":"K","type":"string"},{"indexed":false,"internalType":"string","name":"psk","type":"string"}],"name":"RequestEvent","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"from","type":"address"},{"indexed":false,"internalType":"string","name":"sig","type":"string"}],"name":"SignEvent","type":"event"},{"inputs":[{"internalType":"string","name":"name","type":"string"},{"internalType":"bytes32","name":"Tx","type":"bytes32"}],"name":"Assign","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"string","name":"content","type":"string"},{"internalType":"string","name":"IP","type":"string"},{"internalType":"string","name":"K","type":"string"},{"internalType":"string","name":"psk","type":"string"}],"name":"Request","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"string","name":"sig","type":"string"}],"name":"Sign","outputs":[],"stateMutability":"nonpayable","type":"function"}]
const contract = new web3.eth.Contract(abi, address)


//-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

//[          函数区          ]---------------------------------------------------------------------------------------------------------------------------------------------------------

//[          事件监听函数-链下          ]
var c=0
async function ListenRequest() {//Assign(string memory name,bytes32 Tx)
    const subscription = contract.events.RequestEvent();
    subscription.on('connected', connected => console.log("订阅成功，开始监听Request事件（链下）", connected));
    subscription.on('data', async data => {
        const K=Enc.ECCDecrpt(data.returnValues.K)
        const psk=Enc.ECCDecrpt(data.returnValues.psk)
        //console.log(data)
        const text=data.returnValues.content
        const tx=data.transactionHash
        const name =await pythons.GetName(text)
        console.log(name)
        Assign(name,tx)
    });
}



//----------------------------------------------------------------[         处理函数        ]------------------------------------------------------------

async function Assign(name,tx) {
    const accounts = web3.eth.accounts.wallet[0].address;
    const myContract = contract;

    try {
        c=c+1
        console.log(c)
        const baseGasPrice = await web3.eth.getGasPrice();
        const receipt = await myContract.methods.Assign(name,tx).send({
            from: accounts,
            gas: 1000000,
            gasPrice: baseGasPrice.toString(), 
        }).then(console.log);
        return receipt;
    } catch (error) {
        console.error(error);
    }
}
//---------------------------------------------------------------------------------------------------------------------------------------------------------------------------
function Writetxt(msg){
    const stream=fs.appendFile('Data/gas.txt', msg+"\n", 'utf8', (err) => {
        if (err) {
            console.error('Error appending to file:', err);
        } else {
            console.log(msg)
        }
    });

}

async function GasExperiment(n) {
    const stream = fs.createReadStream('Data/test.txt');
    const rl = readline.createInterface({
        input: stream,
        crlfDelay: Infinity
    });

    const promises = [];
    const accounts = web3.eth.accounts.wallet[0].address;
    
    let nonce = await web3.eth.getTransactionCount(accounts, 'pending'); // 获取当前 nonce
    let count = 0;
    for await (const line of rl) {
        if (count >= n) break; 

        promises.push(
            Request2(line, "66.6.66.1") // 传入不同的 nonce
                .then(receipt => Writetxt(receipt.gasUsed))
                .catch(error => console.error(`Request failed: ${error}`))
        );

        count++;
    }

    await Promise.allSettled(promises);
    console.log("All requests completed.");
}





//------------------------------------------------------------------------------{      调用区     }--------------------------------------------------------------------------------------------

ListenRequest()