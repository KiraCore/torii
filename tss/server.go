package tss

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/binance-chain/tss-lib/ecdsa/signing"
	tsslib "github.com/binance-chain/tss-lib/tss"

	p2p "github.com/saiset-co/saiP2P-go/core"
	"go.uber.org/zap"
)

// tssServer instance initializating
func New(pubkey string, parties, threshold, quorum int, p2p *p2p.Core, l *zap.Logger) *TssServer {
	partyID := PubkeyToPartyID(pubkey)

	return &TssServer{
		ConnectionStorage: make(map[string]string),
		Logger:            l,
		P2p:               p2p,
		RWMutex:           new(sync.RWMutex),
		LocalPartyID:      partyID,
		Pubkey:            pubkey,
		Parties:           parties,
		Threshold:         threshold,
		Quorum:            quorum,
		// t.EndChS = make(chan *signing.SignatureData, 1)
		StopChan:   make(chan struct{}),
		PartiesMap: make(map[tsslib.PartyID]bool),
		// CommStopChan: make(chan struct{}),
		// OutCh:        make(chan tsslib.Message, 2), // @TODO: parties length here,
		// ErrCh:        make(chan *tsslib.Error),
		ErrorMsgMap: map[string]bool{},
	}
}

// tssKeyGen instance initializating
func (t *TssServer) NewTssKeyGen(parties, threshold int) {
	t.KeygenInstance = &TssKeyGen{
		Logger:            t.Logger,
		LocalPartyID:      t.LocalPartyID,
		Pubkey:            t.Pubkey,
		Parties:           parties,
		Threshold:         threshold,
		ConnectionStorage: t.ConnectionStorage,
		// PartiesMap:        map[tsslib.PartyID]bool{},
		P2pComm: t.P2p,
		RWMutex: new(sync.RWMutex),
		KeygenMsgsStorage: &KeygenMsgsStorage{
			RWMutex: new(sync.RWMutex),
			M:       make(map[string]TssMessage),
		},
		ErrCh:    make(chan *tsslib.Error),
		OutCh:    make(chan tsslib.Message, parties),
		EndCh:    make(chan keygen.LocalPartySaveData, parties),
		StopChan: make(chan CommunicationError),
	}
}

// tssKeyGen instance initializating
func (t *TssServer) NewTsskeySign(parties, quorum int) {
	t.KeysignInstance = &TssKeySign{
		Logger:            t.Logger,
		LocalPartyID:      t.LocalPartyID,
		OutCh:             make(chan tsslib.Message, parties),
		EndCh:             make(chan *signing.SignatureData, parties),
		ErrCh:             make(chan *tsslib.Error),
		StopChan:          make(chan CommunicationError),
		Pubkey:            t.Pubkey,
		P2pComm:           t.P2p,
		Parties:           parties,
		Quorum:            quorum,
		ConnectionStorage: t.ConnectionStorage,
		KeysignMsgsStorage: &KeysignMsgsStorage{
			RWMutex: new(sync.RWMutex),
			M:       make(map[string]TssMessage),
		},
		Key:           t.Key,
		OneRoundMsgCh: make(chan *P2pMessage, parties),
	}
}

// send tss handshake to peers
func (t *TssServer) SendHandshake(addr string) error {
	handshakeMsg := &P2pMessage{
		Type:     HandshakeMsgType,
		Pubkey:   t.Pubkey,
		PeerAddr: t.P2p.GetRealAddress(),
	}

	data, err := json.Marshal(handshakeMsg)
	if err != nil {
		return fmt.Errorf("marshal : %w", err)
	}
	err = t.P2p.SendMsg(data, []string{addr}, t.P2p.GetRealAddress())
	if err != nil {
		return fmt.Errorf("SendMsg : %w", err)
	}
	return nil
}

// send tss handshake to peers
func (t *TssServer) SendDisconnect(errCh chan error, resultCh chan bool) {
	handshakeMsg := &P2pMessage{
		Type:     DisconnectMsgType,
		Pubkey:   t.Pubkey,
		PeerAddr: t.P2p.GetRealAddress(),
	}

	data, err := json.Marshal(handshakeMsg)
	if err != nil {
		errCh <- fmt.Errorf("marshal : %w", err)
		resultCh <- false
		return
	}

	t.RWMutex.RLock()
	defer t.RWMutex.RUnlock()

	addrs := make([]string, 0, len(t.ConnectionStorage))

	for _, addr := range t.ConnectionStorage {
		addrs = append(addrs, addr)
	}
	err = t.P2p.SendMsg(data, addrs, t.P2p.GetRealAddress())
	if err != nil {
		errCh <- fmt.Errorf("SendMsg : %w", err)
		resultCh <- false
	}
	resultCh <- true
	return
}
