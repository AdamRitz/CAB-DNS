/**
 *Submitted for verification at Etherscan.io on 2025-03-18
*/

pragma solidity ^0.8.0;

contract ENS {
    function makeCommitment(
        string memory name,
        address owner,
        bytes32 secret
    ) public pure returns (bytes32) {}

    function commit(bytes32 commitment) public {}

    function registerWithConfig(
        string memory name,
        address owner,
        uint256 duration,
        bytes32 secret,
        address resolver,
        address addr
    ) public payable {}

    function rentPrice(string memory name, uint256 duration)
        public
        view
        returns (uint256)
    {}

    function register(
        string calldata name,
        address owner,
        uint256 duration,
        bytes32 secret
    ) external payable {}
}

contract ENSExperiment {
    function commit(string memory name) public {
        ENS(0x7e02892cfc2Bfd53a75275451d73cF620e793fc0).commit(
            ENS(0x7e02892cfc2Bfd53a75275451d73cF620e793fc0).makeCommitment(
                name,
                0x2D26974DB2b3FD5eac63c18CE33c35A5Db03f912,
                0xaa50b881719ee948a45fe3eb545f3ec6388b6c379be52ad1d3e743dda7c0f1cc
            )
        );
    }

    function price(string memory name) public view returns (uint256) {
        return
            ENS(0x7e02892cfc2Bfd53a75275451d73cF620e793fc0).rentPrice(
                name,
                31536000
            );
    }

    function register(string calldata name) public payable {
        uint256 p = ENS(0x7e02892cfc2Bfd53a75275451d73cF620e793fc0).rentPrice(
            name,
            31536000
        );
        ENS(0x7e02892cfc2Bfd53a75275451d73cF620e793fc0).register{value: p}(
            name,
            0x2D26974DB2b3FD5eac63c18CE33c35A5Db03f912,
            31536000,
            0xaa50b881719ee948a45fe3eb545f3ec6388b6c379be52ad1d3e743dda7c0f1cc
        );
    }

    function EXPCommit(string[] calldata a) public {
        for (uint256 i = 0; i < a.length; i++) {
            commit(a[i]);
        }
    }

    function EXPRegister(string[] calldata a) public payable {
        for (uint256 i = 0; i < a.length; i++) {
            register(a[i]);
        }
    }
    function EXPPrice(string[] calldata a) public view returns (uint) {
        uint c=0;
        for (uint256 i=0;i< a.length;i++){
            c=c+price(a[i]);
        }
        return c;
    }
        

        function withdraw() external {
        require(msg.sender == 0x2D26974DB2b3FD5eac63c18CE33c35A5Db03f912, "Only owner can withdraw");
        require(address(this).balance > 0, "No funds available");

        payable(0x2D26974DB2b3FD5eac63c18CE33c35A5Db03f912).transfer(address(this).balance);
    }
    function pay() external payable {}
}