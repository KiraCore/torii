package tss

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/KiraCore/sekai-bridge/types"
	bcrypto "github.com/binance-chain/tss-lib/crypto"
	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	tsslib "github.com/binance-chain/tss-lib/tss"
	p2p "github.com/saiset-co/saiP2P-go/core"
	"go.uber.org/zap"
)

func (t *TssServer) Keygen(req *types.GenerateKeysRequest) (*Response, error) {
	// notify all nodes to start keygen
	err := req.Validate()
	if err != nil {
		return nil, fmt.Errorf(" Validate request : %w", err)
	}

	err = t.KeygenStartNotify(req.Parties, req.Threshold)
	if err != nil {
		return nil, fmt.Errorf("KeygenStartNotify", err)
	}

	err = t.TssKeygen.GenerateNewKey(req)
	if err != nil {
		return nil, fmt.Errorf("GenerateKey : %w", err)
	}

	return &Response{
		Key: t.TssKeygen.Key,
	}, nil

}

func (t *TssKeyGen) GenerateNewKey(req *types.GenerateKeysRequest) error {
	keys := make([]string, 0)
	for pubkey := range t.ConnectionStorage {
		keys = append(keys, pubkey)
	}

	keys = append(keys, t.Pubkey)

	partiesID, localPartyID, err := t.GetParties(keys, t.Pubkey)
	if err != nil {
		return fmt.Errorf("GetParties: %w", err)
	}

	//t.Logger.Debug("tss -> GenerateNewKey", zap.Any("localPartyID", localPartyID), zap.Any("partiesID", partiesID))

	ctx := tsslib.NewPeerContext(partiesID)
	params := tsslib.NewParameters(ctx, localPartyID, len(partiesID), req.Threshold)
	outCh := make(chan tsslib.Message, len(partiesID))
	endCh := make(chan keygen.LocalPartySaveData, len(partiesID))
	errChan := make(chan struct{})

	preParams, err := keygen.GeneratePreParams(1 * time.Minute)
	if err != nil {
		return fmt.Errorf("GeneratePreParams: %w", err)
	}

	keyGenLocalParty := keygen.NewLocalParty(params, outCh, endCh, *preParams)

	// start keygen
	go func() {
		if err := keyGenLocalParty.Start(); nil != err {
			t.Logger.Error("tss -> GenerateNewKey -> Start", zap.Error(err))
			close(errChan)
		}
	}()

	t.IsStarted = true

	go t.processKeyGen(errChan, outCh, endCh)
loop:
	select {
	case <-errChan: // when keyGenParty return
		t.Logger.Error("tss -> keygen -> error from errChan")
		return errors.New("error channel closed fail to start local party")

	case <-t.StopChan: // when TSS processor receive signal to quit
		return errors.New("received exit signal")

	case msg := <-endCh:
		t.Logger.Info("tss -> keygen -> key created", zap.String("key", msg.ECDSAPub.Y().String()))
		t.Key = &msg
		break loop
	}

	return nil

}

func (tKeyGen *TssKeyGen) processKeyGen(errChan chan struct{},
	outCh <-chan tsslib.Message,
	endCh <-chan keygen.LocalPartySaveData) (*bcrypto.ECPoint, error) {
	defer tKeyGen.Logger.Debug("tss -> keygen -> finished keygen process")
	tKeyGen.Logger.Debug("tss -> keygen -> keygen process started")
	for {
		select {
		case msg := <-outCh:
			tKeyGen.Logger.Debug("tss -> keygen -> msg from outCh", zap.String("from", msg.GetFrom().String()), zap.String("our partyID", tKeyGen.LocalPartyID.String()))
			if msg.GetFrom().Moniker == tKeyGen.LocalPartyID.Moniker {
				tKeyGen.Logger.Error("tss -> keygen -> got msg from ourselves")
				continue
			}
			err := tKeyGen.ProcessOutCh(msg)
			if err != nil {
				tKeyGen.Logger.Error("tss -> processKeyGen -> ProcessOutCh", zap.Error(err))
				return nil, err
			}
		}
	}
}

// notify all connected nodes to initialize keygen
// troubles with time?
func (t *TssServer) KeygenStartNotify(parties, threshold int) error {
	tssKeygenStartMsg := P2pMessage{
		Type: KeygenStartMsgType,
		KeygenRequest: types.GenerateKeysRequest{
			Parties:   parties,
			Threshold: threshold,
		},
	}

	tssKeygenStartMsgData, err := json.Marshal(tssKeygenStartMsg)
	if err != nil {
		return fmt.Errorf("Marshal : %w", err)
	}

	err = t.P2p.SendMsg(tssKeygenStartMsgData, nil, t.P2p.GetRealAddress())
	if err != nil {
		return fmt.Errorf("SendMsg : %w", err)
	}

	return nil
}

func (t *TssKeyGen) ProcessOutCh(msg tsslib.Message) error {
	_, r, err := msg.WireBytes()
	// if we cannot get the wire share, the tss will fail, we just quit.
	if err != nil {
		return fmt.Errorf("WireBytes : %w", err)
	}

	p2pMsg := P2pMessage{
		Type:        KeygenMsgType,
		TssMessage:  msg,
		KeygenRound: msg.Type(),
	}

	data, err := json.Marshal(p2pMsg)
	if err != nil {
		return fmt.Errorf("Marshal : %w", err)
	}

	// data, err := msgpack.Marshal(p2pMsg)
	// if err != nil {
	// 	return fmt.Errorf("Marshal : %w", err)
	// }

	m := p2p.Message{
		From: t.P2pComm.GetRealAddress(),
		Data: data,
	}

	msgData, err := json.Marshal(m)
	if err != nil {
		fmt.Println("Marshal error:", err)
	}

	// msgData, err := msgpack.Marshal(m)
	// if err != nil {
	// 	fmt.Println("Marshal error:", err)
	// }

	t.Logger.Info("tss -> keygen - processOutCh", zap.Int("data size", len(msgData)), zap.Int("tssMsg size", len(data)))
	if r.IsBroadcast { // send to all
		err = t.P2pComm.SendMsg(msgData, nil, t.P2pComm.GetRealAddress())
		if err != nil {
			fmt.Println("SendMsg:", err)
		}
	} else { // send to specified peer
		addrs := t.GetPartyIDpeerAddr()
		err = t.P2pComm.SendMsg(msgData, addrs, t.P2pComm.GetRealAddress())
		if err != nil {
			fmt.Println("SendMsg:", err)
		}
	}

	return nil
}
