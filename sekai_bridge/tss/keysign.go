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

	if partiesID == nil || len(partiesID) == 0 {
		t.Logger.Error("tss -> HandleP2PMessage -> KeysignStartMsgType -> partiesID is nil or empty")
		return nil, fmt.Errorf("GetParties: %w", err)
	}

	if localPartyID == nil {
		t.Logger.Error("tss -> HandleP2PMessage -> KeysignStartMsgType -> localPartyID is nil")
		return nil, fmt.Errorf("GetParties: %w", err)
	}

	if t.Key.ECDSAPub == nil {
		return nil, fmt.Errorf("ECDSAPub in signing key is nil")
	}

	signature, err := t.KeysignInstance.SignMessage(req, partiesID, localPartyID, t.Key)
	if err != nil {
		return nil, fmt.Errorf("KeysignInstance.Sign : %w", err)
	}

	// @TODO : remove testing signature

	verified := t.VerifySignature(signature, req.Tx)

	t.Logger.Info("tss - sign - verify", zap.Bool("verified", verified))

	data, err := json.Marshal(signature)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
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
		convertedMsg := new(big.Int).SetBytes([]byte(req.Tx))
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
		return nil, fmt.Errorf("processKeySign : %w", err)
	}
	return sig, nil
}

func (t *TssKeySign) ClearKeySignMessagesStorage() {
	t.KeysignMsgsStorage.Lock()
	t.KeysignMsgsStorage.M = make(map[string]TssMessage)
	t.KeysignMsgsStorage.Unlock()
}

func (t *TssKeySign) processKeySign(req *SignMessageRequest, timeStart time.Time) (*common.ECSignature, error) {
	defer t.ClearKeySignMessagesStorage()

	t.Logger.Info("tss -> keysign -> keysign process started")
	timeout := 5 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

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
	convertedMsg := new(big.Int).SetBytes([]byte(req.Tx))
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
	receivedCount := 0
	timeout := time.After(30 * time.Second)

loop:
	for {
		select {
		case siMsg := <-t.OneRoundMsgCh:
			// Проверим, что siMsg и его поля не равны nil
			if siMsg == nil || siMsg.PartyID == nil || siMsg.Si == nil {
				t.Logger.Error("tss -> keysign -> one round -> received invalid message",
					zap.Any("siMsg", siMsg))
				continue
			}

			t.Logger.Info("tss -> keysign -> one round -> got msg from another peer",
				zap.String("partyID", siMsg.PartyID.Id))

			// Проверим, что мы не добавляем дубликаты
			if _, exists := otherSiMap[siMsg.PartyID]; !exists {
				otherSiMap[siMsg.PartyID] = siMsg.Si
				receivedCount++
			}

			if receivedCount >= t.Parties-1 {
				t.Logger.Info("tss -> keysign -> one round -> break loop")
				break loop
			}

		case <-timeout:
			// Выход по таймауту, если не получили все необходимые сообщения
			t.Logger.Error("tss -> keysign -> one round -> timeout waiting for messages",
				zap.Int("expected", t.Parties-1),
				zap.Int("received", receivedCount))
			return nil, fmt.Errorf("timeout waiting for all parties to send Si values")
		}
	}

	if state == nil {
		return nil, fmt.Errorf("signature state is nil")
	}

	if t.Key == nil || t.Key.ECDSAPub == nil {
		return nil, fmt.Errorf("key or ECDSAPub is nil")
	}

	pubKey := t.Key.ECDSAPub.ToECDSAPubKey()
	if pubKey == nil {
		return nil, fmt.Errorf("failed to convert ECDSAPub to ECDSA public key")
	}

	for partyID, si := range otherSiMap {
		if partyID == nil {
			return nil, fmt.Errorf("partyID in otherSiMap is nil")
		}
		if si == nil {
			return nil, fmt.Errorf("si value for partyID %s is nil", partyID.Id)
		}
	}

	signData, signature, err := signing.FinalizeGetAndVerifyFinalSig(state, t.Key.ECDSAPub.ToECDSAPubKey(),
		convertedMsg, t.LocalPartyID, sI, otherSiMap)
	if err == nil {
		t.Logger.Info("tss -> keysign -> one round", zap.Any("btcec.Signature", signature))
		// Проверим, что signData не нулевой
		if signData == nil {
			return nil, fmt.Errorf("signature data is nil although no error was returned")
		}
		return signData, nil
	} else if strings.Contains(err.Error(), errisNil) {
		t.Logger.Info("tss -> keysign -> one round (with errisNil)", zap.Any("btcec.Signature", signature))
		// Проверим, что signData не нулевой
		if signData == nil {
			return nil, fmt.Errorf("signature data is nil with errisNil: %v", err)
		}
		return signData, nil
	}
	return nil, fmt.Errorf("FinalizeGetAndVerifyFinalSig : %w", err)
}
