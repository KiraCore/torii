package types

import (
	"github.com/binance-chain/tss-lib/tss"
)

// for debbuging sekaiBridge app
type SekaiBridgeStats struct {
	HTTPPort             string            `json:"http_port"`
	P2PPort              string            `json:"p2p_port"`
	P2PSlot              int               `json:"p2p_slot"`
	P2pPeers             []string          `json:"p2p_peers"`
	P2PConnectionStorage map[string]bool   `json:"p2p_connection_storage"` // p2p connections listed here
	P2PSavedMessages     map[string]bool   `json:"p2p_saved_messages"`     // saved messages, to prevent double messages sending
	TssPartyID           tss.PartyID       `json:"tss_partyID"`
	TssConnectionStorage map[string]string `json:"tss_connection_storage"` // map[peerAddr]PartyID
	TssKeygenStarted     bool              `json:"tss_keygen_started"`     // is keygen started
	TssPartiesMap        map[string]bool   `json:"tss_parties"`            // parties map
	TssKeygenMsgStorage  map[string]bool   `json:"tss_keygen_msgs"`        // keygen messages
}

type Config struct {
	P2P struct {
		Port  string   `yaml:"port"`
		Slot  int      `yaml:"slot"`
		Peers []string `yaml:"peers"`
	} `yaml:"p2p"`
	HTTP struct {
		Enabled bool   `yaml:"enabled"`
		Port    string `yaml:"port"`
	} `yaml:"http"`
	Tss struct {
		PublicKey string `yaml:"public_key"`
		Parties   int    `yaml:"parties"`
		Threshold int    `yaml:"threshold"`
		Quorum    int    `yaml:"quorum"`
	} `yaml:"tss"`

	OnBroadcastMessageReceive []string
	OnDirectMessageReceive    []string
	DebugMode                 bool `yaml:"debug"`
}

type CosmosInteractionData struct {
	Type        string `json:"type"`
	NodeAddress string `json:"node_address"`
	Sender      string `json:"sender"`
	From        string `json:"from"`
	To          string `json:"to"`
	ChainId     string `json:"chain_id"`
	Memo        string `json:"memo"`
	Amount      int    `json:"amount"`
	GasLimit    int    `json:"gas_limit"`
	FeeAmount   int    `json:"fee_amount"`
	Signature   []byte `json:"signature"`
}

type CosmosInteractionRequest struct {
	Method   string                `json:"method"`
	Data     CosmosInteractionData `json:"data"`
	Metadata interface{}           `json:"metadata"`
}

type EthInteractionParam struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type EthInteractionData struct {
	Contract  string                `json:"contract"`
	Method    string                `json:"method"`
	Value     string                `json:"value"`
	Params    []EthInteractionParam `json:"params"`
	Signature []byte                `json:"signature"`
}

type EthInteractionRequest struct {
	Method   string             `json:"method"`
	Data     EthInteractionData `json:"data"`
	Metadata interface{}        `json:"metadata"`
}
