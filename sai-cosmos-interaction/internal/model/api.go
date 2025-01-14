package model

type AccountInfo struct {
	Account struct {
		Type          string      `json:"@type"`
		Address       string      `json:"address"`
		PubKey        interface{} `json:"pub_key"`
		AccountNumber string      `json:"account_number"`
		Sequence      string      `json:"sequence"`
	} `json:"account"`
}

type TxBroadcastReq struct {
	TxBytes string `json:"tx_bytes"`
	Mode    string `json:"mode"`
}

type TxBroadcastRes struct {
	TxResponse TxResponse `json:"tx_response"`
}

type TxResponse struct {
	Height    string        `json:"height"`
	Txhash    string        `json:"txhash"`
	Codespace string        `json:"codespace"`
	Code      int           `json:"code"`
	Data      string        `json:"data"`
	RawLog    string        `json:"raw_log"`
	Logs      []interface{} `json:"logs"`
	Info      string        `json:"info"`
	GasWanted string        `json:"gas_wanted"`
	GasUsed   string        `json:"gas_used"`
	Tx        interface{}   `json:"tx"`
	Timestamp string        `json:"timestamp"`
	Events    []interface{} `json:"events"`
}
