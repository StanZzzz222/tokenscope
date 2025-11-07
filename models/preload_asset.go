package models

import (
	"strings"
	"tokenscope/types"
)

/*
   Created by zyx
   Date Time: 2025/9/16
   File: preload_asset.go
*/

type PreloadAsset struct {
	ERC20Tokens  map[string]*Token          `json:"erc_20_tokens"`
	ERC721Tokens map[string]*Token          `json:"erc_721_tokens"`
	Signatures   map[string]*Signatures     `json:"signatures"`
	Abis         map[types.TokenType][]byte `json:"abis"`
}

func NewPreloadAsset(erc20Tokens, erc721Tokens map[string]*Token, signatures map[string]*Signatures, abis map[types.TokenType][]byte) *PreloadAsset {
	preloadAsset := &PreloadAsset{
		ERC20Tokens:  erc20Tokens,
		ERC721Tokens: erc721Tokens,
		Signatures:   signatures,
		Abis:         abis,
	}
	return preloadAsset
}

func (p *PreloadAsset) GetSignaturesHash(methodHash string) *Signatures {
	var ret *Signatures
	if methodHash[:2] == "0x" {
		methodHash = methodHash[2:]
	}
	ret = p.Signatures[strings.ToLower(methodHash)]
	return ret
}

func (p *PreloadAsset) GetERC20Token(address string) *Token {
	if ret, ok := p.ERC20Tokens[strings.ToLower(address)]; ok {
		return ret
	}
	return nil
}

func (p *PreloadAsset) GetERC721Token(address string) *Token {
	if ret, ok := p.ERC721Tokens[strings.ToLower(address)]; ok {
		return ret
	}
	return nil
}
