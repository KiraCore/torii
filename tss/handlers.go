package tss

import (
	"fmt"

	"go.uber.org/zap"
)

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
