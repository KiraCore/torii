package model

import (
	"encoding/json"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	jsoniter "github.com/json-iterator/go"
)

type BlockInfo struct {
	BlockId struct {
		Hash  string `json:"hash"`
		Parts struct {
			Total int    `json:"total"`
			Hash  string `json:"hash"`
		} `json:"parts"`
	} `json:"block_id"`
	Block struct {
		Header struct {
			Version struct {
				Block string `json:"block"`
			} `json:"version"`
			ChainId     string    `json:"chain_id"`
			Height      string    `json:"height"`
			Time        time.Time `json:"time"`
			LastBlockId struct {
				Hash  string `json:"hash"`
				Parts struct {
					Total int    `json:"total"`
					Hash  string `json:"hash"`
				} `json:"parts"`
			} `json:"last_block_id"`
			LastCommitHash     string `json:"last_commit_hash"`
			DataHash           string `json:"data_hash"`
			ValidatorsHash     string `json:"validators_hash"`
			NextValidatorsHash string `json:"next_validators_hash"`
			ConsensusHash      string `json:"consensus_hash"`
			AppHash            string `json:"app_hash"`
			LastResultsHash    string `json:"last_results_hash"`
			EvidenceHash       string `json:"evidence_hash"`
			ProposerAddress    string `json:"proposer_address"`
		} `json:"header"`
		Data struct {
			Txs []string `json:"txs"`
		} `json:"data"`
		Evidence struct {
			Evidence []interface{} `json:"evidence"`
		} `json:"evidence"`
		LastCommit struct {
			Height  string `json:"height"`
			Round   int    `json:"round"`
			BlockId struct {
				Hash  string `json:"hash"`
				Parts struct {
					Total int    `json:"total"`
					Hash  string `json:"hash"`
				} `json:"parts"`
			} `json:"block_id"`
			Signatures []struct {
				BlockIdFlag      int       `json:"block_id_flag"`
				ValidatorAddress string    `json:"validator_address"`
				Timestamp        time.Time `json:"timestamp"`
				Signature        string    `json:"signature"`
			} `json:"signatures"`
		} `json:"last_commit"`
	} `json:"block"`
}

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
	Timestamp time.Time `json:"timestamp"`
	Hash      string    `json:"hash"`
	Height    string    `json:"height"`
	Index     int       `json:"index"`
	TxResult  struct {
		Code      int    `json:"code"`
		Data      string `json:"data"`
		Log       string `json:"log"`
		Info      string `json:"info"`
		GasWanted string `json:"gas_wanted"`
		GasUsed   string `json:"gas_used"`
		Events    []struct {
			Type       string `json:"type"`
			Attributes []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
				Index bool   `json:"index"`
			} `json:"attributes"`
		} `json:"events"`
		Codespace string `json:"codespace"`
	} `json:"tx_result"`
	Tx       string        `json:"tx"`
	Messages []interface{} `json:"messages"`
}

type Message map[string]interface{}

type Amount struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type RPCResponse struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   interface{}     `json:"error,omitempty"`
}

type TxsResultResponse struct {
	Transactions []TransactionResultResponse `json:"transactions"`
	TotalCount   int                         `json:"total_count"`
}

type TransactionResultResponse struct {
	Time      int64         `json:"time"`
	Hash      string        `json:"hash"`
	Status    string        `json:"status"`
	Direction string        `json:"direction"`
	Memo      string        `json:"memo"`
	Fee       sdk.Coins     `json:"fee"`
	Txs       []interface{} `json:"txs"`
}
