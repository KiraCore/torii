package model

type MakeTxRequestBody struct {
	NodeAddress string `json:"node_address"`
	Sender      string `json:"sender"`
	From        string `json:"from"`
	To          string `json:"to"`
	ChainID     string `json:"chain_id"`
	Memo        string `json:"memo"`
	Amount      int64  `json:"amount"`
	GasLimit    int64  `json:"gas_limit"`
	FeeAmount   int64  `json:"fee_amount"`
	Signature   string `json:"signature"`
}
