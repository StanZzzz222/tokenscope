package models

import (
	"bytes"
	"encoding/binary"
	"github.com/bits-and-blooms/bloom/v3"
	"tokenscope/common/logger"
)

/*
   Created by zyx
   Date Time: 2025/9/25
   File: tx_log.go
*/

type TxEvent struct {
	ContractAddress string   `json:"contract_address"`
	MethodHash      string   `json:"method_hash"`
	Values          []string `json:"values"`
}

type TxLog struct {
	BlockNumber uint64             `json:"block_number"`
	TxEvents    []*TxEvent         `json:"events"`
	BloomData   []byte             `json:"bloom_data"`
	Bloom       *bloom.BloomFilter `json:"-"`
}

func NewTxLog(blockNumber uint64, txEvents []*TxEvent) *TxLog {
	ret := &TxLog{
		BlockNumber: blockNumber,
		TxEvents:    txEvents,
		BloomData:   make([]byte, 0),
		Bloom:       nil,
	}
	addTxLogBloomFilter(ret)
	if data, err := ret.Bloom.MarshalBinary(); err == nil {
		ret.BloomData = data
	}
	return ret
}

func NewTxEvent(contractAddress, methodHash string, values []string) *TxEvent {
	ret := &TxEvent{
		ContractAddress: contractAddress,
		MethodHash:      methodHash,
		Values:          values,
	}
	return ret
}

func (t *TxLog) Marshal() []byte {
	buf := bytes.NewBuffer(nil)
	writeString := func(s string) {
		bs := []byte(s)
		_ = binary.Write(buf, binary.BigEndian, uint16(len(bs)))
		buf.Write(bs)
	}
	_ = binary.Write(buf, binary.BigEndian, t.BlockNumber)
	_ = binary.Write(buf, binary.BigEndian, uint32(len(t.TxEvents)))
	for _, ev := range t.TxEvents {
		writeString(ev.ContractAddress)
		writeString(ev.MethodHash)
		_ = binary.Write(buf, binary.BigEndian, uint32(len(ev.Values)))
		for _, v := range ev.Values {
			writeString(v)
		}
	}
	_ = binary.Write(buf, binary.BigEndian, uint32(len(t.BloomData)))
	buf.Write(t.BloomData)
	return buf.Bytes()
}

func UnmarshalTxLog(data []byte) *TxLog {
	ret := &TxLog{}
	buf := bytes.NewReader(data)
	readString := func() string {
		var l uint16
		_ = binary.Read(buf, binary.BigEndian, &l)
		bs := make([]byte, l)
		_, _ = buf.Read(bs)
		return string(bs)
	}
	// BlockNumber
	_ = binary.Read(buf, binary.BigEndian, &ret.BlockNumber)
	// TxEvents
	var eventCount uint32
	_ = binary.Read(buf, binary.BigEndian, &eventCount)
	ret.TxEvents = make([]*TxEvent, eventCount)
	for i := 0; i < int(eventCount); i++ {
		ev := &TxEvent{}
		ev.ContractAddress = readString()
		ev.MethodHash = readString()
		var valueCount uint32
		_ = binary.Read(buf, binary.BigEndian, &valueCount)
		ev.Values = make([]string, valueCount)
		for j := 0; j < int(valueCount); j++ {
			ev.Values[j] = readString()
		}
		ret.TxEvents[i] = ev
	}
	// BloomData
	var bloomLen uint32
	_ = binary.Read(buf, binary.BigEndian, &bloomLen)
	ret.BloomData = make([]byte, bloomLen)
	_, _ = buf.Read(ret.BloomData)
	if len(ret.BloomData) > 0 {
		bf := new(bloom.BloomFilter)
		if err := bf.UnmarshalBinary(ret.BloomData); err != nil {
			logger.Logger().Errorf("Unmarshal TxLog Bloom err:%v", err)
		} else {
			ret.Bloom = bf
		}
	}
	return ret
}

func (t *TxLog) HasBloom(data string) bool {
	if t.Bloom == nil {
		return false
	}
	return t.Bloom.TestString(data)
}

func addTxLogBloomFilter(txLog *TxLog) {
	seen := make(map[string]struct{})
	txLog.Bloom = bloom.NewWithEstimates(2048, 0.01)
	add := func(s string) {
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		txLog.Bloom.AddString(s)
	}
	for _, txEvent := range txLog.TxEvents {
		add(txEvent.ContractAddress)
		add(txEvent.MethodHash)
		for _, value := range txEvent.Values {
			add(value)
		}
	}
}
