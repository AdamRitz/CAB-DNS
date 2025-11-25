## Project Structure

- `Contract/`  
  Contains the Ethereum smart contracts and corresponding test scripts.  
  All contracts are written for **Solidity v0.8.20** and should be compiled and tested with a toolchain that supports this compiler version (e.g., Hardhat, Foundry, or Truffle configured with `0.8.20`).

- `Oracle/`  
  Implements the off-chain oracle service using **ethers.js**.  
  This component listens for on-chain events, calls the local ML model to generate domain-related content, and then sends the results back to the smart contracts via Ethereum transactions.

- `DNS/`  
  Implements a DNS server in **Go**, built on top of the `miekg/dns` library.  
  The DNS server parses incoming DNS queries and resolves them by interacting with the blockchain (via the Oracle) and the off-chain model when necessary.

## Model Dependency

To run the full pipeline (smart contracts + oracle + DNS resolver), you must first deploy the **WR-One2Set** model locally.  
Please follow the setup and deployment instructions in the WR-One2Set repository: https://github.com/DeepLearnXMU/WR-One2Set
