package types

import (
	"github.com/binance-chain/tss-lib/tss"
	"github.com/go-playground/validator"
)

var validate *validator.Validate

// initialize validator
func init() {
	validate = validator.New()
}

type GenerateKeysRequest struct {
	Threshold int `json:"threshold" validate:"required"`
	Parties   int `json:"parties" validate:"required"`
}

func (g *GenerateKeysRequest) Validate() error {
	return validate.Struct(g)
}

type SignMessageRequest struct {
	Msg     string `json:"msg"`
	Parties int    `json:"parties"`
	Quorum  int    `json:"quorum"`
}

// for debbuging sekaiBridge app
type SekaiBridgeStats struct {
	HttpPort             string            `json:"http_port"`
	P2PPort              string            `json:"p2p_port"`
	P2PSlot              int               `json:"p2p_slot"`
	P2pPeers             []string          `json:"p2p_peers"`
	P2PConnectionStorage map[string]bool   `json:"p2p_connection_storage"` // p2p connections listed here
	P2PSavedMessages     map[string]bool   `json:"p2p_saved_messages"`     // saved messages, to prevent double messages sending
	TssPartyID           tss.PartyID       `json:"tss_partyID"`
	TssConnectionStorage map[string]string `json:"tss_connection_storage"` // map[peerAddr]PartyID
	TssKeygenStarted     bool              `json:"tss_keygen_started"`     // is keygen started
	TssPartiesMap        map[string]bool   `json:"tss_parties"`            // parties map
}
