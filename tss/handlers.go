package tss

import (
	"encoding/json"
	"errors"
	"fmt"

	p2p "github.com/saiset-co/saiP2P-go/core"
	"go.uber.org/zap"
)

// handle incoming p2p message
func (t *TssServer) HandleP2Pmessage(p2pMsg *p2p.Message) {
	msg := P2pMessage{}

	err := json.Unmarshal(p2pMsg.Data, &msg)
	if err != nil {
		t.Logger.Error("tss -> HandleP2Pmessage -> Unmarshal", zap.Error(err)) //, zap.Any("msg", p2pMsg))

		// check if this msg was already handled
		commErr, err := t.HandleUnmarshalError(p2pMsg)
		if err != nil {
			t.Logger.Error("tss -> HandleP2Pmessage -> HandleUnmarshalError", zap.Error(err))
			return
		}

		switch commErr.Operation {
		case KeygenOperation:
			if t.KeygenInstance.IsStarted.Load() == true {
				t.KeygenInstance.IsStarted.Store(false)
				t.KeygenInstance.StopChan <- *commErr
			} else {
				t.Logger.Info("service -> HandleP2Pmessage - error -> keygen error already handled")
			}

		case KeysignOperation:
			if t.KeysignInstance.IsStarted.Load() == true {
				t.KeysignInstance.IsStarted.Store(false)
				t.KeysignInstance.StopChan <- *commErr
			} else {
				t.Logger.Info("service -> HandleP2Pmessage - error -> keysign already handled")
			}
		}
		err = t.NotifyAboutError(commErr)
		if err != nil {
			t.Logger.Error("tss -> HandleP2Pmessage -> NotifyAboutError")
			return
		}

		return
	}

	t.Logger.Info("service -> HandleP2Pmessage - got msg", zap.String("from", p2pMsg.From),
		zap.Strings("to", p2pMsg.To), zap.String("type", msg.Type))

	switch msg.Type {
	case HandshakeMsgType: // for adding peers to map[id]peerAddr
		err := t.HandleHandshake(&msg)
		if err != nil {
			t.Logger.Error("tss -> HandleP2Pmessage -> HandleHandshake", zap.Error(err))
			return
		}
	case DisconnectMsgType: // for remove peer from map[id]peerAddr
		err := t.HandleDisconnect(&msg)
		if err != nil {
			t.Logger.Error("tss -> HandleP2Pmessage -> HandleDisconnect", zap.Error(err))
			return
		}
	// keygen
	case KeygenStartMsgType: // start keygen command
		if t.KeygenInstance == nil {
			t.NewTssKeyGen(t.Parties, t.Threshold)
		}
		if t.KeygenInstance.IsStarted.Load() == true {
			t.Logger.Debug("tss -> HandleP2PMessage -> keygen already started")
			return
		}

		t.Logger.Info("tss -> HandleP2PMessage -> keygen start", zap.Int("parties", t.Parties),
			zap.Int("threshold", t.Threshold))

		partiesID, localPartyID, err := t.GetParties(t.Pubkey)
		if err != nil {
			t.Logger.Error("tss -> HandleP2PMessage -> KeygenStartMsgType -> GetParties")
			return
		}

		key, err := t.KeygenInstance.GenerateNewKey(partiesID, localPartyID)
		if err != nil {
			t.Logger.Error("tss -> HandleP2Pmessage -> GenerateNewKey", zap.Error(err))
			return
		}
		t.Key = key

	case KeygenMsgType: // for exchanging tss messages in keygen stage
		t.Logger.Info("service -> HandleP2Pmessage -> keygen ->  got msg", zap.String("from", p2pMsg.From),
			zap.Strings("to", p2pMsg.To), zap.String("type", msg.Type), zap.String("round", msg.Round))
		// @TODO: use not broadcasted msgs?
		if msg.TssMsg.From.Id == t.Pubkey {
			t.Logger.Error("tss -> handlers -> KeygenMsgType", zap.String("in id", msg.TssMsg.From.Id), zap.Error(errors.New("msg from own ID")))
			return
		}

		to := make([]string, 0)
		for _, addr := range msg.TssMsg.To {
			to = append(to, addr.Id)
		}

		key := fmt.Sprintf("Type=%s|From=%s|To=%s|Broadcast=%t", msg.TssMsg.Type,
			msg.TssMsg.From.Id,
			to,
			msg.TssMsg.IsBroadcast)

		t.KeygenInstance.KeygenMsgsStorage.Lock()
		_, ok := t.KeygenInstance.KeygenMsgsStorage.M[key]
		if !ok {
			t.KeygenInstance.KeygenMsgsStorage.M[key] = *msg.TssMsg
		}
		t.KeygenInstance.KeygenMsgsStorage.Unlock()
		for key, _ := range t.KeygenInstance.KeygenMsgsStorage.M {
			t.Logger.Info("KeygenMsgsStorage", zap.String("key", key))
		}
		return

	case KeygenCancelledMsgType:
		if t.KeygenInstance.IsStarted.Load() == true {
			t.KeygenInstance.IsStarted.Store(false)
			t.KeygenInstance.StopChan <- CommunicationError{
				PeerAddr:  msg.CommunicationError.PeerAddr,
				Operation: msg.CommunicationError.Operation,
				Time:      msg.CommunicationError.Time,
			}
		} else {
			t.Logger.Info("service -> HandleP2Pmessage - error -> keygen error already handled")
		}

		// keysign
	case KeysignStartMsgType:
		t.NewTsskeySign(t.Parties, t.Quorum)

		t.Logger.Info("tss -> HandleP2PMessage -> keysign start", zap.Int("parties", t.Parties),
			zap.Int("quorum", t.Quorum), zap.String("msg", msg.KeysignRequest.Msg))

		partiesID, localPartyID, err := t.GetParties(t.Pubkey)
		if err != nil {
			t.Logger.Error("tss -> HandleP2PMessage -> KeysignStartMsgType -> GetParties")
			return
		}

		_, err = t.KeysignInstance.SignMessage(msg.KeysignRequest, partiesID, localPartyID, t.Key)
		if err != nil {
			t.Logger.Error("tss -> HandleP2Pmessage -> SignMessage", zap.Error(err))
			return
		}
	case KeysignMsgType:
		t.Logger.Info("service -> HandleP2Pmessage -> keysign ->  got msg", zap.String("from", p2pMsg.From),
			zap.Strings("to", p2pMsg.To), zap.String("type", msg.Type), zap.String("round", msg.Round))
		// @TODO: use not broadcasted msgs?
		if msg.TssMsg.From.Id == t.Pubkey {
			t.Logger.Error("tss -> handlers -> KeysignMsgType", zap.String("in id", msg.TssMsg.From.Id), zap.Error(errors.New("msg from own ID")))
			return
		}

		to := make([]string, 0)
		for _, addr := range msg.TssMsg.To {
			to = append(to, addr.Id)
		}

		key := fmt.Sprintf("Type=%s|From=%s|To=%s|Broadcast=%t", msg.TssMsg.Type,
			msg.TssMsg.From.Id,
			to,
			msg.TssMsg.IsBroadcast)

		t.KeysignInstance.KeysignMsgsStorage.Lock()
		_, ok := t.KeysignInstance.KeysignMsgsStorage.M[key]
		if !ok {
			t.KeysignInstance.KeysignMsgsStorage.M[key] = *msg.TssMsg
		}
		t.KeysignInstance.KeysignMsgsStorage.Unlock()
		// for key, _ := range t.KeysignInstance.KeysignMsgsStorage.M {
		// 	t.Logger.Info("KeysignMsgsStorage", zap.String("key", key))
		// }
		return

	case KeysignOneRoundMsgType:
		// @TODO: use not broadcasted msgs?
		if msg.PartyID.Id == t.LocalPartyID.Id {
			t.Logger.Error("tss -> handlers -> KeysignOneRoundMsgType", zap.String("in id", msg.TssMsg.From.Id), zap.Error(errors.New("msg from own ID")))
			return
		}

		t.KeysignInstance.OneRoundMsgCh <- &P2pMessage{
			PartyID: msg.PartyID,
			Si:      msg.Si,
		}
	case KeysignCancelledMsgType:
		t.KeysignInstance.IsStarted.Store(false)
		if t.KeysignInstance.IsStarted.Load() == true {
			t.KeysignInstance.StopChan <- CommunicationError{
				PeerAddr:  msg.CommunicationError.PeerAddr,
				Operation: msg.CommunicationError.Operation,
				Time:      msg.CommunicationError.Time,
			}
		} else {
			t.Logger.Info("service -> HandleP2Pmessage - error -> keysign error already handled")
		}
	}
}

func (t *TssServer) HandleHandshake(msg *P2pMessage) error {
	t.RWMutex.Lock()
	defer t.RWMutex.Unlock()

	if addr, ok := t.ConnectionStorage[msg.Pubkey]; ok && addr == msg.PeerAddr {
		t.Logger.Debug("tss -> HandleP2Pmessage -> already in connection storage", zap.String("pubkey", msg.Pubkey), zap.String("addr", msg.PeerAddr))
		return nil
	}
	t.ConnectionStorage[msg.Pubkey] = msg.PeerAddr
	err := t.SendHandshake(msg.PeerAddr)
	if err != nil {
		return fmt.Errorf("SendHandshake :%w", err)
	}
	return nil
}

// handling disconnect signal from peer
func (t *TssServer) HandleDisconnect(msg *P2pMessage) error {
	t.RWMutex.Lock()
	defer t.RWMutex.Unlock()

	if addr, ok := t.ConnectionStorage[msg.Pubkey]; ok && addr == msg.PeerAddr {
		delete(t.ConnectionStorage, msg.Pubkey)
	}

	return nil
}
