package actions

import (
	"maps"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/saiset-co/sai-storage-mongo/logger"
	"github.com/saiset-co/sai-storage-mongo/mongo"
	"github.com/saiset-co/sai-storage-mongo/types"
	"github.com/saiset-co/sai-storage-mongo/utils"
)

const Update = "update"

type UpdateAction struct {
	Client *mongo.Client
}

func NewUpdateAction(client *mongo.Client) *UpdateAction {
	return &UpdateAction{
		Client: client,
	}
}

func (action *UpdateAction) Handle(request types.IRequest) (interface{}, int, error) {
	precessed, err := action.process(request.GetData())
	if err != nil {
		logger.Logger.Error("UpdateAction", zap.Error(err))
		return nil, http.StatusInternalServerError, err
	}

	findResult, err := action.Client.Find(request.GetCollection(), request.GetSelect(), request.GetOptions(), request.GetIncludeFields())
	if err != nil {
		logger.Logger.Error("CreateAction", zap.Error(err))
		return nil, http.StatusInternalServerError, err
	}

	_, err = action.Client.Update(request.GetCollection(), request.GetSelect(), precessed)
	if err != nil {
		logger.Logger.Error("UpdateAction", zap.Error(err))
		return nil, http.StatusInternalServerError, err
	}

	for i, item := range findResult.Result {
		if itemData, itemOk := item.(map[string]interface{}); itemOk {
			if getData, dataOk := precessed.(map[string]interface{}); dataOk {
				if setValue, setOk := getData["$set"].(map[string]interface{}); setOk {
					maps.Copy(itemData, setValue)
					findResult.Result[i] = itemData
				}
				if unsetValue, setOk := getData["$unset"].(map[string]interface{}); setOk {
					utils.MapsDelete(itemData, unsetValue)
					findResult.Result[i] = itemData
				}
			}
		}
	}

	action.Client.Duplicate(Update, request, findResult)

	return findResult, http.StatusOK, nil
}

func (action *UpdateAction) process(data interface{}) (interface{}, error) {
	if itemData, ok := data.(map[string]interface{}); ok {
		if setValue, setOk := itemData["$set"].(map[string]interface{}); setOk {
			setValue["ch_time"] = time.Now().Unix()
			itemData["$set"] = setValue
			data = itemData
		}
		if unsetValue, setOk := itemData["$unset"].(map[string]interface{}); setOk {
			unsetValue["ch_time"] = time.Now().Unix()
			itemData["$unset"] = unsetValue
			data = itemData
		}
	}

	return data, nil
}
