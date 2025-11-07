package models

/*
   Created by zyx
   Date Time: 2025/9/18
   File: signatures.go
*/

type Mehod struct {
	Method string `json:"method"`
	Params []any  `json:"params"`
}

type Signatures struct {
	Source  string   `json:"source"`
	Methods []*Mehod `json:"methods"`
	Hash    string   `json:"hash"`
	Hex     string   `json:"hex"`
}
