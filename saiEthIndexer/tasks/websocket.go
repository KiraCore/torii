package tasks

import (
	"bytes"
	"net/http"

	"github.com/saiset-co/saiEthIndexer/config"
)

type WebsocketManager struct {
	Config config.Configuration
}

func NewWebSocketManager(c config.Configuration) *WebsocketManager {
	return &WebsocketManager{
		Config: c,
	}
}

func (w WebsocketManager) SendMessage(message string, token string) error {
	url := w.Config.Specific.WebSocket.URL + "?method=broadcast&message=" + token + "|" + message
	req, err := http.NewRequest("GET", url, new(bytes.Buffer))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	_, err = client.Do(req)

	if err != nil {
		return err
	}

	client.CloseIdleConnections()
	return nil
}
