package actions

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/saiset-co/sai-storage-mongo/logger"
	"github.com/saiset-co/sai-storage-mongo/mongo"
	"github.com/saiset-co/sai-storage-mongo/types"
)

type CreateIndexesAction struct {
	Client *mongo.Client
}

func NewCreateIndexesAction(client *mongo.Client) *CreateIndexesAction {
	return &CreateIndexesAction{
		Client: client,
	}
}

func (action *CreateIndexesAction) Handle(request types.IRequest) (interface{}, int, error) {
	result, err := action.Client.CreateIndexes(request.GetCollection(), request.GetData())
	if err != nil {
		logger.Logger.Error("CreateIndexAction", zap.Error(err))
		return nil, http.StatusInternalServerError, err
	}

	return result, http.StatusOK, nil
}
