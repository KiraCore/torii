package actions

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/saiset-co/sai-storage-mongo/logger"
	"github.com/saiset-co/sai-storage-mongo/mongo"
	"github.com/saiset-co/sai-storage-mongo/types"
)

const Delete = "delete"

type DeleteAction struct {
	Client *mongo.Client
}

func NewDeleteAction(client *mongo.Client) *DeleteAction {
	return &DeleteAction{
		Client: client,
	}
}

func (action *DeleteAction) Handle(request types.IRequest) (interface{}, int, error) {
	err := action.Client.Remove(request.GetCollection(), request.GetSelect())
	if err != nil {
		logger.Logger.Error("DeleteAction", zap.Error(err))
		return nil, http.StatusInternalServerError, err
	}

	action.Client.Duplicate(Delete, request, []interface{}{})

	return "Documents have been deleted", http.StatusOK, nil
}
