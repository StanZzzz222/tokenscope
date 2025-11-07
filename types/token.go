package types

/*
   Created by zyx
   Date Time: 2025/9/16
   File: token.go
*/

type TokenType uint8

const (
	ERC20 TokenType = iota
	ERC721
	None
)

func (t TokenType) String() string {
	switch t {
	case ERC20:
		return "ERC20"
	case ERC721:
		return "ERC721"
	case None:
		return "N/A"
	}
	return "N/A"
}
