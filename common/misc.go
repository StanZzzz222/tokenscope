package common

import (
	"math/big"
	"strings"
)

/*
   Created by zyx
   Date Time: 2025/9/26
   File: misc.go
*/

// HexToBigInt Hex string to big int
func HexToBigInt(hex string) *big.Int {
	n := new(big.Int)
	if strings.HasPrefix(hex, "0x") || strings.HasPrefix(hex, "0X") {
		hex = hex[2:]
	}
	if _, ok := n.SetString(hex, 16); ok {
		return n
	}
	return nil
}
