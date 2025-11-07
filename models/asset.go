package models

import (
	"fmt"
	"github.com/spf13/viper"
	"math/big"
	"strings"
	"tokenscope/types"
)

/*
   Created by zyx
   Date Time: 2025/9/17
   File: asset.go
*/

type Asset struct {
	Address      string   `json:"address"`
	Balance      string   `json:"balance"`
	ERC20Tokens  []*Token `json:"erc20_tokens"`
	ERC721Tokens []*Token `json:"erc721_tokens"`
	Txs          []*Tx    `json:"txs"`
	TxCount      uint     `json:"tx_count"`
}

type TokenAsset struct {
	Token     *Token          `json:"token"`
	Value     string          `json:"value"`
	TokenType types.TokenType `json:"token_type"`
}

type NFTAsset struct {
	Token       *Token          `json:"token"`
	TokenId     string          `json:"token_id"`
	MetadataUrl string          `json:"metadata_url"`
	SpecialUrl  bool            `json:"special_url"`
	TokenType   types.TokenType `json:"token_type"`
}

func NewAsset(address, balance string, erc20Tokens, erc721Tokens []*Token, txs []*Tx) *Asset {
	return &Asset{
		Address:      address,
		Balance:      balance,
		ERC20Tokens:  erc20Tokens,
		ERC721Tokens: erc721Tokens,
		Txs:          txs,
		TxCount:      uint(len(txs)),
	}
}

func NewNFTAsset(token *Token, tokenId, metadataUrl string, tokenType types.TokenType) *NFTAsset {
	ret := &NFTAsset{
		Token:       token,
		TokenId:     tokenId,
		MetadataUrl: metadataUrl,
		TokenType:   tokenType,
	}
	// Some NFTs do not have a TokenURI method and require special handling
	switch strings.ToLower(token.Address) {
	// ENS
	case strings.ToLower(types.ENS):
		n := new(big.Int)
		n.SetString(tokenId, 10)
		network := "mainnet"
		chainId := viper.GetInt("rpc.chain-id")
		switch chainId {
		case 1:
			network = "mainnet"
		case 3:
			network = "ropsten"
		case 5:
			network = "goerli"
		case 42161:
			network = "arbitrum"
		case 421613:
			network = "arbitrum-goerli"
		case 10:
			network = "optimism"
		case 420:
			network = "optimism-goerli"
		default:
			network = "mainnet"
		}
		ret.MetadataUrl = fmt.Sprintf("https://metadata.ens.domains/%v/0x57f1887a8bf19b14fc0df6fd9b2acc9af147ea85/%v/image", network, fmt.Sprintf("0x%064x", n))
		ret.SpecialUrl = true
		break
	// Crypto Kitties
	case strings.ToLower(types.CryptoKitties):
		ret.MetadataUrl = fmt.Sprintf("https://img.cn.cryptokitties.co/0x06012c8cf97bead5deae237070f9587f8e7a266d/%v.png", tokenId)
		ret.SpecialUrl = true
		break
	}
	return ret
}

func NewTokenAsset(token *Token, value string, tokenType types.TokenType) *TokenAsset {
	return &TokenAsset{
		Token:     token,
		Value:     value,
		TokenType: tokenType,
	}
}

func (n *NFTAsset) TokenIDToBigInt() (*big.Int, error) {
	bi, ok := new(big.Int).SetString(n.TokenId, 10)
	if !ok {
		return nil, fmt.Errorf("invalid tokenId: %s", n.TokenId)
	}
	return bi, nil
}
