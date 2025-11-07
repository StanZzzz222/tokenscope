# TokenScope

- [English](https://github.com/StanZzz11/tokenscope/blob/master/README.md)
**TokenScope** 是一个适用于兼容 EVM 区块链的简易资产查看工具，专注于获取指定地址的链上资产，包括 ERC-20 代币、NFT，并提供区块同步模块，通过 RPC 逐块拉取区块并同步交易数据。

---

## 功能特性

- 获取指定地址的 ETH 余额和交易记录
- 获取 ERC-20 代币余额及交易
- 获取 NFT（ERC-721）资产及交易
- 支持区块同步模块，通过 RPC 逐块拉取历史区块
- 支持增量同步，处理新生成区块

## 安装

### 前端
```bash
git clone https://github.com/StanZzz11/tokenscope.git
cd tokenscope/frontend
pnpm install autoprefixer
pnpm i
pnpm run dev
pnpm run build
```
### 后端
```bash
git clone https://github.com/StanZzz11/tokenscope.git
cd tokenscope
go mod tidy
go build -o tokenscope ./cmd
```

**主要依赖**
- Go 1.24
- [go-ethereum](https://github.com/ethereum/go-ethereum)
- [go-web3](https://github.com/chenzhijie/go-web3)
- [pebble](https://github.com/cockroachdb/pebble)
- [xsync](https://github.com/puzpuzpuz/xsync)
- [bloom](https://github.com/bits-and-blooms/bloom)
- [gin](https://github.com/gin-gonic/gin)

## 注意事项
- 区块同步为 RPC 逐块拉取，历史交易量大时可能较慢。
- 仓库中的ERC20与ERC721数据仅提供了主网的数据，如需使用其他网络则需自行获取数据并进行替换。
- TokenScope仅从公共数据中拉取了2316个ERC20与2934个ERC721数据，目前仅能对这部分ERC20与ERC721进行嗅探，如有需要可自行增加数据。

## 许可证
遵循 [LGPL-3.0](LICENSE(https://github.com/StanZzz11/tokenscope/blob/master/LICENCE)) 许可证。