package notifier

import (
	"bytes"
	jsoniter "github.com/json-iterator/go"
	"github.com/saiset-co/saiEthIndexer/utils"
)

type Notifier interface {
	SendTx(data interface{}) error
}

type notifier struct {
	senderID string
	address  string
	email    string
	password string
	token    string
}

type notificationData struct {
	From string      `json:"from"`
	Tx   interface{} `json:"tx"`
}

type Metadata struct {
	Token string `json:"token"`
}

type notificationRequest struct {
	Method   string           `json:"method"`
	Data     notificationData `json:"data"`
	Metadata Metadata         `json:"metadata"`
}

func NewNotifier(senderID, email, password, token, address string) Notifier {
	return &notifier{
		senderID: senderID,
		email:    email,
		password: password,
		token:    token,
		address:  address,
	}
}

func (n *notifier) SendTx(tx interface{}) error {
	req := notificationRequest{
		Method: "notify",
		Data: notificationData{
			From: n.senderID,
			Tx:   tx,
		},
		Metadata: Metadata{
			Token: n.token,
		},
	}

	payload, err := jsoniter.Marshal(&req)
	if err != nil {
		return err
	}

	_, err = utils.SaiQuerySender(bytes.NewReader(payload), n.address, n.token)

	return err
}
