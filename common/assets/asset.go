package assets

import (
	"errors"
	eCommon "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
	"math/big"
	"slices"
	"strings"
	"tokenscope/common"
	"tokenscope/common/blockchain"
	"tokenscope/common/client"
	"tokenscope/models"
	"tokenscope/repo"
	"tokenscope/types"
)

/*
   Created by zyx
   Date Time: 2025/9/17
   File: asset.go
*/

func GetAssetByAddress(address string) *models.Asset {
	if !checkAddress(address) {
		return nil
	}
	cli := client.ContractEVMClient().GetClient()
	addr := eCommon.HexToAddress(address)
	address = addr.Hex()
	ch := make(chan string, 1)
	wg := common.Wg(1)
	blsRepo := repo.BlockchainRepository()
	txLogRepo := repo.TxLogRepository()
	wg.Go(func() {
		balance, _ := cli.Eth.GetBalance(addr, nil)
		ch <- balance.String()
	})
	if !blsRepo.HasAddressIndexed(address) {
		blsRepo.StoreAddressIndexed(address)
		blockchain.AddressIndexed(address, blsRepo, txLogRepo)
	}
	txs := getTxsByAddress(address, blsRepo.StreamBlocks())
	txLogs := getTxLogsByAddress(address, txLogRepo.StreamTxLogs())
	erc20Tokens, erc721Tokens := getTokens(txs, txLogs, address)
	wg.Wait()
	return models.NewAsset(address, <-ch, erc20Tokens, erc721Tokens, txs)
}

func GetERC20TokenAssets(address string, tokenAddress string) (*models.TokenAsset, error) {
	token, tokenType := getTokenInfo(tokenAddress)
	if tokenType == types.None {
		return nil, errors.New("token is nil")
	}
	balance, err := getERC20Balance(address, tokenAddress)
	if err != nil {
		return nil, err
	}
	ret := models.NewTokenAsset(token, balance, tokenType)
	return ret, nil
}

func GetERC721TokenAssets(address string, tokenAddress string) ([]*models.NFTAsset, error) {
	blsRepo := repo.BlockchainRepository()
	txLogRepo := repo.TxLogRepository()
	txs := getTxsByAddress(address, blsRepo.StreamBlocks())
	txLogs := getTxLogsByAddress(tokenAddress, txLogRepo.StreamTxLogs())
	token, tokenType := getTokenInfo(tokenAddress)
	if token != nil {
		ret := make([]*models.NFTAsset, 0)
		switch strings.ToLower(token.Address) {
		case strings.ToLower(types.ENS):
			ret = searchENSAssets(address, tokenAddress, txLogs, token, tokenType)
		default:
			ret = searchERC721Assets(address, tokenAddress, txs, txLogs, token, tokenType)
		}
		return ret, nil
	}
	return nil, errors.New("nft not found")
}

func GetERC721TokenURI(tokenAddress string) string {
	baseTokenURI := getERC721BaseTokenURI("baseTokenURI", tokenAddress)
	if len(baseTokenURI) <= 0 {
		baseTokenURI = getERC721BaseTokenURI("tokenURI", tokenAddress, big.NewInt(1))
		if len(baseTokenURI) <= 0 {
			baseTokenURI = "N/A"
		}
	}
	// IPFS
	if strings.HasPrefix(baseTokenURI, "ipfs://") {
		erc721Url := viper.GetString("image.erc721-url")
		baseTokenURI = strings.ReplaceAll(baseTokenURI, "ipfs://", "")
		baseTokenURI = strings.ReplaceAll(baseTokenURI, "/", "")
		baseTokenURI = strings.ReplaceAll(erc721Url, "#{cid}", baseTokenURI)
	}
	baseTokenURI = strings.ReplaceAll(baseTokenURI, "1", "")
	return baseTokenURI
}

