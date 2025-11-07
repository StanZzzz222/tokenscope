package handler

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	eCommon "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"tokenscope/common/assets"
	"tokenscope/models"
	"tokenscope/types"
)

/*
   Created by zyx
   Date Time: 2025/9/17
   File: asset.go
*/

func InitAssetRouters(r *gin.RouterGroup) {
	asset := r.Group("/asset")
	{
		asset.GET("/:address", getAssetHandler)
		asset.GET("/metadata/:tokenAddress/:tokenId", getMetadataHandler)
		asset.GET("erc20_token_assets/:address/:tokenAddress", getERC20TokenAssetsHandler)
		asset.GET("erc721_token_assets/:address/:tokenAddress", getERC721TokenAssetsHandler)
		asset.GET("icon/:address", getIconHandler)
	}
}

func getAssetHandler(ctx *gin.Context) {
	address := ctx.Param("address")
	ret := assets.GetAssetByAddress(address)
	if ret == nil {
		models.Response(ctx).Error(http.StatusInternalServerError, errors.New("address format is incorrect"))
		return
	}
	models.Response(ctx).Success(http.StatusOK, ret)
}

func getMetadataHandler(ctx *gin.Context) {
	tokenAddress := ctx.Param("tokenAddress")
	tokenId := ctx.Param("tokenId")
	token := assets.GetERC721Token(tokenAddress)
	if token == nil {
		models.Response(ctx).Error(http.StatusInternalServerError, errors.New("token not found"))
		return
	}
	tokenURI := assets.GetERC721TokenURI(tokenAddress)
	if tokenURI == types.None.String() {
		models.Response(ctx).Error(http.StatusInternalServerError, errors.New("token metadata not found"))
		return
	}
	cli := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	// IPFS
	if strings.HasPrefix(tokenURI, "ipfs://") {
		erc721Url := viper.GetString("image.erc721-url")
		tokenURI = strings.ReplaceAll(tokenURI, "ipfs://", "")
		tokenURI = strings.ReplaceAll(tokenURI, "/", "")
		tokenURI = strings.ReplaceAll(erc721Url, "#{cid}", tokenURI)
	}
	data := new(any)
	resp, err := cli.Get(fmt.Sprintf("%v%v", tokenURI, tokenId))
	if err != nil {
		models.Response(ctx).Error(http.StatusInternalServerError, err)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		models.Response(ctx).Error(http.StatusInternalServerError, err)
		return
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		models.Response(ctx).Error(http.StatusInternalServerError, err)
		return
	}
	models.Response(ctx).Success(http.StatusOK, data)
}

func getERC20TokenAssetsHandler(ctx *gin.Context) {
	address := ctx.Param("address")
	tokenAddress := ctx.Param("tokenAddress")
	ret, err := assets.GetERC20TokenAssets(address, tokenAddress)
	if err != nil {
		models.Response(ctx).Error(http.StatusInternalServerError, err)
		return
	}
	models.Response(ctx).Success(http.StatusOK, ret)
}

func getERC721TokenAssetsHandler(ctx *gin.Context) {
	address := ctx.Param("address")
	tokenAddress := ctx.Param("tokenAddress")
	addr := eCommon.HexToAddress(address)
	tokenAddr := eCommon.HexToAddress(tokenAddress)
	ret, err := assets.GetERC721TokenAssets(addr.Hex(), tokenAddr.Hex())
	if err != nil {
		models.Response(ctx).Error(http.StatusInternalServerError, err)
		return
	}
	models.Response(ctx).Success(http.StatusOK, ret)
}

func getIconHandler(ctx *gin.Context) {
	address := ctx.Param("address")
	chainId := viper.GetInt("rpc.chain-id")
	iconUrl := viper.GetString("image.erc20-url")
	iconUrl = strings.ReplaceAll(iconUrl, "#{chainId}", fmt.Sprintf("%v", chainId))
	iconUrl = strings.ReplaceAll(iconUrl, "#{tokenAddress}", strings.ToLower(address))
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Get(iconUrl)
	if err != nil {
		models.Response(ctx).Error(http.StatusInternalServerError, err)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		models.Response(ctx).Error(http.StatusInternalServerError, err)
		return
	}
	if resp.StatusCode == http.StatusNotFound || strings.Contains(strings.ToLower(string(body)), "error") {
		body, err = os.ReadFile("./data/unknow.png")
		if err != nil {
			models.Response(ctx).Error(http.StatusInternalServerError, err)
			return
		}
	}
	models.Response(ctx).Image(body)
}
