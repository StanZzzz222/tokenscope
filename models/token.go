package models

import "tokenscope/types"

/*
   Created by zyx
   Date Time: 2025/9/16
   File: erc20.go
*/

type Token struct {
	Name        string          `json:"name"`        // Token name
	Description string          `json:"description"` // Token description
	Address     string          `json:"address"`     // Token address
	Symbol      string          `json:"symbol"`      // Token symbol
	Link        string          `json:"link"`        // Token link
	Decimals    uint8           `json:"decimals"`    // Token decimal, ERC721 is 0
	Type        string          `json:"type"`        // Token type string: ERC20 / ERC721
	TokenType   types.TokenType `json:"-"`           // Token type: ERC20 / ERC721
}

func (t *Token) IsERC20() bool {
	return t.TokenType == types.ERC20
}
func (t *Token) IsERC721() bool {
	return t.TokenType == types.ERC721
}