// searchERC721Assets Search ERC721 Assets
func searchERC721Assets(address, tokenAddress string, txs []*models.Tx, txLogs []*models.TxLog, token *models.Token, tokenType types.TokenType) []*models.NFTAsset {
	ret := make([]*models.NFTAsset, 0)
	assets := make(map[string]*models.NFTAsset)
	tokenURI := GetERC721TokenURI(tokenAddress)
	for _, tx := range txs {
		if tx != nil {
			if (address == tx.From || address == tx.To) || tx.Data != nil {
				abiData := models.DecodeABIData(tx.Data)
				if abiData != nil {
					if slices.Contains(abiData.AddressValues, address) {
						for _, v := range abiData.IntValues {
							tokenId := v.String()
							if _, ok := assets[tokenId]; !ok {
								assets[tokenId] = models.NewNFTAsset(token, tokenId, tokenURI, tokenType)
							}
						}
					}
					if tx.To == tokenAddress {
						for _, v := range abiData.IntValues {
							tokenId := v.String()
							if _, ok := assets[tokenId]; !ok {
								assets[tokenId] = models.NewNFTAsset(token, tokenId, tokenURI, tokenType)
							}
						}
					}
				}
			}
		}
	}
	// TxLog search
	for _, txLog := range txLogs {
		if txLog != nil {
			for _, txEvent := range txLog.TxEvents {
				signatures := GetSignatures(txEvent.MethodHash)
				if signatures != nil {
					source := strings.ToLower(signatures.Source)
					if strings.Contains(source, "transfer") && txEvent.ContractAddress == tokenAddress {
						if len(txEvent.Values) > 0 {
							from, to, tokenIdStr := txEvent.Values[0], txEvent.Values[1], txEvent.Values[2]
							if v := common.HexToBigInt(tokenIdStr); v != nil {
								tokenId := v.String()
								if _, ok := assets[tokenId]; !ok {
									if from == address || to == address {
										assets[tokenId] = models.NewNFTAsset(token, tokenId, tokenURI, tokenType)
									}
								}
							}
						}
					}
				}
			}
		}
	}
	wg := common.Wg(8)
	for _, nft := range assets {
		tokenId, err := nft.TokenIDToBigInt()
		if err != nil {
			continue
		}
		if tokenId.Cmp(big.NewInt(0)) > 0 {
			wg.Go(func() {
				ownerAddress, err := getERC721TokenAsset(tokenAddress, tokenId)
				if err == nil && ownerAddress == address {
					ret = append(ret, nft)
				}
			})
		}
	}
	wg.Wait()
	return ret
}

// searchENSAssets Search ENS Assets
func searchENSAssets(address, tokenAddress string, txLogs []*models.TxLog, token *models.Token, tokenType types.TokenType) []*models.NFTAsset {
	ret := make([]*models.NFTAsset, 0)
	assets := make(map[string]*models.NFTAsset)
	for _, txLog := range txLogs {
		for _, txEvent := range txLog.TxEvents {
			signatures := GetSignatures(txEvent.MethodHash)
			if signatures != nil {
				source := strings.ToLower(signatures.Source)
				if strings.Contains(source, "transfer") && txEvent.ContractAddress == tokenAddress {
					from, to, tokenIdStr := eCommon.HexToAddress(txEvent.Values[0]).Hex(), eCommon.HexToAddress(txEvent.Values[1]).Hex(), txEvent.Values[2]
					if v := common.HexToBigInt(tokenIdStr); v != nil {
						tokenId := v.String()
						if _, ok := assets[tokenId]; !ok {
							if from == address || to == address {
								assets[tokenId] = models.NewNFTAsset(token, tokenId, "", tokenType)
							}
						}
					}
				}
			}
		}
	}
	for _, nft := range assets {
		tokenId, err := nft.TokenIDToBigInt()
		if err != nil {
			continue
		}
		if tokenId.Cmp(big.NewInt(0)) > 0 {
			ownerAddress, err := getERC721TokenAsset(tokenAddress, tokenId)
			if err == nil && ownerAddress == address {
				ret = append(ret, nft)
			}
		}
	}
	return ret
}

