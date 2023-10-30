package tss

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/KiraCore/sekai-bridge/tss/signmsg"
	"github.com/KiraCore/sekai-bridge/types"
	tsslib "github.com/binance-chain/tss-lib/tss"

	p2p "github.com/saiset-co/saiP2P-go/core"
	"go.uber.org/zap"
)

// tssServer instance initializating
func New(pubkey string, p2p *p2p.Core, l *zap.Logger) *TssServer {
	partyID := PubkeyToPartyID(pubkey)

	return &TssServer{
		ConnectionStorage: make(map[string]string),
		Logger:            l,
		P2p:               p2p,
		RWMutex:           new(sync.RWMutex),
		LocalPartyID:      partyID,
		Pubkey:            pubkey,
		//t.EndChS = make(chan *signing.SignatureData, 1)
	}

}

// tssKeyGen instance initializating
func (t *TssServer) NewTssKeyGen() *TssKeyGen {
	return &TssKeyGen{
		Logger:                      t.Logger,
		LocalPartyID:                t.LocalPartyID,
		StopChan:                    make(chan struct{}),
		CommStopChan:                make(chan struct{}),
		OutCh:                       make(chan tsslib.Message),
		CachedWireBroadcastMsgLists: new(sync.Map),
		CachedWireUnicastMsgLists:   new(sync.Map),
		Pubkey:                      t.Pubkey,
		ConnectionStorage:           t.ConnectionStorage,
		PartiesMap:                  map[tsslib.PartyID]bool{},
		P2pComm:                     t.P2p,
	}
}

func (t *TssServer) SendHandshake(addr string) error {
	handshakeMsg := &P2pMessage{
		Type:     HandshakeMsgType,
		Pubkey:   t.Pubkey,
		PeerAddr: t.P2p.GetRealAddress(),
	}

	data, err := json.Marshal(handshakeMsg)
	if err != nil {
		return fmt.Errorf("Marshal : %w", err)
	}
	err = t.P2p.SendMsg(data, []string{addr}, t.P2p.GetRealAddress())
	if err != nil {
		return fmt.Errorf("SendMsg : %w", err)
	}
	return nil
}

// handle incoming p2p message
func (t *TssServer) HandleP2Pmessage(p2pMsg *p2p.Message) error {

	msg := P2pMessage{}

	err := json.Unmarshal(p2pMsg.Data, &msg)
	if err != nil {
		return fmt.Errorf("Marshal :%w", err)
	}

	t.Logger.Info("service -> HandleP2Pmessage - got msg", zap.String("from", p2pMsg.From),
		zap.Strings("to", p2pMsg.To),
		zap.String("type", msg.Type),
		zap.Any("msg", msg))

	switch msg.Type {
	case HandshakeMsgType: // for adding peers to map[peerAddr]PartyID
		err := t.HandleHandshake(&msg)
		if err != nil {
			return fmt.Errorf("HandleHandshake : %w", err)
		}
	case KeygenStartMsgType: // start keygen
		if t.TssKeygen.IsStarted {
			t.Logger.Debug("tss -> HandleP2PMessage -> keygen already started")
			return nil
		}
		t.Logger.Info("tss -> HandleP2PMessage -> keygen start command got", zap.Any("request", msg.KeygenRequest))
		err := t.TssKeygen.GenerateNewKey(&msg.KeygenRequest)
		if err != nil {
			return fmt.Errorf("GenerateNewKey : %w", err)
		}

	case KeygenMsgType: // for exchanging tss messages in keygen stage
		if !t.KeygenStarted { //check if keygen was not started in this node
			t.Logger.Info("tss -> HandleP2PMessage -> keygen start command got", zap.Any("request", msg.KeygenRequest))
			return errors.New("keygen was not started")
		}

		err = t.TssKeygen.ProcessOutCh(msg.TssMessage)
		if err != nil {
			return fmt.Errorf("Update : %w", err)
		}

		return nil
	}
	return nil
}

func (t *TssServer) Sign(data types.SignMessageRequest) signmsg.Response {
	signMsgInstance := signmsg.NewTssSignMsg()
	return signMsgInstance.SignMsg()
}

func (t *TssServer) Verify(data interface{}) bool {
	return true
}
