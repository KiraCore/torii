package internal

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KiraCore/sekai-bridge/utils"
	"github.com/go-playground/validator"
	jsoniter "github.com/json-iterator/go"
	"github.com/saiset-co/sai-storage-mongo/external/adapter"
	"go.mongodb.org/mongo-driver/bson"
	"math/big"
	"net/http"
	"strings"

	"github.com/KiraCore/sekai-bridge/tss"
	"github.com/KiraCore/sekai-bridge/types"
	"github.com/binance-chain/tss-lib/common"
	"github.com/saiset-co/saiService"
	"go.uber.org/zap"
)

func (is *InternalService) NewHandler() saiService.Handler {
	return saiService.Handler{
		"keygen": saiService.HandlerElement{
			Name:        "keygen",
			Description: "Start threshold keys generation",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				tokenIsValid, err := is.validateToken(meta)
				if err != nil {
					return "", http.StatusInternalServerError, err
				}

				if !tokenIsValid {
					return "", http.StatusInternalServerError, errors.New("token doe not valid")
				}

				return is.keygen()
			},
		},
		"sign": saiService.HandlerElement{
			Name:        "sign",
			Description: "Sign the data",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				tokenIsValid, err := is.validateToken(meta)
				if err != nil {
					return "", http.StatusInternalServerError, err
				}

				if !tokenIsValid {
					return "", http.StatusInternalServerError, errors.New("token doe not valid")
				}

				return is.sign(data)
			},
		},
		"verify": saiService.HandlerElement{
			Name:        "verify",
			Description: "verify signature",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				tokenIsValid, err := is.validateToken(meta)
				if err != nil {
					return "", http.StatusInternalServerError, err
				}

				if !tokenIsValid {
					return "", http.StatusInternalServerError, errors.New("token doe not valid")
				}

				return is.verify(data)
			},
		},
		"stats": saiService.HandlerElement{
			Name:        "stats",
			Description: "p2p stats",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				tokenIsValid, err := is.validateToken(meta)
				if err != nil {
					return "", http.StatusInternalServerError, err
				}

				if !tokenIsValid {
					return "", http.StatusInternalServerError, errors.New("token doe not valid")
				}

				return is.stats()
			},
		},
		"notify": saiService.HandlerElement{
			Name:        "notify",
			Description: "Notification webhook",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				tokenIsValid, err := is.validateToken(meta)
				if err != nil {
					return "", http.StatusInternalServerError, err
				}

				if !tokenIsValid {
					return "", http.StatusInternalServerError, errors.New("token doe not valid")
				}

				return is.handleTransaction(data, meta)
			},
		},
		"logs": saiService.HandlerElement{
			Name:        "logs",
			Description: "Get logs from storage",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				tokenIsValid, err := is.validateToken(meta)
				if err != nil {
					return "", http.StatusInternalServerError, err
				}

				if !tokenIsValid {
					return "", http.StatusInternalServerError, errors.New("token doe not valid")
				}

				return is.logs(data, meta)
			},
		},
	}
}

// Keygen handler
func (is *InternalService) keygen() (interface{}, int, error) {
	response, err := is.Tss.Keygen(is.Tss.Parties, is.Tss.Threshold)
	if err != nil {
		return "keygen error", 500, err
	}

	return response.Key.ECDSAPub.Y(), 200, nil
}

func (is *InternalService) sign(data interface{}) (interface{}, int, error) {
	var request = tss.SignMessageRequest{
		OneRoundSigning: true,
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return "marshaling error", 500, err
	}

	err = json.Unmarshal(dataJSON, &request)
	if err != nil {
		return "un-marshaling error", 500, err
	}

	response, err := is.Tss.Sign(&request)
	if err != nil {
		return "sign error", 500, err
	}

	return response, 200, nil
}

func (is *InternalService) logs(data, meta interface{}) (interface{}, int, error) {
	var request logsRequest
	var err error

	response := struct {
		cosmos   interface{}
		ethereum interface{}
	}{}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return "marshaling error", 500, err
	}

	err = json.Unmarshal(dataJSON, &request)
	if err != nil {
		return "un-marshaling error", 500, err
	}

	selectData := bson.M{"$or": bson.A{
		bson.M{"From": request.Address},
		bson.M{"To": request.Address},
	}}

	cosmosStorageRequest := adapter.Request{
		Method: "read",
		Data: adapter.ReadRequest{
			Collection:    "Cosmos",
			Select:        selectData,
			Options:       &adapter.Options{},
			IncludeFields: []string{""},
		},
	}

	response.cosmos, err = is.Storage.Send(cosmosStorageRequest)
	if err != nil {
		return nil, 500, err
	}

	ethereumStorageRequest := adapter.Request{
		Method: "read",
		Data: adapter.ReadRequest{
			Collection:    "Cosmos",
			Select:        selectData,
			Options:       &adapter.Options{},
			IncludeFields: []string{""},
		},
	}

	response.ethereum, err = is.Storage.Send(ethereumStorageRequest)
	if err != nil {
		return nil, 500, err
	}

	return response, 200, nil
}

