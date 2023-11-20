package tss

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	tsslib "github.com/binance-chain/tss-lib/tss"
	"go.uber.org/zap"
)

const (
	KeysignRound1Msg1 = "SignRound1Message1"
	KeysignRound1Msg2 = "SignRound1Message2"
	KeysignRound2     = "SignRound2Message"
	KeysignRound3     = "SignRound3Message"
	KeysignRound4     = "SignRound4Message"
	KeysignRound5     = "SignRound5Message"
	KeysignRound6     = "SignRound6Message"
	KeysignRound7     = "SignRound7Message"

	KSRound1Prefix = "SignRound1Message"
)

func (t *TssKeySign) ProcessOutCh(ctx context.Context, msg tsslib.Message, parties int) error {
	b, r, err := msg.WireBytes()
	if err != nil {
		return fmt.Errorf("WireBytes : %w", err)
	}

	tssMsg := TssMessage{
		From:        msg.GetFrom(),
		To:          msg.GetTo(),
		IsBroadcast: msg.IsBroadcast(),
		Bytes:       b,
		Type:        msg.Type(),
		Routing:     r,
	}

	p2pMsg := P2pMessage{
		Type:   KeysignMsgType,
		TssMsg: &tssMsg,
		Round:  msg.Type(),
		Time:   time.Now().Unix(),
	}

	data, err := json.Marshal(p2pMsg)
	if err != nil {
		return fmt.Errorf("marshal : %w", err)
	}
	// time.Sleep(500 * time.Millisecond)
	if msg.IsBroadcast() { // send to all
		err = t.P2pComm.SendMsg(data, nil, t.P2pComm.GetRealAddress())
		if err != nil {
			return fmt.Errorf("SendMsg : %w", err)
		}
	} else { // send to specified peer
		addrs := GetPeersAddresses(t.ConnectionStorage)
		err = t.P2pComm.SendMsg(data, addrs, t.P2pComm.GetRealAddress())
		if err != nil {
			return fmt.Errorf("sendMsg : %w", err)
		}
	}
	t.Logger.Info("processOutCh - msg sent",
		zap.String("sender_ID", p2pMsg.TssMsg.From.Id),
		zap.String("type", p2pMsg.TssMsg.Type),
		zap.Any("to", p2pMsg.TssMsg.To),
		zap.Strings("addrs", GetPeersAddresses(t.ConnectionStorage)))

	if msg.Type() != "SignRound1Message1" { // there are 2 messages in round 1
		go t.UpdateForRound(ctx, &tssMsg, parties)
	}
	t.Logger.Info("RETURN?")
	// time.Sleep(2 * time.Second)
	return nil
}

func (t *TssKeySign) UpdateForRound(ctx context.Context, tssMsg *TssMessage, parties int) {
	var (
		messagesCounter int // how many messages do we need at this round
		msgType         string
		singleMsg       bool
	)

	// @TODO: make it simplier
	switch tssMsg.Type {
	case KeysignRound1Msg2:
		messagesCounter = 2 * (parties - 1)
		msgType = KSRound1Prefix
		singleMsg = false
	default:
		messagesCounter = parties - 1
		msgType = tssMsg.Type
		singleMsg = true
	}

	for {
		select {
		case <-ctx.Done():
			t.Logger.Info("keysign -> updateForRound -> got stop signal")
			return
		default:
			time.Sleep(1 * time.Second)
			t.Logger.Info("UpdateForRound -> new iteration", zap.String("type", tssMsg.Type))
			if len(t.KeysignMsgsStorage.M) == 0 {
				t.Logger.Info("UpdateForRound -> map is empty")
				continue
			}

			//range_loop:
			tempMap := make(map[string]TssMessage)
			t.KeysignMsgsStorage.Lock()
			for key, msg := range t.KeysignMsgsStorage.M {
				if singleMsg {
					if msg.Type == tssMsg.Type {
						if msg.IsBroadcast || msg.To[0].Id == t.LocalPartyID.Id {
							tempMap[key] = msg
							//	t.Logger.Info("ADDED TO TEMP MAP", zap.String("type", msg.Type), zap.String("from", msg.From.GetId()), zap.Bool("broadcast", msg.IsBroadcast))
						}
					}
					continue
				}
				if strings.Contains(msg.Type, msgType) {
					if msg.IsBroadcast || msg.To[0].Id == t.LocalPartyID.Id {
						tempMap[key] = msg
						// t.Logger.Info("ADDED TO TEMP MAP", zap.String("type", msg.Type), zap.String("from", msg.From.GetId()), zap.Bool("broadcast", msg.IsBroadcast))
					}
				}
				continue
			}
			t.KeysignMsgsStorage.Unlock()

			if len(tempMap) != messagesCounter {
				//	time.Sleep(1 * time.Second)
				//		t.Logger.Info("TEMPMAP", zap.Int("map length", len(tempMap)), zap.Int("required", messagesCounter))
				//goto range_loop
				continue
			}
			msgSlice := make([]TssMessage, 0, len(tempMap))
			for _, msg := range tempMap {
				// trying to sort slice to update
				msgSlice = append(msgSlice, msg)
			}

			sort.SliceStable(msgSlice, func(i, j int) bool {
				return msgSlice[i].From.Id < msgSlice[j].From.Id
			})

			for _, m := range msgSlice {
				// time.Sleep(1 * time.Second)
				err := t.Update(&m)
				if err != nil {
					t.Logger.Error("tss - > Update", zap.String("type", m.Type),
						zap.String("from", m.From.Id),
						zap.Error(err))
					continue
				}
			}
			return
		}
	}
}

func (t *TssKeySign) Update(msg *TssMessage) error {
	parsedMsg, err := tsslib.ParseWireMessage(msg.Bytes, msg.From, msg.IsBroadcast)
	if err != nil {
		return fmt.Errorf("ParseWireMessage : %w", err)
	}

	if t.LocalPartyID.Index-1 == parsedMsg.GetFrom().Index {
		return errors.New("tried to send a message to itself")
	}

	if t.PS != nil {
		t.Logger.Info("tss - PS updater started", zap.String("type", msg.Type), zap.String("from", msg.From.Id))
		go t.SharedPartyUpdater(t.PS, parsedMsg, t.ErrCh)
	}
	return nil
}

func (t *TssKeySign) SharedPartyUpdater(party tsslib.Party, msg tsslib.Message, errCh chan<- *tsslib.Error) {
	// time.Sleep(1 * time.Second)

	// do not send a message from this party back to itself
	if party.PartyID() == msg.GetFrom() {
		return
	}
	bz, _, err := msg.WireBytes()
	if err != nil {
		err := fmt.Errorf(" wireBytes : %w", err)
		errCh <- party.WrapError(err)
		return
	}

	if _, err := party.UpdateFromBytes(bz, msg.GetFrom(), msg.IsBroadcast()); err != nil {
		err := fmt.Errorf("UpdateFromBytes err =  %w, type = %s, from = %s", err, msg.Type(), msg.GetFrom().Id)
		errCh <- party.WrapError(err)
		return
	}

	t.Logger.Info("process - SharedPartyUpdater - Success", zap.String("type", msg.Type()),
		zap.String("from", msg.GetFrom().Id))
}
