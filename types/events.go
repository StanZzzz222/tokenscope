package types

import "github.com/ethereum/go-ethereum/core/types"

/*
   Created by zyx
   Date Time: 2025/9/28
   File: events.go
*/

type BlockListenEvent func(block *types.Block)
