package tss

// func (t *TssServer) ProcessOutCh(msg tsslib.Message, parties int) error {
// 	b, r, err := msg.WireBytes()
// 	if err != nil {
// 		return fmt.Errorf("WireBytes : %w", err)
// 	}

// 	tssMsg := TssMessage{
// 		From:        msg.GetFrom(),
// 		To:          msg.GetTo(),
// 		IsBroadcast: msg.IsBroadcast(),
// 		Bytes:       b,
// 		Type:        msg.Type(),
// 		Routing:     r,
// 	}

// 	p2pMsg := P2pMessage{
// 		Type:        KeygenMsgType,
// 		TssMsg:      &tssMsg,
// 		KeygenRound: msg.Type(),
// 	}

// 	data, err := json.Marshal(p2pMsg)
// 	if err != nil {
// 		return fmt.Errorf("Marshal : %w", err)
// 	}

// 	if msg.IsBroadcast() { // send to all
// 		time.Sleep(1 * time.Second)
// 		err = t.P2p.SendMsg(data, nil, t.P2p.GetRealAddress())
// 		if err != nil {
// 			fmt.Println("SendMsg:", err)
// 		}
// 	} else { // send to specified peer
// 		addrs := t.GetPartyIDpeerAddr()
// 		time.Sleep(1 * time.Second)
// 		err = t.P2p.SendMsg(data, addrs, t.P2p.GetRealAddress())
// 		if err != nil {
// 			fmt.Println("SendMsg:", err)
// 		}
// 	}
// 	t.Logger.Info("processOutCh - msg sent",
// 		zap.String("sender_ID", p2pMsg.TssMsg.From.Id),
// 		zap.String("type", p2pMsg.TssMsg.Type))

// 	if msg.Type() != "KGRound2Message1" {
// 		t.UpdateForRound(&tssMsg, parties)
// 	}
// 	t.Logger.Info("RETURN?")
// 	time.Sleep(3 * time.Second)
// 	return nil
// }

// func (t *TssServer) Update(msg *TssMessage) error {
// 	parsedMsg, err := tsslib.ParseWireMessage(msg.Bytes, msg.From, msg.IsBroadcast)
// 	if err != nil {
// 		return fmt.Errorf("ParseWireMessage : %w", err)
// 	}

// 	if t.LocalPartyID.Index-1 == parsedMsg.GetFrom().Index {
// 		return errors.New("tried to send a message to itself")
// 	}

// 	if t.PG != nil {
// 		t.Logger.Info("tss - PG updater started", zap.String("type", msg.Type), zap.String("from", msg.From.Id))
// 		go t.SharedPartyUpdater(t.PG, parsedMsg, t.ErrCh)
// 		// } else if t.TssKeygen.PS != nil {
// 		// 	fmt.Println("Start PS update")
// 		// 	go t.SharedPartyUpdater(t.PS, parsedMsg, t.ErrCh)
// 		// }
// 	}
// 	t.Logger.Info("process - Update - Success", zap.String("type", msg.Type),
// 		zap.String("from", msg.From.Id))
// 	return nil
// }

// func (t *TssServer) SharedPartyUpdater(party tsslib.Party, msg tsslib.Message, errCh chan<- *tss.Error) {
// 	// do not send a message from this party back to itself
// 	if party.PartyID() == msg.GetFrom() {
// 		return
// 	}
// 	bz, _, err := msg.WireBytes()
// 	if err != nil {

// 		err := fmt.Errorf("WireBytes : %w", err)
// 		errCh <- party.WrapError(err)
// 		return
// 	}

// 	if _, err := party.UpdateFromBytes(bz, msg.GetFrom(), msg.IsBroadcast()); err != nil {
// 		err := fmt.Errorf("UpdateFromBytes err =  %w, type = %s, from = %s", err, msg.Type(), msg.GetFrom().Id)
// 		errCh <- party.WrapError(err)
// 	}

// 	t.Logger.Info("process - SharedPartyUpdater - Success", zap.String("type", msg.Type()),
// 		zap.String("from", msg.GetFrom().Id))
// }
