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
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

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

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Logger.Info("keysign -> updateForRound -> got stop signal")
			return
		case <-ticker.C:
			//time.Sleep(1 * time.Second)
			t.Logger.Info("UpdateForRound -> new iteration", zap.String("type", tssMsg.Type))
			if len(t.KeysignMsgsStorage.M) == 0 {
				t.Logger.Info("UpdateForRound -> map is empty")
				continue
			}

			//range_loop:
			t.KeysignMsgsStorage.Lock()

			allMsgDetails := make([]map[string]interface{}, 0)
			for _, msg := range t.KeysignMsgsStorage.M {
				allMsgDetails = append(allMsgDetails, map[string]interface{}{
					"type":      msg.Type,
					"from":      msg.From.Id,
					"broadcast": msg.IsBroadcast,
					"to_count":  len(msg.To),
				})
			}
			t.Logger.Info("UpdateForRound -> ALL messages in storage",
				zap.String("waitingFor", tssMsg.Type),
				zap.Int("total", len(t.KeysignMsgsStorage.M)),
				zap.Any("all_messages", allMsgDetails))

			tempMap := make(map[string]TssMessage)
			filteredOut := make([]map[string]interface{}, 0)
			for key, msg := range t.KeysignMsgsStorage.M {
				var typeMatches bool
				if singleMsg {
					typeMatches = msg.Type == tssMsg.Type
				} else {
					typeMatches = strings.Contains(msg.Type, msgType)
				}

				if typeMatches {
					if msg.IsBroadcast || (len(msg.To) > 0 && msg.To[0].Id == t.LocalPartyID.Id) {
						tempMap[key] = msg
					} else {
						toIds := make([]string, 0)
						if len(msg.To) > 0 {
							for _, to := range msg.To {
								toIds = append(toIds, to.Id)
							}
						}
						filteredOut = append(filteredOut, map[string]interface{}{
							"reason":         "not broadcast and not for local party",
							"type":           msg.Type,
							"from":           msg.From.Id,
							"broadcast":      msg.IsBroadcast,
							"to_ids":         toIds,
							"local_party_id": t.LocalPartyID.Id,
							"to_count":       len(msg.To),
						})
					}
				} else {
					toIds := make([]string, 0)
					if len(msg.To) > 0 {
						for _, to := range msg.To {
							toIds = append(toIds, to.Id)
						}
					}
					var reason string
					if singleMsg {
						reason = "wrong type (exact match)"
					} else {
						reason = "wrong type (prefix match)"
					}
					filteredOut = append(filteredOut, map[string]interface{}{
						"reason":         reason,
						"type":           msg.Type,
						"from":           msg.From.Id,
						"to_ids":         toIds,
						"local_party_id": t.LocalPartyID.Id,
						"waiting_for":    tssMsg.Type,
						"msgType_prefix": msgType,
					})
				}
			}

			if len(filteredOut) > 0 {
				t.Logger.Info("UpdateForRound -> FILTERED OUT messages",
					zap.String("waitingFor", tssMsg.Type),
					zap.Int("filtered_count", len(filteredOut)),
					zap.Any("filtered_messages", filteredOut))
			}
			t.KeysignMsgsStorage.Unlock()

			msgDetails := make([]map[string]interface{}, 0)
			for _, msg := range tempMap {
				msgDetails = append(msgDetails, map[string]interface{}{
					"type":      msg.Type,
					"from":      msg.From.Id,
					"broadcast": msg.IsBroadcast,
				})
			}

			t.Logger.Info("UpdateForRound -> messages status",
				zap.String("waitingFor", tssMsg.Type),
				zap.Int("got", len(tempMap)),
				zap.Int("required", messagesCounter),
				zap.Any("messages", msgDetails))

			if len(tempMap) != messagesCounter {
				t.Logger.Info("UpdateForRound -> NOT ENOUGH MESSAGES",
					zap.String("type", tssMsg.Type),
					zap.Int("have", len(tempMap)),
					zap.Int("need", messagesCounter),
					zap.Any("collected", msgDetails))
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
	if msg == nil {
		return fmt.Errorf("message is nil")
	}

	if msg.From == nil {
		return fmt.Errorf("message.From is nil")
	}

	if msg.Bytes == nil {
		return fmt.Errorf("message.Bytes is nil")
	}

	parsedMsg, err := tsslib.ParseWireMessage(msg.Bytes, msg.From, msg.IsBroadcast)
	if err != nil {
		return fmt.Errorf("ParseWireMessage : %w", err)
	}

	if parsedMsg == nil {
		return fmt.Errorf("parsed message is nil")
	}

	if parsedMsg.GetFrom() == nil {
		return fmt.Errorf("parsed message sender is nil")
	}

	if t.LocalPartyID == nil {
		return fmt.Errorf("local party ID is nil")
	}

	if t.LocalPartyID.Index == parsedMsg.GetFrom().Index {
		return errors.New("tried to send a message to itself")
	}

	if t.PS != nil {
		t.Logger.Info("tss - PS updater started", zap.String("type", msg.Type), zap.String("from", msg.From.Id))
		go t.SharedPartyUpdater(t.PS, parsedMsg, t.ErrCh)
	} else {
		return fmt.Errorf("party is nil")
	}

	return nil
}

func (t *TssKeySign) SharedPartyUpdater(party tsslib.Party, msg tsslib.Message, errCh chan<- *tsslib.Error) {
	// time.Sleep(1 * time.Second)

	if party == nil {
		t.Logger.Error("tss -> SharedPartyUpdater -> party is nil")
		if errCh != nil {
			errCh <- party.WrapError(fmt.Errorf("party is nil"))
		}
		return
	}

	// Проверяем, что msg не nil и все необходимые методы возвращают не nil
	if msg == nil {
		t.Logger.Error("tss -> SharedPartyUpdater -> msg is nil")
		if errCh != nil {
			errCh <- party.WrapError(fmt.Errorf("message is nil"))
		}
		return
	}

	// do not send a message from this party back to itself
	if party.PartyID() == msg.GetFrom() {
		return
	}

	if msg.GetFrom() == nil {
		t.Logger.Error("tss -> SharedPartyUpdater -> msg.GetFrom() is nil")
		if errCh != nil {
			errCh <- party.WrapError(fmt.Errorf("message sender is nil"))
		}
		return
	}

	bz, _, err := msg.WireBytes()
	if err != nil {
		err := fmt.Errorf(" wireBytes : %w", err)
		errCh <- party.WrapError(err)
		return
	}

	t.psLock.Lock()
	if _, err := party.UpdateFromBytes(bz, msg.GetFrom(), msg.IsBroadcast()); err != nil {
		err := fmt.Errorf("UpdateFromBytes err =  %w, type = %s, from = %s", err, msg.Type(), msg.GetFrom().Id)
		errCh <- party.WrapError(err)
		return
	}
	t.psLock.Unlock()

	t.Logger.Info("process - SharedPartyUpdater - Success", zap.String("type", msg.Type()),
		zap.String("from", msg.GetFrom().Id))
}
