package assets

import (
	"encoding/hex"
	"errors"
	"fmt"
	w3Types "github.com/chenzhijie/go-web3/types"
	eCommon "github.com/ethereum/go-ethereum/common"
	"math/big"
	"strings"
	"tokenscope/common/client"
	"tokenscope/common/logger"
	"tokenscope/types"
)

/*
   Created by zyx
   Date Time: 2025/9/26
   File: contract.go
*/

// getERC20Balance Get ERC20 token balance
func getERC20Balance(address string, tokenAddress string) (string, error) {
	var out []any
	var result []byte
	var decodeData []byte
	abi := GetABI(types.ERC20)
	args := []any{eCommon.HexToAddress(address)}
	token, tokenType := getTokenInfo(tokenAddress)
	if tokenType == types.None {
		return "0", errors.New("token is nil")
	}
	data, err := abi.Pack("balanceOf", args...)
	if err != nil {
		logger.Logger().Errorf("GetERC20TokenAssets abi.Pack falied: %v", err)
		return "0", err
	}
	result, err = callContract(token.Address, data)
	if err != nil {
		logger.Logger().Errorf("GetERC20TokenAssets CallContract falied: %v", err)
		return "0", err
	}
	if len(result) == 0 || string(result) == "0x" {
		return "0", errors.New("result is empty")
	}
	decodeData, err = hex.DecodeString(strings.TrimPrefix(string(result), "0x"))
	if err != nil {
		logger.Logger().Errorf("GetERC20TokenAssets decode falied: %v", err)
		return "0", err
	}
	out, err = abi.Methods["balanceOf"].Outputs.UnpackValues(decodeData)
	if err != nil {
		logger.Logger().Errorf("GetERC20TokenAssets Get balanceOf function falied: %v", err)
		return "0", err
	}
	if len(out) == 1 {
		switch v := out[0].(type) {
		case []byte:
			return string(v), nil
		default:
			return fmt.Sprintf("%v", v), nil
		}
	}
	return "0", nil
}

// getERC721TokenAsset Get ERC721 token asset
func getERC721TokenAsset(tokenAddress string, tokenId *big.Int) (string, error) {
	var out []any
	var result []byte
	var decodeData []byte
	abi := GetABI(types.ERC721)
	args := []any{tokenId}
	token, tokenType := getTokenInfo(strings.ToLower(tokenAddress))
	if tokenType == types.None {
		return "", errors.New("token is nil")
	}
	data, err := abi.Pack("ownerOf", args...)
	if err != nil {
		logger.Logger().Errorf("getERC721TokenAsset abi.Pack falied: %v, tokenId: %v", err, tokenId)
		return "", err
	}
	result, err = callContract(token.Address, data)
	if err != nil {
		return "", err
	}
	if len(result) == 0 || string(result) == "0x" {
		return "", errors.New("result is empty")
	}
	decodeData, err = hex.DecodeString(strings.TrimPrefix(string(result), "0x"))
	if err != nil {
		logger.Logger().Errorf("getERC721TokenAsset decode falied: %v", err)
		return "", err
	}
	out, err = abi.Methods["ownerOf"].Outputs.UnpackValues(decodeData)
	if err != nil {
		logger.Logger().Errorf("getERC721TokenAsset Get ownerOf function falied: %v", err)
		return "", err
	}
	if len(out) == 1 {
		switch v := out[0].(type) {
		case []byte:
			return string(v), nil
		default:
			return fmt.Sprintf("%v", v), nil
		}
	}
	return "", nil
}

// getERC721TokenTotalSupply Get ERC721 token total supply safely
func getERC721TokenTotalSupply(tokenAddress string) *big.Int {
	abi := GetABI(types.ERC721)
	data, err := abi.Pack("totalSupply")
	if err != nil {
		logger.Logger().Errorf("getERC721TokenTotalSupply abi.Pack failed: %v", err)
		return nil
	}
	result, err := callContract(tokenAddress, data)
	if err != nil {
		return nil
	}
	if len(result) == 0 || string(result) == "0x" {
		return nil
	}
	decodeData, err := hex.DecodeString(strings.TrimPrefix(string(result), "0x"))
	if err != nil {
		logger.Logger().Errorf("getERC721TokenTotalSupply decode failed: %v", err)
		return nil
	}
	out, err := abi.Methods["totalSupply"].Outputs.UnpackValues(decodeData)
	if err != nil {
		logger.Logger().Errorf("getERC721TokenTotalSupply unpack failed: %v", err)
		return nil
	}
	if len(out) != 1 {
		return nil
	}
	switch v := out[0].(type) {
	case *big.Int:
		return v
	case uint64:
		return new(big.Int).SetUint64(v)
	default:
		return nil
	}
}

// getERC721BaseTokenURI Get ERC721 base token token URI
func getERC721BaseTokenURI(method, tokenAddress string, args ...any) string {
	abi := GetABI(types.ERC721)
	data, err := abi.Pack(method, args...)
	if err != nil {
		return ""
	}
	result, err := callContract(tokenAddress, data)
	if err != nil {
		return ""
	}
	if len(result) == 0 || string(result) == "0x" {
		return ""
	}
	decodeData, err := hex.DecodeString(strings.TrimPrefix(string(result), "0x"))
	if err != nil {
		return ""
	}
	out, err := abi.Methods[method].Outputs.UnpackValues(decodeData)
	if err != nil {
		return ""
	}
	if len(out) == 1 {
		switch v := out[0].(type) {
		case []byte:
			return string(v)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

// callContract Call contract
func callContract(to string, data []byte) ([]byte, error) {
	evmClient := client.ContractEVMClient()
	cli := evmClient.GetClient()
	res, err := cli.Eth.Call(&w3Types.CallMsg{
		From: eCommon.Address{},
		To:   eCommon.HexToAddress(to),
		Data: data,
	}, nil)
	return []byte(res), err
}
