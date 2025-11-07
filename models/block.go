package models

import (
	"bytes"
	"encoding/binary"
	"github.com/bits-and-blooms/bloom/v3"
	"tokenscope/common/logger"
)

/*
   Created by zyx
   Date Time: 2025/9/17
   File: block.go
*/

type Block struct {
	Hash              string             `json:"hash"`
	ParentHash        string             `json:"parent_hash"`
	UncleHash         string             `json:"uncle_hash"`
	TransactionsRoot  string             `json:"transactions_root"`
	Miner             string             `json:"miner"`
	BlockNumber       uint64             `json:"block_number"`
	Timestamp         uint64             `json:"timestamp"`
	BaseFee           uint64             `json:"base_fee"`
	TransactionsCount uint64             `json:"transactions_count"`
	Txs               []*Tx              `json:"txs"`
	BloomData         []byte             `json:"bloom_data"`
	Bloom             *bloom.BloomFilter `json:"-"`
}

func NewBlock(hash, parentHash, uncleHash, transactionsRoot, miner string, blockNumber, timestamp, baseFee uint64, transactionsCount uint64, txs []*Tx) *Block {
	ret := &Block{
		Hash:              hash,
		ParentHash:        parentHash,
		UncleHash:         uncleHash,
		TransactionsRoot:  transactionsRoot,
		Miner:             miner,
		BlockNumber:       blockNumber,
		Timestamp:         timestamp,
		BaseFee:           baseFee,
		TransactionsCount: transactionsCount,
		Txs:               txs,
		BloomData:         make([]byte, 0),
		Bloom:             nil,
	}
	addBloomFilter(ret)
	if data, err := ret.Bloom.MarshalBinary(); err == nil {
		ret.BloomData = data
	}
	return ret
}

func UnmarshalBlock(data []byte) *Block {
	ret := &Block{}
	buf := bytes.NewReader(data)
	readString := func() string {
		var l uint16
		_ = binary.Read(buf, binary.BigEndian, &l)
		bs := make([]byte, l)
		_, _ = buf.Read(bs)
		return string(bs)
	}
	ret.Hash = readString()
	ret.ParentHash = readString()
	ret.UncleHash = readString()
	ret.TransactionsRoot = readString()
	ret.Miner = readString()
	_ = binary.Read(buf, binary.BigEndian, &ret.BlockNumber)
	_ = binary.Read(buf, binary.BigEndian, &ret.Timestamp)
	_ = binary.Read(buf, binary.BigEndian, &ret.BaseFee)
	_ = binary.Read(buf, binary.BigEndian, &ret.TransactionsCount)
	var txCount uint32
	_ = binary.Read(buf, binary.BigEndian, &txCount)
	ret.Txs = make([]*Tx, txCount)
	for i := 0; i < int(txCount); i++ {
		tx := &Tx{}
		tx.Hash = readString()
		tx.From = readString()
		tx.To = readString()
		_ = binary.Read(buf, binary.BigEndian, &tx.Value)
		var dataLen uint32
		_ = binary.Read(buf, binary.BigEndian, &dataLen)
		tx.Data = make([]byte, dataLen)
		_, _ = buf.Read(tx.Data)
		_ = binary.Read(buf, binary.BigEndian, &tx.Timestamp)
		ret.Txs[i] = tx
	}
	var bloomLen uint32
	_ = binary.Read(buf, binary.BigEndian, &bloomLen)
	ret.BloomData = make([]byte, bloomLen)
	_, _ = buf.Read(ret.BloomData)
	if len(ret.BloomData) > 0 {
		bf := new(bloom.BloomFilter)
		if err := bf.UnmarshalBinary(ret.BloomData); err != nil {
			logger.Logger().Errorf("Unmarshal Block Bloom err:%v", err)
		} else {
			ret.Bloom = bf
		}
	}
	return ret
}

func (b *Block) Marshal() []byte {
	buf := bytes.NewBuffer(nil)
	writeString := func(s string) {
		bs := []byte(s)
		_ = binary.Write(buf, binary.BigEndian, uint16(len(bs)))
		buf.Write(bs)
	}
	writeString(b.Hash)
	writeString(b.ParentHash)
	writeString(b.UncleHash)
	writeString(b.TransactionsRoot)
	writeString(b.Miner)
	_ = binary.Write(buf, binary.BigEndian, b.BlockNumber)
	_ = binary.Write(buf, binary.BigEndian, b.Timestamp)
	_ = binary.Write(buf, binary.BigEndian, b.BaseFee)
	_ = binary.Write(buf, binary.BigEndian, b.TransactionsCount)
	_ = binary.Write(buf, binary.BigEndian, uint32(len(b.Txs)))
	for _, tx := range b.Txs {
		writeString(tx.Hash)
		writeString(tx.From)
		writeString(tx.To)
		_ = binary.Write(buf, binary.BigEndian, tx.Value)
		_ = binary.Write(buf, binary.BigEndian, uint32(len(tx.Data)))
		buf.Write(tx.Data)
		_ = binary.Write(buf, binary.BigEndian, tx.Timestamp)
	}
	_ = binary.Write(buf, binary.BigEndian, uint32(len(b.BloomData)))
	buf.Write(b.BloomData)
	return buf.Bytes()
}

func (b *Block) HasBloom(data string) bool {
	if b.Bloom == nil {
		return false
	}
	return b.Bloom.TestString(data)
}

func addBloomFilter(ret *Block) {
	seen := make(map[string]struct{})
	ret.Bloom = bloom.NewWithEstimates(2048, 0.01)
	add := func(s string) {
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		ret.Bloom.AddString(s)
	}
	for _, tx := range ret.Txs {
		if tx != nil {
			abiData := DecodeABIData(tx.Data)
			if abiData != nil {
				for _, addr := range abiData.AddressValues {
					add(addr)
				}
			}
			add(tx.From)
			add(tx.To)
			add(tx.Hash)
		}
	}
	add(ret.Hash)
	add(ret.Miner)
}
