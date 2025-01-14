package tss

import (
	"encoding/json"
	"fmt"
	"time"

	p2p "github.com/saiset-co/saiP2P-go/core"
)

// error for communication, for example, when some peer got error
// and we should stop keygen/keysign operation

const (
	KeygenOperation  = "keygen_operation"
	KeysignOperation = "keysign_operation"
)

type CommunicationError struct {
	PeerAddr  string    `json:"peer_id"`
	Operation string    `json:"operation"`
	Time      time.Time `json:"time"`
}

// when we should decide, which operation (keysign, keygen) was failed
func (t *TssServer) HandleUnmarshalError(p2pMsg *p2p.Message) (*CommunicationError, error) {
	var operation string
	if t.KeygenInstance.IsStarted.Load() == true || t.KeysignInstance == nil {
		operation = KeygenOperation
	} else {
		operation = KeysignOperation
	}

	return &CommunicationError{
		PeerAddr:  p2pMsg.From,
		Operation: operation,
		Time:      time.Now(),
	}, nil
}

// notify nodes about error
func (t *TssServer) NotifyAboutError(commError *CommunicationError) error {
	var msgType string
	switch commError.Operation {
	case KeygenOperation:
		msgType = KeygenCancelledMsgType
	case KeysignCancelledMsgType:
		msgType = KeysignCancelledMsgType
	}

	errMsg := P2pMessage{
		Type:               msgType,
		CommunicationError: *commError,
	}
	data, err := json.Marshal(errMsg)
	if err != nil {
		return fmt.Errorf("notifyAboutError : %w", err)
	}

	err = t.P2p.SendMsg(data, nil, t.P2p.GetRealAddress())
	if err != nil {
		return fmt.Errorf("SendMsg : %w", err)
	}
	return nil
}
