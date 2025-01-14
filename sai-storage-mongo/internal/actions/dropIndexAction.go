package actions

import (
	"net/http"

	"github.com/saiset-co/sai-storage-mongo/logger"
	"github.com/saiset-co/sai-storage-mongo/mongo"
	"github.com/saiset-co/sai-storage-mongo/types"
	"go.uber.org/zap"
)

type DropIndexesAction struct {
	Client *mongo.Client
}

func NewDropIndexesAction(client *mongo.Client) *DropIndexesAction {
	return &DropIndexesAction{
		Client: client,
	}
}

func (action *DropIndexesAction) Handle(request types.IRequest) (interface{}, int, error) {
	result, err := action.Client.DropIndexes(request.GetCollection())
	if err != nil {
		logger.Logger.Error("DropIndexesAction", zap.Error(err))
		return nil, http.StatusInternalServerError, err
	}

	return result, http.StatusOK, nil
}
