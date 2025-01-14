package core

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/saiset-co/saiP2P-go/config"
	"github.com/saiset-co/saiP2P-go/utils"
	"go.uber.org/zap"
)

const (
	handshakeRequest  = "handshakeRequest"
	handshakeResponse = "handshakeResponse"
	punchRequest      = "punchRequest"
	punchResponse     = "punchResponse"
	MessageRequest    = "message"
	EventRequest      = "event"

	statusConnectionRejected = "CONNECTION_REJECTED"
)

// var (
// 	testChunkMap = make(map[string]ChunkMsg)
// )

// type ChunkMsg struct {
// 	Time time.Time
// 	Part int
// }

// Main p2p struct
type Core struct {
	Server *Server       // servert part
	Client *Client       // client part
	Logger *zap.Logger   // logger
	Config config.Config // configuration
	sync.RWMutex
	ConnectionStorage map[string]bool // p2p connections listed here
	//	SavedMessages     map[string]bool // saved messages, to prevent double messages sending
	MsgCh chan *Request // to handle messages inside core
	Cache *bigcache.BigCache
}

// filter connections func type
type filterConnections func(interface{}) bool

// Init core
func Init(config config.Config, f filterConnections) *Core {
	// initialize cache
	cacheConfig := bigcache.Config{
		Shards:      1024,
		OnRemove:    nil,
		LifeWindow:  time.Duration(config.Cache.TTL) * time.Second,
		CleanWindow: time.Duration(config.Cache.CleanPeriod) * time.Second,
	}

	cache, err := bigcache.New(context.Background(), cacheConfig)
	if err != nil {
		log.Fatalf("p2p -> Server -> bigcache.New : %s", err.Error())
	}

	logger, err := utils.BuildLogger(true)
	if err != nil {
		log.Fatalf("p2p -> Server -> BuildLogger : %s", err.Error())
	}

	core := &Core{
		ConnectionStorage: make(map[string]bool),
		//	SavedMessages:     make(map[string]bool),
		MsgCh:  make(chan *Request),
		Config: config,
		Logger: logger,
		Cache:  cache,
	}

	server := &Server{
		AddrChan:          make(chan string),
		FilterConnections: f,
		RWMutex:           new(sync.RWMutex),
		BufferedMsgs:      map[string][]*Message{},
	}

	core.Server = server

	return core
}

// Init client
func (c *Core) CreateClient(peer string) *Core {
	client := Client{}
	client.Address.Local = c.Server.Address.IP.String() + ":" + c.Config.P2P.Port
	client.Address.Remote = peer

	c.Client = &client

	return c
}

// Run main p2p struct
func (c *Core) Run(f filterConnections) {
	c.Logger.Debug("Start", zap.Any("peers", c.Config.Peers))

	go c.ProcessHandshakes()

	if len(c.Config.Peers) > 0 {
		go c.RunClient()
	}

	if err := c.Serve(f); err != nil {
		c.Logger.Error("Start", zap.Error(err))
	}
}

// Broadcasting messages logic
func (c *Core) SendMsg(mes []byte, to []string, senderAddr string) error {
	time.Sleep(5 * time.Millisecond)
	if len(mes) > c.Config.UDP.MsgBufferSize {
		msgs := c.PrepareMsgChunks(mes, to, senderAddr)
		for _, msg := range msgs {
			time.Sleep(500 * time.Millisecond)
			c.DistributeMsg(to, *msg)

		}
	} else {
		c.Cache.Set(string(mes), nil)

		var message = Message{
			From: senderAddr,
			To:   to,
			Data: mes,
		}

		c.DistributeMsg(to, message)

	}
	return nil
}

// prepare messages, if incoming data bigger than udp datagram size
func (c *Core) PrepareMsgChunks(data []byte, to []string, senderAddr string) []*Message {
	msgs := make([]*Message, 0)
	h := sha1.New()
	h.Write(data)
	msgHash := hex.EncodeToString(h.Sum(nil))
	byteChunks := utils.ChunkSlice(data, c.Config.UDP.MsgBufferSize)
	for idx, chunk := range byteChunks {
		c.Cache.Set(string(chunk), nil)

		msg := &Message{
			From:       senderAddr,
			To:         to,
			Data:       chunk,
			Hash:       msgHash,
			TotalParts: len(byteChunks),
			Part:       idx + 1,
		}
		if idx+1 == len(byteChunks) {
			msg.Last = true
		}

		msgs = append(msgs, msg)
	}
	return msgs
}

