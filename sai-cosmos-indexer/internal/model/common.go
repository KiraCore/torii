package model

import (
	"time"

	jsoniter "github.com/json-iterator/go"
)

type BlockTransactions struct {
	Txs         []Tx         `json:"txs"`
	TxResponses []TxResponse `json:"tx_responses"`
	Pagination  Pagination   `json:"pagination"`
}

type Pagination struct {
	NextKey interface{} `json:"next_key"`
	Total   string      `json:"total"`
}

type TxResponse struct {
	Height    string              `json:"height"`
	Txhash    string              `json:"txhash"`
	Codespace string              `json:"codespace"`
	Code      int                 `json:"code"`
	Tx        Tx                  `json:"tx"`
	Timestamp time.Time           `json:"timestamp"`
	Events    jsoniter.RawMessage `json:"events"`
}

type Tx struct {
	Body struct {
		Messages []Message `json:"messages"`
	} `json:"body"`
}

type Message struct {
	Type        string   `json:"@type"`
	FromAddress string   `json:"from"`
	ToAddress   string   `json:"to"`
	Hash        string   `json:"hash"`
	Amount      []Amount `json:"amount"`
}

type Amount struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}
