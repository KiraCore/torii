package mongo

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/saiset-co/sai-storage-mongo/logger"
	"github.com/saiset-co/sai-storage-mongo/types"
)

var DuplicateCh chan DuplicateRequest

type DuplicateData struct {
	Method    string      `json:"method"`
	Data      interface{} `json:"data"`
	Processed interface{} `json:"processed"`
}

type DuplicateRequest struct {
	Method string        `json:"method"`
	Data   DuplicateData `json:"data"`
}

func (c Client) Duplicate(method string, request types.IRequest, data interface{}) {
	if !c.Config.Duplicate {
		return
	}

	duplicateRequest := DuplicateRequest{
		Method: c.Config.DuplicateMethod,
		Data: DuplicateData{
			Method:    method,
			Data:      request,
			Processed: data,
		},
	}

	DuplicateCh <- duplicateRequest
}

func (c Client) DuplicateProcessor() {
	for {
		duplicateRequest := <-DuplicateCh

		go func() {
			dupRequestBytes, err := json.Marshal(duplicateRequest)
			if err != nil {
				logger.Logger.Error("DuplicateProcessor", zap.Error(err))
				return
			}

			req, err := http.NewRequest("POST", c.Config.DuplicateURL, bytes.NewBuffer(dupRequestBytes))
			if err != nil {
				logger.Logger.Error("DuplicateProcessor", zap.Error(err))
				return
			}

			client := http.Client{
				Timeout: time.Duration(c.Config.DuplicateTimeout) * time.Millisecond,
			}

			resp, err := client.Do(req)
			if err != nil {
				logger.Logger.Error("DuplicateProcessor", zap.Error(err))
				return
			}

			if resp.StatusCode != 200 {
				logger.Logger.Error("DuplicateProcessor", zap.Any("response", resp))
				return
			}

			time.Sleep(time.Duration(c.Config.DuplicatePause))
		}()
	}
}