// stats handler (for debugging)
func (is *InternalService) stats() (interface{}, int, error) {
	stats := &types.SekaiBridgeStats{
		HTTPPort:             is.P2P.Config.P2P.Http.Port,
		P2PPort:              is.P2P.Config.P2P.Port,
		P2PSlot:              is.P2P.Config.P2P.Slot,
		P2pPeers:             is.P2P.Config.Peers,
		P2PConnectionStorage: is.P2P.ConnectionStorage,
		P2PSavedMessages:     make(map[string]bool),
		TssPartyID:           *is.Tss.LocalPartyID,
		TssConnectionStorage: is.Tss.ConnectionStorage,
		TssPartiesMap:        make(map[string]bool),
		TssKeygenMsgStorage:  make(map[string]bool),
	}
	if is.Tss.PartiesMap != nil {
		for id := range is.Tss.PartiesMap {
			stats.TssPartiesMap[id.Id+"|"+id.Moniker+"|"+fmt.Sprintf("index = %d", id.Index)] = true
		}
	}
	if is.Tss.KeygenInstance != nil {
		for _, msg := range is.Tss.KeygenInstance.KeygenMsgsStorage.M {
			stats.TssKeygenMsgStorage[msg.From.GetId()+"|"+msg.Type] = false
		}
	}

	it := is.P2P.Cache.Iterator()
	for it.SetNext() {
		entryInfo, err := it.Value()
		if err != nil {
			is.P2P.Logger.Error("p2p -> server -> GetStats -> it.Value", zap.Error(err))
			continue
		}
		stats.P2PSavedMessages[entryInfo.Key()] = true
	}

	return stats, 200, nil
}

func (is *InternalService) verify(data interface{}) (interface{}, int, error) {
	var request verifyRequest

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return "marshaling error", 500, err
	}

	err = json.Unmarshal(dataJSON, &request)
	if err != nil {
		return "un-marshaling error", 500, err
	}

	signatureData, err := base64.StdEncoding.DecodeString(request.Signature)
	if err != nil {
		return "DecodeString error", 500, err
	}

	signature := common.ECSignature{}

	err = json.Unmarshal(signatureData, &signature)
	if err != nil {
		return "Unmarshal error", 500, err
	}

	isValid := is.Tss.VerifySignature(&signature, request.Msg)

	return verifyResponse{Valid: isValid}, 200, nil
}

func (is *InternalService) handleTransaction(data, meta interface{}) (interface{}, int, error) {
	var request = new(NotificationRequest)

	dataJson, err := json.Marshal(data)
	if err != nil {
		return nil, 500, err
	}

	err = json.Unmarshal(dataJson, request)
	if err != nil {
		return nil, 500, err
	}

	err = validator.New().Struct(request)
	if err != nil {
		return nil, 500, err
	}

	signature, status, err := is.sign(request.Tx)
	if err != nil {
		return nil, status, err
	}

	saveStorageRequest := adapter.Request{
		Method: "create",
		Data:   adapter.CreateRequest{},
	}

	switch request.From {
	case "Cosmos":
		err = is.callEthContract(request.Tx, signature, meta)
		if err != nil {
			return nil, 500, err
		}

		saveStorageRequest.Data = adapter.CreateRequest{
			Collection:    "Cosmos",
			Documents:     []interface{}{request.Tx},
			Options:       &adapter.Options{},
			IncludeFields: []string{""},
		}

	case "Ethereum":
		err = is.callCosmosContract(request.Tx, signature, meta)
		if err != nil {
			return nil, 500, err
		}

		saveStorageRequest.Data = adapter.CreateRequest{
			Collection:    "Ethereum",
			Documents:     []interface{}{request.Tx},
			Options:       &adapter.Options{},
			IncludeFields: []string{""},
		}
	}

	_, err = is.Storage.Send(saveStorageRequest)
	if err != nil {
		return nil, 500, err
	}

	return data, 200, nil
}

