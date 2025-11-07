package models

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

/*
   Created by zyx
   Date Time: 2025/9/19
   File: abi_data.go
*/

var (
	hexLookup = [...]byte{
		'0': 0, '1': 1, '2': 2, '3': 3, '4': 4,
		'5': 5, '6': 6, '7': 7, '8': 8, '9': 9,
		'a': 10, 'b': 11, 'c': 12, 'd': 13, 'e': 14, 'f': 15,
		'A': 10, 'B': 11, 'C': 12, 'D': 13, 'E': 14, 'F': 15,
	}
)

type ABIData struct {
	MethodHash    string     `json:"method_hash"`
	AddressValues []string   `json:"address_values"`
	IntValues     []*big.Int `json:"int_values"`
}

func NewABIData(methodHash string, addresses []string, intValues []*big.Int) *ABIData {
	return &ABIData{
		MethodHash:    methodHash,
		AddressValues: addresses,
		IntValues:     intValues,
	}
}

// DecodeABIData Decode abi data
func DecodeABIData(input []byte) *ABIData {
	if len(input) < 10 {
		return nil
	}
	start := 0
	if len(input) >= 2 && input[0] == '0' && input[1] == 'x' {
		start = 2
	}
	clean := input[start:]
	if len(clean) < 8 {
		return nil
	}
	methodHash := string(clean[:8])
	args := clean[8:]
	chunkCount := len(args) / 64
	if chunkCount == 0 {
		return NewABIData(methodHash, nil, nil)
	}
	addresses := make([]string, chunkCount)
	intValues := make([]*big.Int, chunkCount)
	var bytesChunk [32]byte
	for i := 0; i < chunkCount; i++ {
		chunk := args[i*64 : (i+1)*64]
		for j := 0; j < 32; j++ {
			b1 := hexLookup[chunk[j*2]]
			b2 := hexLookup[chunk[j*2+1]]
			bytesChunk[j] = (b1 << 4) | b2
		}
		n := big.NewInt(0).SetBytes(bytesChunk[:])
		intValues[i] = n
		addresses[i] = common.BytesToAddress(bytesChunk[12:]).Hex()
	}
	return NewABIData(methodHash, addresses, intValues)
}
