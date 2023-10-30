package tss

import (
	"sync"

	"github.com/KiraCore/sekai-bridge/types"
	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/binance-chain/tss-lib/tss"

	tsslib "github.com/binance-chain/tss-lib/tss"
	p2p "github.com/saiset-co/saiP2P-go/core"
	"go.uber.org/zap"
)

const (
	HandshakeMsgType   = "tss_handshake_msg"    // for registering other sekai-bridge instanses
	KeygenMsgType      = "tss_keygen_msg"       // for keygen exchanging messages
	KeygenStartMsgType = "tss_keygen_start_msg" // for start keygen
)

// main tss struct
type TssServer struct {
	LocalPartyID *tss.PartyID `json:"local_partyID,omitempty"`
	Pubkey       string       `json:"pubkey,omitempty"`
	*sync.RWMutex
	ConnectionStorage map[string]string // map[pubkey]peerAddr
	Logger            *zap.Logger
	P2p               *p2p.Core
	PG                *keygen.LocalParty `json:"-"`
	KeygenStarted     bool               `json:"keygen_started,omitempty"` // is keygen started flag
	Key               *keygen.LocalPartySaveData
	TssKeygen         *TssKeyGen `json:"tss_keygen,omitempty"`
	//EndChS            chan *signing.SignatureData
}

type TssKeyGen struct {
	Logger                      *zap.Logger
	Pubkey                      string        `json:"pubkey,omitempty"`
	LocalPartyID                *tss.PartyID  `json:"local_partyID,omitempty"`
	StopChan                    chan struct{} // channel to indicate whether we should stop
	PartiesMap                  map[tss.PartyID]bool
	CommStopChan                chan struct{}
	OutCh                       chan tsslib.Message
	EndCh                       chan keygen.LocalPartySaveData
	P2pComm                     *p2p.Core
	Key                         *keygen.LocalPartySaveData // generated key
	IsStarted                   bool                       `json:"is_started"` // is keygen was already started
	ConnectionStorage           map[string]string          // map[pubkey]peerAddr
	CachedWireBroadcastMsgLists *sync.Map
	CachedWireUnicastMsgLists   *sync.Map
}

// tss message struct
type Message struct {
	From        *tsslib.PartyID   `json:"from"`
	To          []*tsslib.PartyID `json:"to"`
	IsBroadcast bool              `json:"is_broadcast"`
	Bytes       []byte            `json:"bytes"`
	Type        string            `json:"type"`
}

// message to communicate through p2p
// for example to register id, send tss messages through p2p, initiate keygen ...
type P2pMessage struct {
	TssMessage    tsslib.Message            `json:"tss_message,omitempty"`  //tss message
	Type          string                    `json:"type,omitempty"`         //message type
	PeerAddr      string                    `json:"peer_addr,omitempty"`    // tss peerAddr
	Pubkey        string                    `json:"pubkey,omitempty"`       //tss party id
	KeygenRound   string                    `json:"keygen_round,omitempty"` // keygen round
	KeygenRequest types.GenerateKeysRequest `json:"keygen_request,omitempty"`
}

func NewTssKeyGen(partyID int, logger *zap.Logger) *TssKeyGen {
	return &TssKeyGen{}
}

// Response keygen response
type Response struct {
	Key *keygen.LocalPartySaveData `json:"pubkey"`
}

type BulkWireMsg struct {
	WiredBulkMsgs []byte
	MsgIdentifier string
	Routing       *tsslib.MessageRouting
}
