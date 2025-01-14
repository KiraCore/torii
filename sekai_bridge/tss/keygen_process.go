package tss

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	tsslib "github.com/binance-chain/tss-lib/tss"
	"go.uber.org/zap"
)

func (t *TssKeyGen) ProcessOutCh(ctx context.Context, msg tsslib.Message, parties int) error {
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
		Type:   KeygenMsgType,
		TssMsg: &tssMsg,
		Round:  msg.Type(),
		Time:   time.Now().UnixNano(),
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
			return fmt.Errorf("SendMsg : %w", err)
		}
	}
	t.Logger.Info("processOutCh - msg sent",
		zap.String("sender_ID", p2pMsg.TssMsg.From.Id),
		zap.String("type", p2pMsg.TssMsg.Type))

	if msg.Type() != "KGRound2Message1" { // there are 2 messages in round 2
		go t.UpdateForRound(ctx, &tssMsg, parties)
	}
	t.Logger.Info("RETURN?")
	// time.Sleep(3 * time.Second)
	return nil
}

func (t *TssKeyGen) Update(msg *TssMessage) error {
	parsedMsg, err := tsslib.ParseWireMessage(msg.Bytes, msg.From, msg.IsBroadcast)
	if err != nil {
		return fmt.Errorf("ParseWireMessage : %w", err)
	}

	if t.LocalPartyID.Index-1 == parsedMsg.GetFrom().Index {
		return errors.New("tried to send a message to itself")
	}

	if t.PG != nil {
		t.Logger.Info("tss - PG updater started", zap.String("type", msg.Type), zap.String("from", msg.From.Id))
		go t.SharedPartyUpdater(t.PG, parsedMsg, t.ErrCh)
	}
	return nil
}

func (t *TssKeyGen) SharedPartyUpdater(party tsslib.Party, msg tsslib.Message, errCh chan<- *tsslib.Error) {
	// do not send a message from this party back to itself
	if party.PartyID() == msg.GetFrom() {
		return
	}
	bz, _, err := msg.WireBytes()
	if err != nil {
		err := fmt.Errorf("WireBytes : %w", err)
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
