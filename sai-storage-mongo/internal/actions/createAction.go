package actions

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"

	"github.com/saiset-co/sai-storage-mongo/logger"
	"github.com/saiset-co/sai-storage-mongo/mongo"
	"github.com/saiset-co/sai-storage-mongo/types"
)

const Create = "create"

type CreateAction struct {
	Client *mongo.Client
}

func NewSaveAction(client *mongo.Client) *CreateAction {
	return &CreateAction{
		Client: client,
	}
}

func (action *CreateAction) Handle(request types.IRequest) (interface{}, int, error) {
	precessed, err := action.process(request.GetData())
	if err != nil {
		logger.Logger.Error("CreateAction", zap.Error(err))
		return nil, http.StatusInternalServerError, err
	}

	insertResult, err := action.Client.InsertMany(request.GetCollection(), precessed)
	if err != nil {
		logger.Logger.Error("CreateAction", zap.Error(err))
		return nil, http.StatusInternalServerError, err
	}

	result, err := action.Client.Find(request.GetCollection(), bson.M{"_id": bson.M{"$in": insertResult.InsertedIDs}}, request.GetOptions(), request.GetIncludeFields())
	if err != nil {
		logger.Logger.Error("CreateAction", zap.Error(err))
		return nil, http.StatusInternalServerError, err
	}

	action.Client.Duplicate(Create, request, result)

	return result, http.StatusOK, nil
}

func (action *CreateAction) process(data interface{}) ([]interface{}, error) {
	processedData := data.([]interface{})

	for i, item := range processedData {
		if itemData, ok := item.(map[string]interface{}); ok {
			id, ok := itemData["internal_id"].(string)
			if !ok || id == "" {
				id = uuid.New().String()
			}

			itemData["internal_id"] = id
			itemData["cr_time"] = time.Now().Unix()
			itemData["ch_time"] = time.Now().Unix()

			processedData[i] = itemData
		}
	}
	return processedData, nil
}
