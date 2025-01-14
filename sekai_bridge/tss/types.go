package tss

import (
	"math/big"
	"sync"

	keygenlib "github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/binance-chain/tss-lib/ecdsa/signing"

	tsslib "github.com/binance-chain/tss-lib/tss"
	p2p "github.com/saiset-co/saiP2P-go/core"
	"go.uber.org/zap"
)

const (
	HandshakeMsgType  = "tss_handshake_msg"  // for registering other sekai-bridge instanses
	DisconnectMsgType = "tss_disconnect_msg" // for deregister other sekai-bridge instanses

	KeygenMsgType          = "tss_keygen_msg"        // for keygen exchanging messages
	KeygenStartMsgType     = "tss_keygen_start_msg"  // for start keygen
	KeygenCancelledMsgType = "tss_keygen_cancel_msg" // for cancelling keygen due error at some peer

	KeysignMsgType          = "tss_keysign_msg"       // for keygen exchanging messages
	KeysignStartMsgType     = "tss_keysign_start_msg" // for start keygen
	KeysignOneRoundMsgType  = "tss_keysing_one_round"
	KeysignCancelledMsgType = "tss_keysign_cancel_msg" // for cancelling keygen due error at some peer
)

// main tss struct
type TssServer struct {
	LocalPartyID *tsslib.PartyID `json:"local_partyID,omitempty"`
	Pubkey       string          `json:"pubkey,omitempty"`
	Parties      int
	Threshold    int
	Quorum       int
	*sync.RWMutex
	ConnectionStorage map[string]string // map[pubkey]peerAddr
	Logger            *zap.Logger
	P2p               *p2p.Core
	PG                tsslib.Party `json:"-"`
	PS                *signing.LocalParty
	Key               *keygenlib.LocalPartySaveData
	KeygenInstance    *TssKeyGen `json:"tss_keygen,omitempty"`
	KeysignInstance   *TssKeySign
	// CommStopChan      chan struct{}
	// OutCh             chan tsslib.Message
	// ErrCh             chan *tsslib.Error
	PartiesMap        map[tsslib.PartyID]bool
	StopChan          chan struct{} // channel to indicate whether we should stop
	BufferedKeygenMsg *P2pMessage   // buffered p2p message for keygen
	ErrorMsgMap       map[string]bool
}

// tss message struct
type TssMessage struct {
	From        *tsslib.PartyID        `json:"from"`
	To          []*tsslib.PartyID      `json:"to"`
	IsBroadcast bool                   `json:"is_broadcast"`
	Bytes       []byte                 `json:"bytes"`
	Type        string                 `json:"type"`
	Routing     *tsslib.MessageRouting `json:"routing"`
}

// message to communicate through p2p
// for example to register id, send tss messages through p2p, initiate keygen ...
type P2pMessage struct {
	TssMsg             *TssMessage         `json:"tss_message,omitempty"`     // tss message
	Type               string              `json:"type,omitempty"`            // message type
	PeerAddr           string              `json:"peer_addr,omitempty"`       // tss peerAddr
	Pubkey             string              `json:"pubkey,omitempty"`          // tss party id
	Round              string              `json:"round,omitempty"`           // keygen round
	KeysignRequest     *SignMessageRequest `json:"keysign_request,omitempty"` // message to sign
	Si                 *big.Int            `json:"si,omitempty"`              // si for one round signing
	PartyID            *tsslib.PartyID     `json:"party_id,omitempty"`
	Time               int64               `json:"sent_time,omitempty"`           // sent time, to avoid filtering (for start keygen msg)
	CommunicationError CommunicationError  `json:"communication_error,omitempty"` // when communication error got
}

// to detect in which operation error was occured
type P2pMessageSimple struct {
	Type   string `json:"type,omitempty"`   // message type
	Pubkey string `json:"pubkey,omitempty"` // tss party id
}
