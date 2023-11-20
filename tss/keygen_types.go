package tss

import (
	"sync"
	"sync/atomic"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	tsslib "github.com/binance-chain/tss-lib/tss"
	p2p "github.com/saiset-co/saiP2P-go/core"
	"go.uber.org/zap"
)

type TssKeyGen struct {
	Logger       *zap.Logger
	Pubkey       string `json:"pubkey,omitempty"`
	Parties      int
	Threshold    int
	LocalPartyID *tsslib.PartyID         `json:"local_partyID,omitempty"`
	StopChan     chan CommunicationError // channel to indicate whether we should stop
	// PartiesMap        map[tsslib.PartyID]bool
	CommStopChan      chan struct{}
	OutCh             chan tsslib.Message
	EndCh             chan keygen.LocalPartySaveData
	ErrCh             chan *tsslib.Error
	P2pComm           *p2p.Core
	Key               *keygen.LocalPartySaveData // generated key
	IsStarted         atomic.Bool                `json:"is_started"` // is keygen was already started
	ConnectionStorage map[string]string          // map[pubkey]peerAddr
	PG                tsslib.Party               `json:"-"`
	*sync.RWMutex                                // mutex for map[keygenMsg]bool
	KeygenMsgsStorage *KeygenMsgsStorage         // storage for keygen msgs from another nodes
}

type KeygenMsgsStorage struct {
	*sync.RWMutex
	M map[string]TssMessage
}

// Response keygen response
type Response struct {
	Key *keygen.LocalPartySaveData `json:"pubkey"`
}
