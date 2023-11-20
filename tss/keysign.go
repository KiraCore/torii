package tss

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/binance-chain/tss-lib/common"
	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/binance-chain/tss-lib/ecdsa/signing"
	tsslib "github.com/binance-chain/tss-lib/tss"
	"go.uber.org/zap"
)

var (
	errisNil = "Error is nil"
)

// keysign
func (t *TssServer) Sign(req *SignMessageRequest) (*SignMessageResponse, error) {
	t.NewTsskeySign(t.Parties, t.Quorum)

	if t.Key == nil {
		return nil, fmt.Errorf("signing key was not generated")
	}

	err := t.KeysignStartNotify(req)
	if err != nil {
		return nil, fmt.Errorf("KeysignStartNotify : %w", err)
	}

	partiesID, localPartyID, err := t.GetParties(t.Pubkey)
	if err != nil {
		return nil, fmt.Errorf("GetParties: %w", err)
	}

	signature, err := t.KeysignInstance.SignMessage(req, partiesID, localPartyID, t.Key)
	if err != nil {
		return nil, fmt.Errorf("KeysignInstance.Sign : %w", nil)
	}

	// @TODO : remove testing signature

	verified := t.VerifySignature(signature, req.Msg)

	t.Logger.Info("tss - sign - verify", zap.Bool("verified", verified))

	data, err := json.Marshal(signature)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", nil)
	}

	return &SignMessageResponse{
		SignatureMarshalled: data,
	}, nil
}

func (t *TssKeySign) SignMessage(req *SignMessageRequest, partiesID []*tsslib.PartyID, localPartyID *tsslib.PartyID, key *keygen.LocalPartySaveData) (*common.ECSignature, error) {
	timeStart := time.Now()
	ctx := tsslib.NewPeerContext(partiesID)
	params := tsslib.NewParameters(ctx, localPartyID, len(partiesID), t.Quorum)

	switch req.OneRoundSigning {
	case true:
		t.Logger.Info("tss -> keysign -> one round signing requested")
		t.PS = signing.NewLocalPartyWithOneRoundSign(params, *key, t.OutCh, t.EndCh).(*signing.LocalParty)
	case false:
		convertedMsg := new(big.Int).SetBytes([]byte(req.Msg))
		t.PS = signing.NewLocalParty(convertedMsg, params, *key, t.OutCh, t.EndCh).(*signing.LocalParty)
	}

	// start keygen
	go func() {
		if err := t.PS.Start(); nil != err {
			t.Logger.Error("tss -> SignMessage -> Start", zap.Error(err))
			t.ErrCh <- err
		}
	}()

	sig, err := t.processKeySign(req, timeStart)
	if err != nil {
		return nil, fmt.Errorf("processKeySign : %w", nil)
	}
	return sig, nil
}

func (t *TssKeySign) processKeySign(req *SignMessageRequest, timeStart time.Time) (*common.ECSignature, error) {
	defer func() {
		t.KeysignMsgsStorage.Lock()
		t.KeysignMsgsStorage.M = make(map[string]TssMessage)
		t.KeysignMsgsStorage.Unlock()
	}()
	t.Logger.Info("tss -> keysign -> keysign process started")
	ctx, cancel := context.WithCancel(context.Background())
	for {
		select {
		case err := <-t.ErrCh:
			t.Logger.Error("tss -> keysign -> error from errChan", zap.Error(err))
			cancel()
			return nil, errors.New("error channel closed fail to start local party")

		case stopMsg := <-t.StopChan:
			t.Logger.Error("keysign -> received stop signal", zap.String("operation", stopMsg.Operation), zap.String("peerAddr", stopMsg.PeerAddr), zap.Time("time", stopMsg.Time))
			cancel()
			return nil, fmt.Errorf("received stop signal from peerAddr = %s, operation = %s,time = %s", stopMsg.PeerAddr, stopMsg.Operation, stopMsg.Time)

		case msg := <-t.OutCh:
			t.Logger.Debug("tss -> keysign -> msg from outCh",
				zap.String("msg", msg.String()),
				zap.String("our partyID", t.LocalPartyID.String()))
			err := t.ProcessOutCh(ctx, msg, t.Parties)
			if err != nil {
				t.Logger.Error("tss -> processKeyGen -> ProcessOutCh", zap.Error(err))
				return nil, err
			}

		case msg := <-t.EndCh:
			if !req.OneRoundSigning {
				t.Logger.Info("tss -> keysign -> signature created", zap.String("signature", msg.String()), zap.Duration("time", time.Since(timeStart)))
				return msg.Signature, nil
			} else {
				resp, err := t.HandleOneRoundSigning(msg, req)
				if err != nil {
					t.Logger.Error("tss -> processKeyGen -> HandleOneRoundSigning", zap.Error(err))
					return nil, err
				}
				t.Logger.Info("tss -> keysign -> one round signature created", zap.String("signature", resp.String()), zap.Duration("time", time.Since(timeStart)))

				return resp.Signature, nil
			}
		}
	}
}

// notify all connected nodes to initialize keygen
// troubles with time?
func (t *TssServer) KeysignStartNotify(request *SignMessageRequest) error {
	tssKeysignStartMsg := P2pMessage{
		Type:           KeysignStartMsgType,
		KeysignRequest: request,
	}

	tssKeysignStartMsgData, err := json.Marshal(tssKeysignStartMsg)
	if err != nil {
		return fmt.Errorf("marshal : %w", err)
	}

	err = t.P2p.SendMsg(tssKeysignStartMsgData, nil, t.P2p.GetRealAddress())
	if err != nil {
		return fmt.Errorf("sendMsg : %w", err)
	}

	return nil
}

// handle one round signing
func (t *TssKeySign) HandleOneRoundSigning(state *signing.SignatureData, req *SignMessageRequest) (*signing.SignatureData, error) {
	convertedMsg := new(big.Int).SetBytes([]byte(req.Msg))
	sI := signing.FinalizeGetOurSigShare(state, convertedMsg)

	msg := P2pMessage{
		Type:    KeysignOneRoundMsgType,
		Si:      sI,
		PartyID: t.LocalPartyID,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal : %w", err)
	}

	err = t.P2pComm.SendMsg(data, nil, t.P2pComm.GetRealAddress())
	if err != nil {
		return nil, fmt.Errorf("sendMsg : %w", err)
	}
	otherSiMap := make(map[*tsslib.PartyID]*big.Int)
loop:
	for {
		siMsg := <-t.OneRoundMsgCh
		t.Logger.Info("tss -> keysign -> one round -> got msg from another peer", zap.String("partyID", siMsg.PartyID.Id))
		otherSiMap[siMsg.PartyID] = siMsg.Si
		if len(otherSiMap) == t.Parties-1 {
			t.Logger.Info("tss -> keysign -> one round -> break loop")
			break loop
		}
	}

	signData, signature, err := signing.FinalizeGetAndVerifyFinalSig(state, t.Key.ECDSAPub.ToECDSAPubKey(),
		convertedMsg, t.LocalPartyID, sI, otherSiMap)
	if strings.Contains(err.Error(), errisNil) { // @TODO: error is nil always here
		t.Logger.Info("tss -> keysign -> one round", zap.Any("btcec.Signature", signature))
		return signData, nil
	}
	return nil, fmt.Errorf("FinalizeGetAndVerifyFinalSig : %w", err)
}
