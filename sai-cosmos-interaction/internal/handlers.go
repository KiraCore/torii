package internal

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/saiset-co/saiCosmosInteraction/internal/model"
	"github.com/saiset-co/saiCosmosInteraction/utils"
	"github.com/saiset-co/saiService"
)

func (is *InternalService) NewHandler() saiService.Handler {
	return saiService.Handler{
		"make_tx": saiService.HandlerElement{
			Name:        "make tx",
			Description: "Make new transaction with type /cosmos.bank.v1beta1.MsgSend",
			Function:    is.makeTx,
		},
		"make_tx_signed": saiService.HandlerElement{
			Name:        "make tx signed",
			Description: "Make new transaction with type /cosmos.bank.v1beta1.MsgSend",
			Function:    is.makeTxSigned,
		},
	}
}

func (is *InternalService) makeTx(data, meta interface{}) (interface{}, int, error) {
	tokenIsValid, err := is.validateToken(meta)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	if !tokenIsValid {
		return "", http.StatusInternalServerError, errors.New("token doe not valid")
	}

	body, err := is.validateBody(data)
	if err != nil {
		return "", http.StatusBadRequest, err
	}

	fileBytes, err := os.ReadFile(body.Sender)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("don't have private key for %s", body.From)
	}

	txMaker, err := NewTransactionMaker(
		body.Sender,
		body.NodeAddress,
		body.ChainID,
		body.From,
		body.To,
		body.Signature,
		fileBytes,
	)

	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	err = txMaker.BuildTx(uint64(body.GasLimit), body.Amount, body.FeeAmount, body.Memo)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	err = txMaker.SignTx()
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	txHash, err := txMaker.BroadcastTx()
	if err != nil {
		//logger.Logger.Error(body.From, zap.Error(err))
		return "", http.StatusInternalServerError, err
	}

	return txHash, http.StatusOK, nil
}

func (is *InternalService) makeTxSigned(data, meta interface{}) (interface{}, int, error) {
	body, err := is.validateBody(data)
	if err != nil {
		return "", http.StatusBadRequest, err
	}

	fileBytes, err := os.ReadFile(body.Sender)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("don't have private key for %s", body.From)
	}

	txMaker, err := NewTransactionMaker(
		body.Sender,
		body.NodeAddress,
		body.ChainID,
		body.From,
		body.To,
		body.Signature,
		fileBytes,
	)

	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	err = txMaker.BuildTx(uint64(body.GasLimit), body.Amount, body.FeeAmount, body.Memo)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	err = txMaker.SignTx()
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	txHash, err := txMaker.BroadcastTx()
	if err != nil {
		//logger.Logger.Error(body.From, zap.Error(err))
		return "", http.StatusInternalServerError, err
	}

	return txHash, http.StatusOK, nil
}

func (is *InternalService) validateBody(data interface{}) (model.MakeTxRequestBody, error) {
	body := model.MakeTxRequestBody{}
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return body, fmt.Errorf("wrong request body")
	}

	body.NodeAddress, ok = dataMap["node_address"].(string)
	if !ok {
		return body, fmt.Errorf("node_address field not string")
	}

	body.Sender, ok = dataMap["sender"].(string)
	if !ok {
		return body, fmt.Errorf("sender field not string")
	}

	body.From, ok = dataMap["from"].(string)
	if !ok {
		return body, fmt.Errorf("from field not string")
	}

	body.To, ok = dataMap["to"].(string)
	if !ok {
		return body, fmt.Errorf("to field not string")
	}

	body.ChainID, ok = dataMap["chain_id"].(string)
	if !ok {
		return body, fmt.Errorf("chain_id field not string")
	}

	body.Signature, ok = dataMap["signature"].(string)
	if !ok {
		return body, fmt.Errorf("signature field not string")
	}

	var err error
	body.Amount, err = utils.IfaceToInt64(dataMap["amount"])
	if err != nil {
		return body, fmt.Errorf("amount field not int64")
	}

	body.GasLimit, err = utils.IfaceToInt64(dataMap["gas_limit"])
	if err != nil {
		return body, fmt.Errorf("gas_limit field not int64")
	}

	body.FeeAmount, err = utils.IfaceToInt64(dataMap["fee_amount"])
	if err != nil {
		return body, fmt.Errorf("fee_amount field not int64")
	}

	return body, nil
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
