package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/saiset-co/sai-eth-interaction/models"
	saiService "github.com/saiset-co/sai-service/service"
	"go.uber.org/zap"
)

const (
	contractsPath = "contracts.json"
)

func (is *InternalService) NewHandler() saiService.Handler {
	return saiService.Handler{
		"api": saiService.HandlerElement{
			Name:        "api",
			Description: "transact encoded transaction to contract by ABI",
			Function: func(data interface{}, metadata interface{}) (interface{}, int, error) {
				tokenIsValid, err := is.validateToken(metadata)
				if err != nil {
					return "", http.StatusInternalServerError, err
				}

				if !tokenIsValid {
					return "", http.StatusInternalServerError, errors.New("token doe not valid")
				}

				contractData, ok := data.(map[string]interface{})
				if !ok {
					Service.Logger.Sugar().Debugf("handling connect method, wrong type, current type : %+v", reflect.TypeOf(data))
					return nil, 400, errors.New("wrong type of incoming data")
				}

				b, err := json.Marshal(contractData)
				if err != nil {
					Service.Logger.Error("handlers - api - marshal incoming data", zap.Error(err))
					return nil, 400, err
				}

				req := models.EthRequest{}
				err = json.Unmarshal(b, &req)
				if err != nil {
					Service.Logger.Error("handlers - api - unmarshal data to struct", zap.Error(err))
					return nil, 400, err
				}

				contract, err := Service.GetContractByName(req.Contract)
				if err != nil {
					Service.Logger.Error("handlers - api - GetContractByName", zap.Error(err))
					return nil, 500, err
				}

				abiEl, err := abi.JSON(strings.NewReader(contract.ABI))
				if err != nil {
					log.Fatalf("Could not read ABI: %v", err)
					return nil, 500, err
				}

				ethURL := contract.Server
				if !ok {
					Service.Logger.Sugar().Fatalf("wrong type of eth_server value in config, type : %+v", reflect.TypeOf(Service.Context.GetConfig("specific.eth_server", "")))
				}

				ethClient, err := ethclient.Dial(ethURL)
				if err != nil {
					Service.Logger.Error("handlers - api - dial eth server", zap.Error(err))
					return nil, 500, err
				}

				var args []interface{}
				for _, v := range req.Params {
					arg := v.Value

					switch arg.(type) {
					case float64:
						Service.Logger.Error("handlers - api - wrong value format, please use strings always: 'value': '1'")
						return nil, 500, errors.New("handlers - api - wrong value format, please use strings always: 'value': '1'")
					case []float64:
						Service.Logger.Error("handlers - api - wrong value format, please use strings always: 'value': ['1']")
						return nil, 500, errors.New("handlers - api - wrong value format, please use strings always: 'value': ['1']")
					}

					if v.Type == "address" {
						arg = common.HexToAddress(v.Value.(string))
					}

					if v.Type == "uint16" {
						num, err := strconv.ParseUint(v.Value.(string), 10, 16)
						if err != nil {
							Service.Logger.Error("handlers - api - can't convert to uint16")
							return nil, 500, errors.New("handlers - api - can't convert to uint16")
						}
						arg = uint16(num)
					}

					if v.Type == "uint8" {
						num, err := strconv.ParseUint(v.Value.(string), 10, 8)
						if err != nil {
							Service.Logger.Error("handlers - api - can't convert to uint8")
							return nil, 500, errors.New("handlers - api - can't convert to uint8")
						}
						arg = uint8(num)
					}

					if v.Type == "uint256" {
						arg, ok = new(big.Int).SetString(v.Value.(string), 10)
						if !ok {
							Service.Logger.Error("handlers - api - can't convert to bigInt")
							return nil, 500, errors.New("handlers - api - can't convert to bigInt")
						}
					}

					if v.Type == "address[]" {
						t := v.Value.([]interface{})
						s := make([]common.Address, len(t))
						for i, a := range t {
							s[i] = common.HexToAddress(a.(string))
						}
						arg = s
					}

					if v.Type == "string[]" {
						t := v.Value.([]interface{})
						s := make([]string, len(t))
						for i, a := range t {
							s[i] = fmt.Sprint(a)
						}
						arg = s
					}

					if v.Type == "uint256[]" {
						t := v.Value.([]interface{})
						s := make([]*big.Int, len(t))
						for i, a := range t {
							s[i], ok = new(big.Int).SetString(a.(string), 10)
							if !ok {
								Service.Logger.Error("handlers - api - can't convert to bigInt uint256[]")
								return nil, 500, errors.New("handlers - api - can't convert to bigInt uint256[]")
							}
						}
						arg = s
					}

					if v.Type == "uint16[]" {
						t := v.Value.([]interface{})
						s := make([]uint16, len(t))
						for i, a := range t {
							num, err := strconv.ParseUint(a.(string), 10, 16)
							if err != nil {
								Service.Logger.Error("handlers - api - can't convert to uint16 uint16[]")
								return nil, 500, errors.New("handlers - api - can't convert to uint16 uint16[]")
							}
							s[i] = uint16(num)
						}
						arg = s
					}

					if v.Type == "uint8[]" {
						t := v.Value.([]interface{})
						s := make([]uint8, len(t))
						for i, a := range t {
							num, err := strconv.ParseUint(a.(string), 10, 8)
							if err != nil {
								Service.Logger.Error("handlers - api - can't convert to uint8 uint8[]")
								return nil, 500, errors.New("handlers - api - can't convert to uint8 uint8[]")
							}
							s[i] = uint8(num)
						}
						arg = s
					}

					args = append(args, arg)

					Service.Logger.Info("handlers - api", zap.Any("args", args))
				}

				input, err := abiEl.Pack(req.Method, args...)

				if err != nil {
					Service.Logger.Error("handlers - api - pack eth server", zap.Error(err))
					return nil, 500, err
				}

				value := big.NewInt(0)
				if req.Value != "" {
					value, ok = new(big.Int).SetString(req.Value, 10)
					if !ok {
						Service.Logger.Error("handlers - api - can't convert value to bigInt")
						return nil, 500, errors.New("handlers - api - can't convert value `to bigInt")
					}
				}

				res, err := Service.RawTransaction(ethClient, value, input, contract)
				if err != nil {
					return nil, 500, err
				}

				return map[string]interface{}{
					"Status": "OK",
					"Result": res,
				}, 200, nil
			},
		},

		"add": saiService.HandlerElement{
			Name:        "add",
			Description: "add contract to contracts",
			Function: func(data interface{}, metadata interface{}) (interface{}, int, error) {
				contractData, ok := data.(map[string]interface{})
				if !ok {
					Service.Logger.Sugar().Debugf("handlers - add - wrong data type, current type : %+v", data)
					return nil, 400, errors.New("wrong type of incoming data")
				}

				b, err := json.Marshal(contractData)
				if err != nil {
					Service.Logger.Error("api - add - marshal incoming data", zap.Error(err))
					return nil, 400, err
				}

				contracts := models.Contracts{}
				err = json.Unmarshal(b, &contracts)
				if err != nil {
					Service.Logger.Error("handlers - add - unmarshal data to struct", zap.Error(err))
					return nil, 400, err
				}

				// validate all incoming contracts
				validatedContracts := make([]models.Contract, 0)
				for _, contract := range contracts.Contracts {
					err = contract.Validate()
					if err != nil {
						Service.Logger.Error("handlers - add - validate incoming contracts", zap.Any("contract", contract), zap.Error(err))
						continue
					}
					validatedContracts = append(validatedContracts, contract)
				}

				// check if incoming contracts already exists
				Service.Mutex.RLock()
				checkedContracts := Service.filterUniqueContracts(validatedContracts)
				Service.Mutex.RUnlock()

				Service.Mutex.Lock()
				Service.Contracts = append(Service.Contracts, checkedContracts...)
				Service.Mutex.Unlock()

				Service.Logger.Sugar().Debugf("ACTUAL CONTRACTS : %+v", Service.Contracts)

				err = Service.RewriteContractsConfig(contractsPath)
				if err != nil {
					Service.Logger.Error("handlers - add - rewrite contracts file", zap.Error(err))
					return nil, 500, err
				}
				return "ok", 200, nil

			},
		},

		"delete": saiService.HandlerElement{
			Name:        "delete",
			Description: "delete contract by name",
			Function: func(data interface{}, metadata interface{}) (interface{}, int, error) {
				deleteData, ok := data.(map[string]interface{})
				if !ok {
					Service.Logger.Sugar().Debugf("handlers - delete - wrong data type, current type : %+v", data)
					return nil, 400, errors.New("wrong type of incoming data")
				}

				b, err := json.Marshal(deleteData)
				if err != nil {
					Service.Logger.Error("api - delete - marshal incoming data", zap.Error(err))
					return nil, 400, err
				}

				deleteContractName := models.DeleteData{}
				err = json.Unmarshal(b, &deleteContractName)
				if err != nil {
					Service.Logger.Error("handlers - add - unmarshal data to struct", zap.Error(err))
					return nil, 400, err
				}

				Service.Mutex.Lock()
				Service.DeleteContracts(&deleteContractName)
				Service.Mutex.Unlock()

				Service.Logger.Sugar().Debugf("CONTRACTS AFTER DELETION : %+v", Service.Contracts)

				err = Service.RewriteContractsConfig(contractsPath)
				if err != nil {
					Service.Logger.Error("handlers - delete - rewrite contracts file", zap.Error(err))
					return nil, 500, err
				}
				return "ok", 200, nil

			},
		},
	}

}

func (is *InternalService) validateToken(meta interface{}) (bool, error) {
	metaMap, ok := meta.(map[string]interface{})
	if !ok {
		return false, errors.New("wrong metadata format")
	}

	token, ok := metaMap["token"]
	if !ok {
		return false, errors.New("token does not found")
	}

	return token == is.Context.GetConfig("token", "").(string), nil
}
