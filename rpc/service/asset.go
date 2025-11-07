package service

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	eCommon "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/types/known/anypb"
	"io"
	"net/http"
	"strings"
	"time"
	"tokenscope/common/assets"
	"tokenscope/rpc/pb"
	"tokenscope/types"
)

/*
   Created by zyx
   Date Time: 2025/10/24
   File: asset.go
*/

type Asset struct {
	pb.UnimplementedAssetServiceServer
}

func AssetService() *Asset {
	return &Asset{}
}

func (Asset) GetAsset(ctx context.Context, request *pb.GetAssetRequest) (*pb.GetAssetResponse, error) {
	address := eCommon.HexToAddress(request.Address).Hex()
	ret := assets.GetAssetByAddress(address)
	erc20Tokens := make([]*pb.Token, len(ret.ERC20Tokens))
	erc721Tokens := make([]*pb.Token, len(ret.ERC721Tokens))
	txs := make([]*pb.Tx, len(ret.Txs))
	for i, token := range ret.ERC20Tokens {
		erc20Tokens[i] = &pb.Token{
			Name:        token.Name,
			Description: token.Description,
			Address:     token.Address,
			Symbol:      token.Symbol,
			Link:        token.Link,
			Decimals:    uint32(token.Decimals),
			Type:        token.Type,
		}
	}
	for i, token := range ret.ERC721Tokens {
		erc721Tokens[i] = &pb.Token{
			Name:        token.Name,
			Description: token.Description,
			Address:     token.Address,
			Symbol:      token.Symbol,
			Link:        token.Link,
			Decimals:    uint32(token.Decimals),
			Type:        token.Type,
		}
	}
	for i, tx := range ret.Txs {
		txs[i] = &pb.Tx{
			Hash:      tx.Hash,
			From:      tx.From,
			To:        tx.To,
			Value:     tx.Value,
			Data:      string(tx.Data),
			Timestamp: tx.Timestamp,
		}
	}
	return &pb.GetAssetResponse{
		Address:      ret.Address,
		Balance:      ret.Balance,
		Erc20Tokens:  erc20Tokens,
		Erc721Tokens: erc721Tokens,
		Txs:          txs,
	}, nil
}

func (Asset) GetMetadata(ctx context.Context, request *pb.GetMetadataRequest) (*anypb.Any, error) {
	token := assets.GetERC721Token(request.TokenAddress)
	if token == nil {
		return nil, errors.New("token not found")
	}
	tokenURI := assets.GetERC721TokenURI(request.TokenAddress)
	cli := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	// IPFS
	if tokenURI == types.None.String() {
		return nil, errors.New("token metadata not found")
	}
	if strings.HasPrefix(tokenURI, "ipfs://") {
		erc721Url := viper.GetString("image.erc721-url")
		tokenURI = strings.ReplaceAll(tokenURI, "ipfs://", "")
		tokenURI = strings.ReplaceAll(tokenURI, "/", "")
		tokenURI = strings.ReplaceAll(erc721Url, "#{cid}", tokenURI)
	}
	resp, err := cli.Get(fmt.Sprintf("%v%v", tokenURI, request.TokenId))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &anypb.Any{
		Value: body,
	}, nil
}

func (Asset) GetERC20TokenAssets(ctx context.Context, request *pb.GetERC20TokenAssetsRequest) (*pb.TokenAsset, error) {
	ret, err := assets.GetERC20TokenAssets(request.Address, request.TokenAddress)
	if err != nil {
		return nil, err
	}
	return &pb.TokenAsset{
		Token: &pb.Token{
			Name:        ret.Token.Name,
			Description: ret.Token.Description,
			Address:     ret.Token.Address,
			Symbol:      ret.Token.Symbol,
			Link:        ret.Token.Link,
			Decimals:    uint32(ret.Token.Decimals),
			Type:        ret.Token.TokenType.String(),
		},
		Value: ret.Value,
		Type:  ret.TokenType.String(),
	}, nil
}

func (Asset) GetERC721TokenAssets(ctx context.Context, request *pb.GetERC721TokenAssetsRequest) (*pb.GetERC721TokenAssetsResponse, error) {
	addr := eCommon.HexToAddress(request.Address)
	tokenAddr := eCommon.HexToAddress(request.TokenAddress)
	ret, err := assets.GetERC721TokenAssets(addr.Hex(), tokenAddr.Hex())
	if err != nil {
		return nil, err
	}
	tokens := make([]*pb.NFTAsset, 0)
	for _, token := range ret {
		tokens = append(tokens, &pb.NFTAsset{
			Token: &pb.Token{
				Name:        token.Token.Name,
				Description: token.Token.Description,
				Address:     token.Token.Address,
				Symbol:      token.Token.Symbol,
				Link:        token.Token.Link,
				Decimals:    uint32(token.Token.Decimals),
				Type:        token.Token.TokenType.String(),
			},
			TokenId:     token.TokenId,
			MetadataUrl: token.MetadataUrl,
			SpecialUrl:  token.SpecialUrl,
			TokenType:   token.TokenType.String(),
		})
	}
	return &pb.GetERC721TokenAssetsResponse{
		Erc721Tokens: tokens,
	}, nil
}
