// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract DNS {
    event RequestEvent(address from,string content,string IP,string  K,string  psk);
    event AssignEvent(address from,bytes32 Tx,string indexed  IndexedName,string Name);
    event SignEvent(address from,string sig);
    modifier CheckOracle(){//验证是否为预言机
        require(msg.sender==0x2D26974DB2b3FD5eac63c18CE33c35A5Db03f912,"Only Oracle can access this method:>");
        _;
    }
    // 存储
    function Request(string memory content, string memory IP,string memory K,string memory psk) public payable{
        emit RequestEvent(msg.sender, content,IP,K,psk);
    }
    function Assign(string memory name,bytes32 Tx) public  payable{
        emit AssignEvent(msg.sender,Tx, name,name);
    }
    function Sign(string memory sig)public {
        emit SignEvent(msg.sender,sig);
    }

}