// getTxsByAddress Get Txs by address
func getTxsByAddress(address string, ch <-chan *models.Block) []*models.Tx {
	txs := make([]*models.Tx, 0)
	retCh := make(chan *models.Tx, 4096)
	wg := common.Wg(32)
	count := 0
	for block := range ch {
		count += 1
		wg.Go(func() {
			if block != nil && block.HasBloom(address) {
				for _, tx := range block.Txs {
					if tx.From == address || tx.To == address {
						retCh <- tx
						continue
					}
					abiData := models.DecodeABIData(tx.Data)
					if abiData != nil && slices.Contains(abiData.AddressValues, address) {
						retCh <- tx
						continue
					}
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(retCh)
	}()
	for tx := range retCh {
		txs = append(txs, tx)
	}
	slices.SortFunc(txs, func(a, b *models.Tx) int {
		if a.Timestamp > b.Timestamp {
			return -1
		}
		if a.Timestamp < b.Timestamp {
			return 1
		}
		return 0
	})
	return txs
}

// getTxLogsByAddress Get Tx log by address
func getTxLogsByAddress(address string, ch <-chan *models.TxLog) []*models.TxLog {
	ret := make([]*models.TxLog, 0)
	retCh := make(chan *models.TxLog, 1024)
	wg := common.Wg(32)
	for txLog := range ch {
		wg.Go(func() {
			if txLog != nil && txLog.HasBloom(address) {
				for _, txEvent := range txLog.TxEvents {
					if slices.Contains(txEvent.Values, address) {
						retCh <- txLog
						continue
					}
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(retCh)
	}()
	for tx := range retCh {
		ret = append(ret, tx)
	}
	return ret
}

// getTokens Get Token Assets
func getTokens(txs []*models.Tx, txLogs []*models.TxLog, address string) ([]*models.Token, []*models.Token) {
	erc20Map := make(map[string]*models.Token)
	erc721Map := make(map[string]*models.Token)
	erc20Tokens := make([]*models.Token, 0, len(erc20Map))
	erc721Tokens := make([]*models.Token, 0, len(erc721Map))
	addToken := func(token *models.Token, tokenType types.TokenType) {
		if token == nil {
			return
		}
		switch tokenType {
		case types.ERC20:
			if _, exists := erc20Map[token.Address]; !exists {
				erc20Map[token.Address] = token
			}
		case types.ERC721:
			if _, exists := erc721Map[token.Address]; !exists {
				erc721Map[token.Address] = token
			}
		default:
		}
	}
	// Sniff the ERC20 / ERC721 address that the address may have interacted with
	for _, tx := range txs {
		if tx != nil {
			if (address == tx.From || address == tx.To) || tx.Data != nil {
				abiData := models.DecodeABIData(tx.Data)
				if abiData != nil && slices.Contains(abiData.AddressValues, address) {
					for _, value := range abiData.AddressValues {
						if value != address {
							token, tokenType := getTokenInfo(value)
							addToken(token, tokenType)
						}
					}
				}
				token, tokenType := getTokenInfo(tx.To)
				addToken(token, tokenType)
			}
		}
	}
	// TxLog search
	for _, txLog := range txLogs {
		if txLog != nil {
			for _, txEvent := range txLog.TxEvents {
				signatures := GetSignatures(txEvent.MethodHash)
				if signatures != nil {
					source := strings.ToLower(signatures.Source)
					if strings.Contains(source, "transfer") && slices.Contains(txEvent.Values, address) {
						token, tokenType := getTokenInfo(txEvent.ContractAddress)
						addToken(token, tokenType)
						for _, v := range txEvent.Values {
							if v != address {
								token, tokenType = getTokenInfo(v)
								addToken(token, tokenType)
							}
						}
					}
				}
			}
		}
	}
	for _, t := range erc20Map {
		erc20Tokens = append(erc20Tokens, t)
	}
	for _, t := range erc721Map {
		erc721Tokens = append(erc721Tokens, t)
	}
	return erc20Tokens, erc721Tokens
}

// getTokenInfo Get token info
func getTokenInfo(address string) (*models.Token, types.TokenType) {
	token := GetERC20Token(address)
	if token != nil {
		return token, types.ERC20
	}
	token = GetERC721Token(address)
	if token != nil {
		return token, types.ERC721
	}
	return token, types.None
}

// checkAddress EVM address check (ignore case, no checksum validation)
func checkAddress(address string) bool {
	if len(address) != 42 || !strings.HasPrefix(address, "0x") {
		return false
	}
	return eCommon.IsHexAddress(address)
}