// messages distribution
func (c *Core) DistributeMsg(to []string, message Message) {
	// send msg only to recepients if recepients exists
	if len(to) != 0 {
		for _, address := range to {
			// check if recepient is from connected list
			if c.ConnectionStorage[address] {
				err := c.sendMsg(message, address)
				if err != nil {
					continue
				}
			} else {
				c.Logger.Debug("p2p -> server -> Send -> recepient is not from connected list", zap.String("recepient", address))
				continue
			}

		}
		//send msg to all connections if recepients is empty
	} else {
		for address := range c.ConnectionStorage {
			err := c.sendMsg(message, address)
			if err != nil {
				continue
			}
		}
	}
}

func (c *Core) sendMsg(message Message, address string) error {

	if address == net.JoinHostPort(c.Server.Address.IP.String(), c.Config.P2P.Port) || address == net.JoinHostPort(c.Server.Address.IP.String(), c.Server.Address.PunchPort) {
		c.Logger.Error("p2p -> server -> Send : skip send to myself", zap.String("target", address))
		return errors.New("skip sending to myself")
	}
	clientAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		c.Logger.Error("p2p -> server -> sendMsg -> ResolveUDPAddr", zap.Error(err), zap.String("addr", address))
		return err
	}

	var localAddr string
	if c.Server.Address.PunchPort == "" {
		localAddr = net.JoinHostPort(c.Server.Address.IP.String(), c.Config.P2P.Port)
	} else {
		localAddr = net.JoinHostPort(c.Server.Address.IP.String(), c.Server.Address.PunchPort)
	}

	request := Request{Type: "message", LocalAddr: localAddr, RemoteAddr: clientAddr.String(), Message: message}

	if c.Server.Connections.Out != nil {
		err = request.Send(c.Server.Connections.Out, clientAddr)
		if err != nil {
			c.Logger.Error("p2p -> server -> sendMsg -> request.Send(out)", zap.Error(err), zap.String("addr", clientAddr.String()))
		}
	} else {
		err = request.Send(c.Server.Connections.In, clientAddr)
		if err != nil {
			c.Logger.Error("p2p -> server -> sendMsg -> request.Send(In)", zap.Error(err), zap.String("addr", clientAddr.String()))
		}
	}
	c.Logger.Debug("p2p -> server -> sendMsg", zap.String("target", address))
	return nil
}

// Run client core part, if peers provided
func (c *Core) RunClient() {
	for _, peer := range c.Config.Peers {
		//c.Logger.Debug("StartClient", zap.String("peer", peer))
		err := c.CreateClient(peer).ConnectToPeer()
		if err != nil {
			c.Logger.Error("StartClient", zap.Error(err))
		}
	}
}

// Process incoming connections
func (c *Core) ProcessHandshakes() {
	for {
		peer := <-c.Server.AddrChan
		//		c.Logger.Debug("HandleList", zap.Any("peer", peer))

		if peer == net.JoinHostPort(c.Server.Address.IP.String(), c.Config.P2P.Port) || peer == net.JoinHostPort(c.Server.Address.IP.String(), c.Server.Address.PunchPort) {
			continue
		}

		if existing := c.ConnectionStorage[peer]; existing != true || existing == true {
			serverUDPAddr, err := net.ResolveUDPAddr("udp4", peer)
			if err != nil {
				c.Logger.Error("HandleList", zap.Error(err))
				continue
			}

			request := Request{Type: handshakeRequest, LocalAddr: c.Server.Address.IP.String() + ":" + c.Server.Address.PunchPort, RemoteAddr: peer}

			err = request.Send(c.Server.Connections.Out, serverUDPAddr)
			if err != nil {
				c.Logger.Error("p2p -> core -> request.Send", zap.Error(err))
				continue
			}
		}
	}
}

// Return address ip+port if it is peernode, ip+punchport if not
func (c *Core) GetRealAddress() string {
	var address string
	if c.Server.Address.PunchPort == "" {
		address = net.JoinHostPort(c.Server.Address.IP.String(), c.Config.P2P.Port)
	} else {
		address = net.JoinHostPort(c.Server.Address.IP.String(), c.Server.Address.PunchPort)
	}
	return address
}

