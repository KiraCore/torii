package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"

	"go.uber.org/zap"
)

// Server part of core
type Server struct {
	Address struct {
		IP        net.IP
		PunchPort string
	}
	Connections struct {
		In  *net.UDPConn
		Out *net.UDPConn
	}
	AddrChan          chan string
	FilterConnections func(interface{}) bool
	*sync.RWMutex                           // for BufferedMsgs map
	BufferedMsgs      map[string][]*Message // map[buffered msg hash]*Message
}

// server stats. For testing/debugging
type AppStats struct {
	PunchPort         string          `json:"p,omitempty"`
	ConnectionStorage map[string]bool `json:"m,omitempty"`
	SavedMessages     map[string]bool `json:"h,omitempty"`
	IP                string          `json:"ip,omitempty"`
}

// Server logic
func (c *Core) Serve(filter filterConnections) (err error) {
	serverAddr, err := net.ResolveUDPAddr("udp4", ":"+c.Config.P2P.Port)
	if err != nil {
		return err
	}

	c.Server.Address.IP = serverAddr.IP

	c.Server.Connections.In, err = net.ListenUDP("udp4", serverAddr)
	if err != nil {
		return err
	}

	defer func() {
		//s.L.Debug("StartServer -> stopped")
		c.Server.Connections.In.Close()
	}()

	// listen msgs from broadcast

	for {
		buf := make([]byte, 500000)
		n, newClientAddr, err := c.Server.Connections.In.ReadFromUDP(buf)
		if err != nil {
			c.Logger.Warn("StartServer", zap.Error(err))
			continue
		}

		//	go func(buf []byte) {
		incomingRequest := Request{}
		//s.L.Debug("StartServer", zap.Any("incomingRequest", incomingRequest))
		err = json.Unmarshal(buf[:n], &incomingRequest)
		if err != nil {
			c.Logger.Error("StartServer", zap.Error(err))
			continue
		}

		c.Logger.Debug("StartServer -> incomingRequest",
			zap.String("Type", incomingRequest.Type),
			zap.String("from", newClientAddr.String()),
			zap.String("to", incomingRequest.RemoteAddr))

		addr, err := net.ResolveUDPAddr("udp4", incomingRequest.RemoteAddr)
		if err != nil {
			c.Logger.Error("p2p -> StartServer -> ResolveUDPAddr", zap.Error(err))
			continue
		}

		c.Server.Address.IP = addr.IP

		if incomingRequest.Type == punchRequest {
			err = c.CheckAllowConn(filter, newClientAddr.IP.String(), fmt.Sprintf("%d", (newClientAddr.Port)))
			if err != nil {
				continue
			}
		}

		switch incomingRequest.Type {
		case punchRequest:
			c.HandlePunch(incomingRequest, newClientAddr)
		case handshakeRequest:
			c.HandleHandshake(incomingRequest, newClientAddr)
		case EventRequest:
			c.HandleEvent(&incomingRequest)
		default:
			c.HandleMessage(incomingRequest, newClientAddr)
		}
	}
}

// Add client to connection list
// if ip already exists - do not overwrite (for local tests)
func (c *Core) addConnection(ip, port, punchPort string) {
	// avoid overwriting for local run
	if port == punchPort && ip == c.Server.Address.IP.String() {
		return
	}

	//s.Logger.Debug("p2p - server - addClient", zap.Any("actual connection list before", s.ConnMap))
	address := net.JoinHostPort(ip, port)
	c.Lock()
	c.ConnectionStorage[address] = true
	c.Unlock()
	c.Logger.Debug("p2p - server - addClient", zap.String("address", address))
}

// Get connection
func (c *Core) GetConnection(ip, port string) (string, error) {
	c.RLock()
	defer c.RUnlock()
	address := net.JoinHostPort(ip, port)
	i := c.ConnectionStorage[address]
	if i {
		return address, nil
	}
	return "", errors.New("no such client")
}

// Remove connection from connection storage
func (c *Core) RemoveConnection(address string) {
	c.RLock()
	defer c.RUnlock()
	delete(c.ConnectionStorage, address)
	c.Logger.Debug("p2p - server - deleteClient", zap.String("address", address))
}

// disconnect to connected nodes
func (c *Core) Disconnect() {
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

		data, err := json.Marshal(disconnectionReq)
		if err != nil {
			c.Logger.Error("p2p - server - Disconnect - Unmarshal", zap.Error(err))
			return
		}

		addr, err := net.ResolveUDPAddr("udp4", node)
		if err != nil {
			c.Logger.Error("p2p - server - Disconnect - ResolveUDPAddr", zap.Error(err))
			continue
		}

		_, err = c.Server.Connections.In.WriteToUDP(data, addr)
		if err != nil {
			c.Logger.Error("p2p - server - Disconnect - WriteToUDP", zap.Error(err))
			continue
		}
	}
	c.Logger.Debug("p2p -> server -> disconnected to all peers")
}

// (p2p-new) get new p2p stats of p2p App
// return stats of p2p server for debbuging purposes
func (c *Core) GetStats() AppStats {
	p2pStats := AppStats{
		SavedMessages:     make(map[string]bool),
		PunchPort:         c.Server.Address.PunchPort,
		ConnectionStorage: c.ConnectionStorage,
		IP:                c.Server.Address.IP.String(),
	}

	it := c.Cache.Iterator()
	for it.SetNext() {
		entryInfo, err := it.Value()
		if err != nil {
			c.Logger.Error("p2p -> server -> GetStats -> it.Value", zap.Error(err))
			continue
		}
		p2pStats.SavedMessages[entryInfo.Key()] = true
	}

	return p2pStats
}