func (is *InternalService) callEthContract(txData, signature, meta interface{}) error {
	var tx = new(CosmosTx)

	txJson, err := json.Marshal(txData)
	if err != nil {
		return err
	}

	err = json.Unmarshal(txJson, tx)
	if err != nil {
		return err
	}

	url := is.Context.GetConfig("interaction.ethereum", "").(string)

	newRequest := types.EthInteractionRequest{
		Method: "api",
		Data: types.EthInteractionData{
			Method:   "recordData",
			Contract: "Bridge",
			Value:    "0",
			Params: []types.EthInteractionParam{
				{Type: "address", Value: tx.Messages[0].To},
				{Type: "string", Value: strings.ToLower(tx.Messages[0].To[2:] + tx.Messages[0].Hash[:24])},
				{Type: "uint256", Value: tx.Messages[0].Amount[0].Amount},
			},
			Signature: signature.(*tss.SignMessageResponse).SignatureMarshalled,
		},
		Metadata: meta,
	}

	payload, err := jsoniter.Marshal(&newRequest)
	if err != nil {
		return err
	}

	_, err = utils.SaiQuerySender(bytes.NewReader(payload), url, "")

	return err
}

func (is *InternalService) callCosmosContract(txData, signature, meta interface{}) error {
	var tx = new(EthereumTx)

	txJson, err := json.Marshal(txData)
	if err != nil {
		return err
	}

	err = json.Unmarshal(txJson, tx)
	if err != nil {
		return err
	}

	url := is.Context.GetConfig("interaction.cosmos", "").(string)
	sekaiUrl := is.Context.GetConfig("sekai.url", "").(string)
	sekaiWallet := is.Context.GetConfig("sekai.wallet", "").(string)
	sekaiNetwork := is.Context.GetConfig("sekai.network", "").(string)
	sekaiGaslimit := is.Context.GetConfig("sekai.gas_limit", "").(int)
	sekaiFee := is.Context.GetConfig("sekai.fee", "").(int)

	newRequest := types.CosmosInteractionRequest{
		Method: "make_tx",
		Data: types.CosmosInteractionData{
			Type:        "bridge",
			NodeAddress: sekaiUrl,
			Sender:      sekaiWallet,
			From:        tx.From,
			To:          tx.Input.CyclAddress,
			ChainId:     sekaiNetwork,
			Memo:        "Bridge exchange",
			Amount:      tx.Input.Amount,
			GasLimit:    sekaiGaslimit,
			FeeAmount:   sekaiFee,
			Signature:   signature.(*tss.SignMessageResponse).SignatureMarshalled,
		},
		Metadata: meta,
	}

	payload, err := jsoniter.Marshal(&newRequest)
	if err != nil {
		return err
	}

	_, err = utils.SaiQuerySender(bytes.NewReader(payload), url, "")

	return err
}

type CosmosTx struct {
	Hash     string `json:"hash" validate:"required"`
	Height   string `json:"height" validate:"required"`
	Index    int    `json:"index"`
	Messages []struct {
		Amount []struct {
			Amount string `json:"amount" validate:"required"`
			Denom  string `json:"denom"`
		} `json:"amount"`
		From    string `json:"from"`
		Hash    string `json:"hash"`
		To      string `json:"to"`
		TypeUrl string `json:"typeUrl"`
	} `json:"messages" validate:"required"`
	Tx       string `json:"tx" validate:"required"`
	TxResult struct {
		Code      int    `json:"code"`
		Codespace string `json:"codespace"`
		Data      string `json:"data"`
		Events    []struct {
			Attributes []struct {
				Index bool   `json:"index"`
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"attributes"`
			Type string `json:"type"`
		} `json:"events"`
		GasUsed   string `json:"gas_used"`
		GasWanted string `json:"gas_wanted"`
		Info      string `json:"info"`
		Log       string `json:"log"`
	} `json:"tx_result"`
}

type NotificationRequest struct {
	From string      `json:"from"`
	Tx   interface{} `json:"tx"`
}

type verifyRequest struct {
	Msg       string `json:"msg"`
	Signature string `json:"signature"`
}

type verifyResponse struct {
	Valid bool `json:"is_valid"`
}

type logsRequest struct {
	Address string `json:"address"`
}

type EthereumTx struct {
	Amount big.Int `json:"Amount"`
	From   string  `json:"From"`
	Hash   string  `json:"Hash"`
	Input  struct {
		Amount      int    `json:"amount"`
		CyclAddress string `json:"cyclAddress"`
		Hash        string `json:"hash"`
	} `json:"Input"`
	Number    int    `json:"Number"`
	Status    bool   `json:"Status"`
	To        string `json:"To"`
	Signature string `json:"Signature"`
}

func (is *InternalService) getToken(meta interface{}) (string, error) {
	metaMap, ok := meta.(map[string]interface{})
	if !ok {
		return "", errors.New("wrong metadata format")
	}

	token, ok := metaMap["token"].(string)
	if !ok {
		return "", errors.New("token does not found")
	}

	return token, nil
}

func (is *InternalService) validateToken(meta interface{}) (bool, error) {
	token, err := is.getToken(meta)
	if err != nil {
		return false, err
	}

	return token == is.Context.GetConfig("token", "").(string), nil
}
