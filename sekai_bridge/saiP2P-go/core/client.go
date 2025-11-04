package core

import (
	"encoding/json"
	"net"
	"strconv"

	"go.uber.org/zap"
)

// Client part of core
type Client struct {
	Address struct {
		Remote string
		Local  string
	}
	Messages struct { // client statistics
		System   int
		Unique   int
		Messages int
	}
}

// ConnectToPeer - client logic when peer connect to a provider
func (c *Core) ConnectToPeer() error {
	defer func() {
		c.Logger.Error("close connection")
	}()

	request := Request{Type: punchRequest, LocalAddr: c.Client.Address.Local, RemoteAddr: c.Client.Address.Remote}
	buf := make([]byte, 500000)

	serverUDPAddr, err := net.ResolveUDPAddr("udp4", c.Client.Address.Remote)
	if err != nil {
		c.Logger.Error("Connect", zap.Error(err))
		return err
	}

	c.Server.Connections.Out, err = net.ListenUDP("udp4", nil)
	if err != nil {
		c.Logger.Error("Connect", zap.Error(err))
		return err
	}

	defer func() {
		c.Server.Connections.Out.Close()
	}()

	// c.Logger.Debug("Connect", zap.Any("request", request))

	err = request.Send(c.Server.Connections.Out, serverUDPAddr)
	if err != nil {
		c.Logger.Error("p2p -> client -> request.Send", zap.Error(err))
		return err
	}

	for {
		n, serverAddr, err := c.Server.Connections.Out.ReadFromUDP(buf)
		serverIP := serverAddr.IP.String()
		serverPort := strconv.Itoa(serverAddr.Port)
		if err != nil {
			c.Logger.Error("Connect", zap.Error(err))
			continue
		}

		var message map[string]interface{}

		err = json.Unmarshal(buf[:n], &message)
		if err != nil {
			c.Logger.Error("Connect", zap.String("buf", string(buf)), zap.Error(err))
			continue
		}

		//c.Logger.Debug("p2p -> client -> incoming request", zap.Any("message", message["type"]))

		if messageType, ok := message["type"]; ok {
			switch messageType {
			case punchResponse:
				c.Logger.Debug("punchResponse", zap.Any("message", message))

				err := c.ClientPunchResponseHandler(message, serverIP, serverPort)
				if err != nil {
					c.Logger.Error("p2p -> client -> ClientPunchResponseHandler", zap.Error(err))
					continue
				}
			case handshakeRequest:
				c.Logger.Debug("handshakeRequest", zap.Any("message", message))

				err := c.ClientHandshakeRequestHandler(message)
				if err != nil {
					c.Logger.Error("p2p -> client -> ClientHandshakeRequestHandler", zap.Error(err))
					continue
				}
			case handshakeResponse:
				c.Logger.Debug("handshakeResponse", zap.Any("message", message))

				err := c.ClientHandshakeResponseHandler(message, serverIP, serverPort)
				if err != nil {
					c.Logger.Error("p2p -> client -> ClientHandshakeResponseHandler", zap.Error(err))
					continue
				}
			case MessageRequest:
				c.Logger.Debug("MessageRequest", zap.Any("message", message))

				err := c.ClientMessageHandler(buf[0:n])
				if err != nil {
					c.Logger.Error("p2p -> client -> ClientMessageHandler", zap.Error(err))
					continue
				}
			case EventRequest:
				err := c.ClientEventHandler(buf[0:n])
				if err != nil {
					c.Logger.Error("p2p -> client -> ClientEventHandler", zap.Error(err))
					continue
				}

			default:
				err := c.ClientDefaultHandler(message)
				if err != nil {
					c.Logger.Error("p2p -> client -> ClientDefaultHandler", zap.Error(err))
					continue
				}
			}
		}
		// c.Logger.Debug("client - Messages",
		// 	zap.Int("messages", c.Client.Messages.Messages),
		// 	zap.Int("unique messages", c.Client.Messages.Unique),
		// 	zap.Int("system messages", c.Client.Messages.System),
		// )
	}
}