// Send disconnect event type to all connected nodes
func (c *Core) ProvideDisconnection() {
	c.RLock()
	defer c.RUnlock()

	address := c.GetRealAddress()

	disconnectionReq := &Request{
		Type: EventRequest,
		Event: &Event{
			Address: address,
			Type:    DisconnectionEventType,
		},
		LocalAddr: address,
	}

	for node := range c.ConnectionStorage {
		disconnectionReq.RemoteAddr = node

		addr, err := net.ResolveUDPAddr("udp4", node)
		if err != nil {
			c.Logger.Error("p2p - server - Disconnect - ResolveUDPAddr", zap.Error(err))
			continue
		}

		err = disconnectionReq.Send(c.Server.Connections.In, addr)
		if err != nil {
			c.Logger.Error("p2p -> core -> ProvideDisconnection -> request.Send", zap.Error(err))
			continue
		}
	}
	c.Logger.Debug("p2p -> server -> disconnected to all peers")
}

// get next message from p2p. If message sending by chunks, NextMsg will return whole message, when all chunks
// got and checked. If error is nil and message is nil -> msg was not fully assembled
func (c *Core) NextMsg(ctx context.Context) (*Message, error) {
	select {
	case req, ok := <-c.MsgCh:
		if !ok {
			return nil, errors.New("error")
		}
		c.Logger.Debug("NextMsg", zap.String("hash", req.Message.Hash),
			zap.Int("part", req.Message.Part),
			zap.Int("total parts", req.Message.TotalParts),
			zap.Bool("last", req.Message.Last))

		if req.Message.Hash != "" && req.Message.Part != 0 && req.Message.TotalParts != 0 {
			msg, err := c.HandleIncomingChunkedMsg(&req.Message)
			if err != nil {
				return nil, fmt.Errorf("core -> NextMsg -> HandleIncomingChunkedMsg : %w", err)
			}

			if msg == nil { // not full message assembled yet
				return nil, nil
			}

			return msg, nil
		}
		return &req.Message, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *Core) HandleIncomingChunkedMsg(msg *Message) (*Message, error) {
	c.Logger.Debug("HandleIncomingChunkedMsg", zap.String("hash", msg.Hash),
		zap.Int("part", msg.Part),
		zap.Int("total parts", msg.TotalParts),
		zap.Bool("last", msg.Last))

	c.Server.RWMutex.Lock()
	defer c.Server.Unlock()
	slice, ok := c.Server.BufferedMsgs[msg.Hash]
	if !ok { //first msg
		slice = make([]*Message, 0)
		slice = append(slice, msg)
		c.Server.BufferedMsgs[msg.Hash] = slice
		return nil, nil
	}
	if len(slice) == msg.TotalParts-1 { // last msg got
		slice = append(slice, msg)
		// sort message
		sort.SliceStable(slice, func(i, j int) bool {
			return slice[i].Part < slice[j].Part
		})
		finalData := make([]byte, 0)
		for _, msg := range slice {
			finalData = append(finalData, msg.Data...)
		}
		finalMsg := &Message{
			From:       msg.From,
			To:         msg.To,
			Data:       finalData,
			Hash:       msg.Hash,
			TotalParts: msg.TotalParts,
			Part:       msg.Part,
			Last:       msg.Last,
		}

		h := sha1.New()
		h.Write(finalMsg.Data)
		calculatedHash := hex.EncodeToString(h.Sum(nil))
		if msg.Hash != calculatedHash {
			//c.Logger.Sugar().Errorf("MAP : %+v\n", testChunkMap)
			return nil, fmt.Errorf("HandleIncomingChunkedMsg error, expected hash : %s, got %s", finalMsg.Hash, calculatedHash)
		}
		// clean buffered msgs storage
		delete(c.Server.BufferedMsgs, msg.Hash)
		return finalMsg, nil
	}
	slice = append(slice, msg)
	c.Server.BufferedMsgs[msg.Hash] = slice
	return nil, nil
}

// @TODO: old version
// func (c *Core) HandleIncomingChunkedMsg(msg *Message) (*Message, error) {
// 	c.Logger.Debug("HandleIncomingChunkedMsg", zap.String("hash", msg.Hash),
// 		zap.Int("part", msg.Part),
// 		zap.Int("total parts", msg.TotalParts),
// 		zap.Bool("last", msg.Last))

// 	// testChunkMap[msg.Hash] = ChunkMsg{
// 	// 	Time: time.Now(),
// 	// 	Part: msg.Part,
// 	// }
// 	if msg.Part == 1 { // if first message was got, save to bufferedMsg map
// 		finalMsg := &Message{
// 			From:       msg.From,
// 			To:         msg.To,
// 			Data:       msg.Data,
// 			Hash:       msg.Hash,
// 			TotalParts: msg.TotalParts,
// 			Part:       msg.Part,
// 			Last:       msg.Last,
// 		}
// 		c.RWMutex.Lock()
// 		c.Server.BufferedMsgs[finalMsg.Hash] = finalMsg
// 		c.RWMutex.Unlock()
// 	} else {
// 		c.RWMutex.Lock()
// 		finalMsg, ok := c.Server.BufferedMsgs[msg.Hash]
// 		if !ok {
// 			//c.Logger.Sugar().Errorf("MAP : %+v\n", testChunkMap)
// 			return nil, fmt.Errorf("msg was not found in buffered map, hash : %s", msg.Hash)
// 		}
// 		c.RWMutex.Unlock()
// 		finalMsg.Data = append(finalMsg.Data, msg.Data...)
// 		if msg.Last { // check final msg hash
// 			h := sha1.New()
// 			h.Write(finalMsg.Data)
// 			calculatedHash := hex.EncodeToString(h.Sum(nil))
// 			if msg.Hash != calculatedHash {
// 				//c.Logger.Sugar().Errorf("MAP : %+v\n", testChunkMap)
// 				return nil, fmt.Errorf("HandleIncomingChunkedMsg error, expected hash : %s, got %s", finalMsg.Hash, calculatedHash)
// 			}
// 			// clean buffered msgs storage
// 			c.RWMutex.Lock()
// 			delete(c.Server.BufferedMsgs, msg.Hash)
// 			c.RWMutex.Unlock()

// 			return finalMsg, nil

// 		}
// 		c.RWMutex.Lock()
// 		c.Server.BufferedMsgs[msg.Hash] = finalMsg
// 		c.RWMutex.Unlock()
// 	}
// 	return nil, nil
// }

// Check connection ability
func (c *Core) CheckAllowConn(filter filterConnections, ip, port string) error {
	// check if connection is allowed
	// s.IP.String() as example
	if !filter(c.Server.Address.IP.String()) {
		addr, err := net.ResolveUDPAddr("udp4", net.JoinHostPort(ip, port))
		if err != nil {
			c.Logger.Error("p2p -> server -> CheckAllowConn -> ResolveUDPAddr", zap.Error(err))
			return err
		}
		c.Logger.Debug("p2p - server - StartServer - AllowConnection - reject connection", zap.String("host", addr.IP.String()), zap.String("port", port))
		var usedPort string
		if c.Server.Address.PunchPort == "" {
			usedPort = c.Config.P2P.Port
		} else {
			usedPort = c.Server.Address.PunchPort
		}
		response := Response{Type: punchResponse, Status: statusConnectionRejected, Ip: c.Server.Address.IP.String(), Port: usedPort}
		err = response.Send(c.Server.Connections.In, addr)
		if err != nil {
			c.Logger.Error("p2p -> server -> CheckAllowConn -> response.Send", zap.Error(err))
			return fmt.Errorf("response.Send :%w", err)
		}
		return errors.New("p2p -> server -> CheckAllowConn -> reject connection")
	}
	return nil
}

// Graceful shutdown
func (c *Core) GracefulShutdown() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		c.Logger.Info("main -> got interrupt", zap.String("signal", s.String()))
		c.ProvideDisconnection()
		c.Logger.Info("main -> interrupt - success", zap.String("signal", s.String()))
	}
}

// check cache, return nil,if found, if not and error is EntryNotFound = set cache
func (c *Core) CheckCache(key string) (bool, error) {
	_, err := c.Cache.Get(key)
	if err != nil {
		if errors.Is(err, bigcache.ErrEntryNotFound) {
			err = c.Cache.Set(key, nil)
			if err != nil {
				return false, err
			}
			return false, nil
		}
		return false, err
	}
	return true, nil
}
