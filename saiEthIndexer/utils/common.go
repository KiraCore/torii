package utils

import (
	"encoding/json"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"io"
	"net/http"
	"reflect"

	"github.com/saiset-co/saiEthIndexer/config"
)

func InArray(val interface{}, array interface{}) (index int) {
	values := reflect.ValueOf(array)

	if reflect.TypeOf(array).Kind() == reflect.Slice || values.Len() > 0 {
		for i := 0; i < values.Len(); i++ {
			if reflect.DeepEqual(val, values.Index(i).Interface()) {
				return i
			}
		}
	}

	return -1
}

func ConvertInterfaceToJson(obj interface{}) []byte {
	jsonResult, _ := json.Marshal(obj)
	return jsonResult
}

func RemoveContract(slice []config.Contract, s int) []config.Contract {
	return append(slice[:s], slice[s+1:]...)
}

func RemoveAddress(slice []string, s int) []string {
	return append(slice[:s], slice[s+1:]...)
}

func SaiQuerySender(body io.Reader, address, token string) ([]byte, error) {
	const failedResponseStatus = "NOK"

	type responseWrapper struct {
		Status string              `json:"Status"`
		Error  string              `json:"Error"`
		Result jsoniter.RawMessage `json:"result"`
		Count  int                 `json:"count"`
	}

	req, err := http.NewRequest(http.MethodPost, address, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Token", token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s", resBytes)
	}

	result := responseWrapper{}
	err = jsoniter.Unmarshal(resBytes, &result)
	if err != nil {
		return nil, err
	}

	if result.Status == failedResponseStatus {
		return nil, fmt.Errorf(result.Error)
	}

	return resBytes, nil
}
