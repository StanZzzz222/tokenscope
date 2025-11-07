# TokenScope

- [中文文档](https://github.com/StanZzzz222/tokenscope/blob/master/README_zh.md)

  **TokenScope** is a lightweight asset viewer tool for EVM-compatible chains, designed to retrieve on-chain assets for a specified address, including ERC-20 tokens and NFTs. It also provides a block synchronization module that pulls blocks via RPC and synchronizes transaction data incrementally.

---

## Features

- Retrieve ETH balance and transaction history for a given address
- Retrieve ERC-20 token balances and transactions
- Retrieve NFT (ERC-721) assets and transactions
- Block synchronization module that pulls historical blocks via RPC
- Incremental synchronization for newly generated blocks

## Installation

### Frontend
```bash
git clone https://github.com/StanZzzz222/tokenscope.git
cd tokenscope/frontend
pnpm install autoprefixer
pnpm i
pnpm run dev
pnpm run build
```

### Backend
```bash
git clone https://github.com/StanZzzz222/tokenscope.git
cd tokenscope
go mod tidy
go build -o tokenscope ./cmd
```

**Main Dependencies**
- Go 1.24
- [go-ethereum](https://github.com/ethereum/go-ethereum)
- [go-web3](https://github.com/chenzhijie/go-web3)
- [pebble](https://github.com/cockroachdb/pebble)
- [xsync](https://github.com/puzpuzpuz/xsync)
- [bloom](https://github.com/bits-and-blooms/bloom)
- [gin](https://github.com/gin-gonic/gin)

## Notes
- Block synchronization is performed block by block, which may be slow when historical transaction volume is high.
- The ERC20 and ERC721 data in the repository only provides data from the mainnet. If you need to use other networks, you must obtain and replace the data yourself.
- TokenScope only pulls data for 2,316 ERC20 and 2,934 ERC721 from the public database. Only this portion of ERC20 and ERC721 data is being sniffed. You can add more data as needed.

## License
Licensed under [LGPL-3.0](LICENSE(https://github.com/StanZzzz222/tokenscope/blob/master/LICENCE)).