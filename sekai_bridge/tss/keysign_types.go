package tss

import (
	"math/big"
	"sync"
	"sync/atomic"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/binance-chain/tss-lib/ecdsa/signing"
	tsslib "github.com/binance-chain/tss-lib/tss"
	p2p "github.com/saiset-co/saiP2P-go/core"
	"go.uber.org/zap"
)

type TssKeySign struct {
	Logger             *zap.Logger
	Pubkey             string          `json:"pubkey,omitempty"`
	LocalPartyID       *tsslib.PartyID `json:"local_partyID,omitempty"`
	StopChan           chan CommunicationError
	OutCh              chan tsslib.Message
	EndCh              chan *signing.SignatureData
	ErrCh              chan *tsslib.Error
	P2pComm            *p2p.Core
	Parties            int
	Quorum             int
	PS                 tsslib.Party
	ConnectionStorage  map[string]string
	KeysignMsgsStorage *KeysignMsgsStorage // storage for keygen msgs from another nodes
	Key                *keygen.LocalPartySaveData
	OneRoundMsgCh      chan *P2pMessage // to get info from others peers to make one round signing
	IsStarted          atomic.Bool
}

type KeysignMsgsStorage struct { // storage for keygen msgs from another nodes
	*sync.RWMutex
	M map[string]TssMessage
}

type SignMessageRequest struct {
	Msg             string `json:"msg"`
	OneRoundSigning bool   `json:"one_round_signing"`
}

type SignMessageResponse struct {
	SignatureMarshalled []byte `json:"signature"`
}

// to exchange info to construct map[*tss.PartyID]S_i
type OneRoundSigningMsg struct {
	PartyID *tsslib.PartyID
	Si      *big.Int
}
