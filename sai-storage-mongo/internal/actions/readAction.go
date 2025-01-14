package actions

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/saiset-co/sai-storage-mongo/logger"
	"github.com/saiset-co/sai-storage-mongo/mongo"
	"github.com/saiset-co/sai-storage-mongo/types"
)

const Read = "read"

type ReadAction struct {
	Client *mongo.Client
}

func NewGetAction(client *mongo.Client) *ReadAction {
	return &ReadAction{
		Client: client,
	}
}

func (action *ReadAction) Handle(request types.IRequest) (interface{}, int, error) {
	data, err := action.Client.Find(request.GetCollection(), request.GetSelect(), request.GetOptions(), request.GetIncludeFields())
	if err != nil {
		logger.Logger.Error("ReadAction", zap.Error(err))
		return nil, http.StatusInternalServerError, err
	}

	action.Client.Duplicate(Read, request, data.Result)

	return data, http.StatusOK, nil
}
