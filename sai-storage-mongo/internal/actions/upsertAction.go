package actions

import (
	"maps"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"

	"github.com/saiset-co/sai-storage-mongo/logger"
	"github.com/saiset-co/sai-storage-mongo/mongo"
	"github.com/saiset-co/sai-storage-mongo/types"
	"github.com/saiset-co/sai-storage-mongo/utils"
)

const (
	UpsertUpdate = "upsert_update"
	UpsertCreate = "upsert_create"
)

type UpsertAction struct {
	Client *mongo.Client
}

func NewUpsertAction(client *mongo.Client) *UpsertAction {
	return &UpsertAction{
		Client: client,
	}
}

func (action *UpsertAction) Handle(request types.IRequest) (interface{}, int, error) {
	exists, err := action.Client.Find(request.GetCollection(), request.GetSelect(), request.GetOptions(), []string{"_id"})
	if err != nil {
		logger.Logger.Error("UpsertAction", zap.Error(err))
		return nil, http.StatusInternalServerError, err
	}

	if len(exists.Result) > 0 {
		precessed, err := action.processUpdate(request.GetData())
		if err != nil {
			logger.Logger.Error("UpsertAction", zap.Error(err))
			return nil, http.StatusInternalServerError, err
		}

		findResult, err := action.Client.Find(request.GetCollection(), request.GetSelect(), request.GetOptions(), request.GetIncludeFields())
		if err != nil {
			logger.Logger.Error("CreateAction", zap.Error(err))
			return nil, http.StatusInternalServerError, err
		}

		_, err = action.Client.Update(request.GetCollection(), request.GetSelect(), precessed)
		if err != nil {
			logger.Logger.Error("UpsertAction", zap.Error(err))
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

		action.Client.Duplicate(UpsertUpdate, request, findResult)

		return findResult, http.StatusOK, nil
	} else {
		precessed, err := action.processInsert(request.GetData())
		if err != nil {
			logger.Logger.Error("UpsertAction", zap.Error(err))
			return nil, http.StatusInternalServerError, err
		}

		insertResult, err := action.Client.Insert(request.GetCollection(), precessed)
		if err != nil {
			logger.Logger.Error("UpsertAction", zap.Error(err))
			return nil, http.StatusInternalServerError, err
		}

		result, err := action.Client.Find(request.GetCollection(), bson.M{"_id": insertResult.InsertedID}, request.GetOptions(), request.GetIncludeFields())
		if err != nil {
			logger.Logger.Error("CreateAction", zap.Error(err))
			return nil, http.StatusInternalServerError, err
		}

		action.Client.Duplicate(UpsertCreate, request, result)

		return result, http.StatusOK, nil
	}
}

func (action *UpsertAction) processUpdate(data interface{}) (interface{}, error) {
	if itemData, ok := data.(map[string]interface{}); ok {
		itemData["ch_time"] = time.Now().Unix()
		data = itemData
	}

	return data, nil
}

func (action *UpsertAction) processInsert(data interface{}) (interface{}, error) {
	if itemData, ok := data.(map[string]interface{}); ok {
		id, ok := itemData["internal_id"].(string)
		if !ok || id == "" {
			id = uuid.New().String()
		}

		itemData["internal_id"] = id
		itemData["cr_time"] = time.Now().Unix()
		itemData["ch_time"] = time.Now().Unix()

		data = itemData
	}

	return data, nil
}
