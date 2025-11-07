package assets

import (
	"bytes"
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/tidwall/gjson"
	"os"
	"path/filepath"
	"strings"
	"tokenscope/common/logger"
	"tokenscope/models"
	"tokenscope/types"
)

/*
   Created by zyx
   Date Time: 2025/9/16
   File: preload.go
*/

var preloadAsset = models.NewPreloadAsset(make(map[string]*models.Token), make(map[string]*models.Token), make(map[string]*models.Signatures), make(map[types.TokenType][]byte))

// PreloadAssets Asset data preload
func PreloadAssets() {
	preloadABIs()
	preloadTokens()
	preloadSignatures()
}

// GetABI Get abi
func GetABI(tokenType types.TokenType) *abi.ABI {
	if abiBytes, ok := preloadAsset.Abis[tokenType]; ok {
		parsedABI, err := abi.JSON(bytes.NewReader(abiBytes))
		if err != nil {
			logger.Logger().Errorf("GetABI abi.JSON err: %v", err)
			return nil
		}
		return &parsedABI
	}
	return nil
}

// GetERC20Token Get ERC20 token
func GetERC20Token(address string) *models.Token {
	token := preloadAsset.GetERC20Token(address)
	if token != nil {
		return token
	}
	return nil
}

// GetERC721Token Get ERC721 token
func GetERC721Token(address string) *models.Token {
	token := preloadAsset.GetERC721Token(address)
	if token != nil {
		return token
	}
	return nil
}

// GetSignatures Get signatures
func GetSignatures(methodHash string) *models.Signatures {
	signatures := preloadAsset.GetSignaturesHash(methodHash)
	if signatures != nil {
		return signatures
	}
	return nil
}

// preloadABIs Preload ABIs
func preloadABIs() {
	err := filepath.Walk("./data/abis", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".abi") {
			var ret []byte
			logger.Logger().Infof("Loading abi: %s", info.Name())
			ret, err = os.ReadFile(path)
			if err != nil {
				return err
			}
			// ERC20
			if checkType(ret, types.ERC20) {
				preloadAsset.Abis[types.ERC20] = ret
				return nil
			}
			// ERC721
			preloadAsset.Abis[types.ERC721] = ret
		}
		return nil
	})
	if err != nil {
		logger.Logger().Errorf("preloadABIs filepath.Walk() returned %v", err)
		os.Exit(0)
	}
	logger.Logger().Infof("ABIs loaded: %d", len(preloadAsset.Abis))
}

// preloadTokens Preload Tokens
func preloadTokens() {
	err := filepath.Walk("./data/tokens", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".json") {
			var ret []byte
			logger.Logger().Infof("Loading token metadata: %s", info.Name())
			ret, err = os.ReadFile(path)
			if err != nil {
				return err
			}
			var tokens []*models.Token
			err = json.Unmarshal(ret, &tokens)
			if err != nil {
				return err
			}
			// ERC20
			if strings.HasPrefix(info.Name(), "erc20") {
				for _, token := range tokens {
					preloadAsset.ERC20Tokens[strings.ToLower(token.Address)] = token
				}
				return nil
			}
			// ERC721
			for _, token := range tokens {
				preloadAsset.ERC721Tokens[strings.ToLower(token.Address)] = token
			}
		}
		return nil
	})
	if err != nil {
		logger.Logger().Errorf("preloadTokens filepath.Walk() returned %v", err)
		os.Exit(0)
	}
	logger.Logger().Infof("ERC20Tokens loaded: %d", len(preloadAsset.ERC20Tokens))
	logger.Logger().Infof("ERC721Tokens loaded: %d", len(preloadAsset.ERC721Tokens))
}

// preloadSignatures Preload Signatures
func preloadSignatures() {
	var ret []*models.Signatures
	data, err := os.ReadFile("./data/signatures.json")
	if err != nil {
		logger.Logger().Errorf("preloadSignatures read signatures.json returned %v", err)
		os.Exit(0)
	}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		logger.Logger().Errorf("preloadSignatures json unmarshal returned %v", err)
		os.Exit(0)
	}
	for _, signature := range ret {
		preloadAsset.Signatures[strings.ToLower(signature.Hash)] = signature
	}
	logger.Logger().Infof("Signatures loaded: %d", len(preloadAsset.Signatures))
}

// checkType Check is ERC20 / ERC721 ABI
func checkType(data []byte, tokenType types.TokenType) bool {
	count := gjson.Get(string(data), "[#(name=supportsInterface)#|#(type=function)].#").Int()
	switch tokenType {
	case types.ERC20:
		return count <= 0
	case types.ERC721:
		return count > 0
	default:
	}
	return false
}
