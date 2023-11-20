package tss

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/KiraCore/sekai-bridge/utils"
	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	tsslib "github.com/binance-chain/tss-lib/tss"
	"go.uber.org/zap"
)

const (
	Round1     = "KGRound1Message"
	Round2Msg1 = "KGRound2Message1"
	Round2Msg2 = "KGRound2Message2"
	Round3     = "KGRound3Message"

	Round1Prefix = "KGRound1"
	Round2Prefix = "KGRound2"
	Round3Prefix = "KGRound3"
)

func (t *TssServer) Keygen(parties, threshold int) (*Response, error) {
	// initialize keygen struct
	t.NewTssKeyGen(parties, threshold)

	// notify all nodes to start keygen
	t.KeygenInstance.IsStarted.Store(true)

	err := t.KeygenStartNotify()
	if err != nil {
		return nil, fmt.Errorf("KeygenStartNotify : %w", err)
	}

	partiesID, localPartyID, err := t.GetParties(t.Pubkey)
	if err != nil {
		return nil, fmt.Errorf("GetParties: %w", err)
	}

	// generate key
	key, err := t.KeygenInstance.GenerateNewKey(partiesID, localPartyID)
	if err != nil {
		return nil, fmt.Errorf("GenerateKey : %w", err)
	}

	t.Key = key

	return &Response{
		Key: key,
	}, nil
}

func (t *TssKeyGen) GenerateNewKey(partiesID []*tsslib.PartyID, localPartyID *tsslib.PartyID) (*keygen.LocalPartySaveData, error) {
	t.IsStarted.Store(true)
	defer func() {
		t.IsStarted.Store(false)
	}()
	timeStart := time.Now()

	ctx := tsslib.NewPeerContext(partiesID)
	params := tsslib.NewParameters(ctx, localPartyID, len(partiesID), t.Threshold)

	// @TODO: config value
	preParams, err := keygen.GeneratePreParams(10 * time.Minute)
	if err != nil {
		return nil, fmt.Errorf("GeneratePreParams: %w", err)
	}

	t.PG = keygen.NewLocalParty(params, t.OutCh, t.EndCh, *preParams).(*keygen.LocalParty)

	// start keygen
	go func() {
		if err := t.PG.Start(); nil != err {
			t.Logger.Error("tss -> GenerateNewKey -> Start", zap.Error(err))
			t.ErrCh <- err
		}
	}()

	key, err := t.processKeyGen(t.Parties, timeStart)
	if err != nil {
		return key, fmt.Errorf("processKeyGen: %w", err)
	}
	return key, nil
}

func (t *TssKeyGen) processKeyGen(parties int, timeStart time.Time) (*keygen.LocalPartySaveData, error) {
	defer func() {
		t.KeygenMsgsStorage.Lock()
		t.KeygenMsgsStorage.M = make(map[string]TssMessage)
		t.KeygenMsgsStorage.Unlock()
	}()
	t.Logger.Info("tss -> keygen -> keygen process started")
	ctx, cancel := context.WithCancel(context.Background())
	for {
		select {
		case err := <-t.ErrCh: // when keyGenParty return
			t.Logger.Error("tss -> keygen -> error from errChan", zap.Error(err))
			cancel()
			return nil, errors.New("error channel closed fail to start local party")

		case stopMsg := <-t.StopChan: // when TSS processor receive signal to quit
			t.Logger.Error("keygen -> received stop signal", zap.String("operation", stopMsg.Operation), zap.String("peerAddr", stopMsg.PeerAddr), zap.Time("time", stopMsg.Time))
			cancel()
			return nil, fmt.Errorf("received stop signal from peerAddr = %s, operation = %s,time = %s", stopMsg.PeerAddr, stopMsg.Operation, stopMsg.Time)

		case msg := <-t.OutCh:
			t.Logger.Debug("tss -> keygen -> msg from outCh",
				zap.String("msg", msg.String()),
				zap.String("our partyID", t.LocalPartyID.String()))
			err := t.ProcessOutCh(ctx, msg, parties)
			if err != nil {
				t.Logger.Error("tss -> processKeyGen -> ProcessOutCh", zap.Error(err))
				return nil, err
			}

		case msg := <-t.EndCh:
			t.Logger.Info("tss -> keygen -> key created", zap.String("key", msg.ECDSAPub.Y().String()), zap.Duration("time", time.Since(timeStart)))

			err := utils.SaveKeyFile(&msg)
			if err != nil {
				t.Logger.Error("tss -> processKeyGen -> SaveKeyFile", zap.Error(err))
				return nil, err
			}
			return &msg, nil
		}
	}
}

// notify all connected nodes to initialize keygen
// troubles with time?
func (t *TssServer) KeygenStartNotify() error {
	tssKeygenStartMsg := P2pMessage{
		Type: KeygenStartMsgType,
		Time: time.Now().Unix(), // to prevent filtering this msg
	}

	tssKeygenStartMsgData, err := json.Marshal(tssKeygenStartMsg)
	if err != nil {
		return fmt.Errorf("marshal : %w", err)
	}

	err = t.P2p.SendMsg(tssKeygenStartMsgData, nil, t.P2p.GetRealAddress())
	if err != nil {
		return fmt.Errorf("sendMsg : %w", err)
	}

	return nil
}

// @TODO: channel to stop when key is generated
func (t *TssKeyGen) UpdateForRound(ctx context.Context, tssMsg *TssMessage, parties int) {
	var (
		messagesCounter int // how many messages do we need at this round
		roundPrefix     string
	)

	switch tssMsg.Type {
	case Round1:
		messagesCounter = parties - 1
		roundPrefix = Round1Prefix
	case Round2Msg2:
		messagesCounter = 2 * (parties - 1)
		roundPrefix = Round2Prefix

	case Round3:
		messagesCounter = parties - 1
		roundPrefix = Round3Prefix
	}

	for {
		select {
		case <-ctx.Done():
			t.Logger.Info("keysign -> updateForRound -> got stop signal")
			return
		default:
			time.Sleep(1 * time.Second)
			t.Logger.Info("UpdateForRound -> new iteration", zap.String("type", tssMsg.Type))
			if len(t.KeygenMsgsStorage.M) == 0 {
				t.Logger.Info("UpdateForRound -> map is empty")
				continue
			}

			//range_loop:
			tempMap := make(map[string]TssMessage)
			t.KeygenMsgsStorage.Lock()
			for key, msg := range t.KeygenMsgsStorage.M {
				if strings.Contains(msg.Type, roundPrefix) {
					if msg.IsBroadcast || msg.To[0].Id == t.LocalPartyID.Id {
						tempMap[key] = msg
						t.Logger.Info("ADDED TO TEMP MAP", zap.String("type", msg.Type), zap.String("from", msg.From.GetId()), zap.Bool("broadcast", msg.IsBroadcast))
					}
				}
				continue
			}
			t.KeygenMsgsStorage.Unlock()

			if len(tempMap) != messagesCounter {
				time.Sleep(1 * time.Second)
				t.Logger.Info("TEMPMAP", zap.Int("map length", len(tempMap)), zap.Int("required", messagesCounter))
				//goto range_loop
				continue
			}
			for _, msg := range tempMap {
				err := t.Update(&msg)
				if err != nil {
					t.Logger.Error("tss - > Update", zap.String("type", msg.Type),
						zap.String("from", msg.From.Id),
						zap.Error(err))
					continue
				}
			}
			return
		}
	}
}
