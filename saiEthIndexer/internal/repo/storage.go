package repository

import (
	"bytes"
	jsoniter "github.com/json-iterator/go"
	"github.com/saiset-co/sai-storage-mongo/external/adapter"
	"github.com/saiset-co/saiEthIndexer/utils"
)

type Repo interface {
	Create(data interface{}) error
}

type MongoRepo struct {
	collection string
	address    string
	email      string
	password   string
	token      string
}

func NewMongoRepo(collection, email, password, token, address string) Repo {
	return &MongoRepo{
		collection: collection,
		email:      email,
		password:   password,
		token:      token,
		address:    address,
	}
}

func (mr *MongoRepo) Create(data interface{}) error {
	req := adapter.Request{
		Method: "create",
		Data: adapter.CreateRequest{
			Collection: mr.collection,
			Documents:  []interface{}{data},
		},
	}

	payload, err := jsoniter.Marshal(&req)
	if err != nil {
		return err
	}

	_, err = utils.SaiQuerySender(bytes.NewReader(payload), mr.address, mr.token)

	return err
}
