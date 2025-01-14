package internal

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/saiset-co/sai-eth-interaction/models"
	"go.uber.org/zap"
)

var mux sync.Mutex
var nonceList = map[string]map[uint64]bool{}

type resTX struct {
	Transaction *types.Transaction
	Status      string `json:"status"`
	Result      string `json:"result"`
}

func response(tx *types.Transaction, res string) (resTX, error) {
	result := resTX{
		Transaction: tx,
		Status:      "error",
		Result:      res,
	}
	resultS, _ := json.Marshal(result)
	err := errors.New(string(resultS))

	return result, err
}

func (is *InternalService) RawTransaction(client *ethclient.Client, value *big.Int, data []byte, contract *models.Contract) (string, error) {
	d := time.Now().Add(20 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()

	privateKey, err := crypto.HexToECDSA(contract.Private)
	if err != nil {
		is.Logger.Error("handlers - api - RawTransaction - HexToECDSA", zap.Error(err))
		return "", err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		is.Logger.Error("handlers - api - RawTransaction - cast publicKey to ecdsa", zap.Error(err))
		return "", err
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	s := fromAddress.String()
	is.Logger.Debug("1", zap.String("", s))
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return "", err
	}

	toAddress := common.HexToAddress(contract.Address)

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		is.Logger.Error("handlers - api - RawTransaction - get suggested gas price", zap.Error(err))
		return "", err
	}

	is.Logger.Sugar().Debugf("GAS PRICE : %v", gasPrice)

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &toAddress,
		Value:    value,
		Gas:      contract.GasLimit,
		GasPrice: gasPrice,
		Data:     data,
	})

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		res, err := response(tx, err.Error())
		is.Logger.Error("handlers - api - RawTransaction - get networkID", zap.Any("TX", res))
		return "", err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		res, err := response(tx, err.Error())
		is.Logger.Error("handlers - api - RawTransaction - signTx", zap.Any("TX", res))
		return "", err
	}

	hash := signedTx.Hash().String()

	mux.Lock()
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		res, err := response(tx, err.Error())
		mux.Unlock()
		is.Logger.Error("handlers - api - RawTransaction - sendTx", zap.Any("TX", res))
		return "", err
	}

	res, _ := response(tx, hash)
	//resS, _ := json.Marshal(res)

	for {
		resTx, _, err := client.TransactionByHash(ctx, signedTx.Hash())
		if err != nil {
			res, err := response(tx, err.Error())
			is.Logger.Error("handlers - api - RawTransaction - sendTx", zap.Any("TX", res))
			mux.Unlock()
			return "", err
		}

		if resTx == nil {
			is.Logger.Debug("handlers - api - RawTransaction - sendTx - tx was not created", zap.Any("TX", res))
			mux.Unlock()
			goto done
		} else if resTx != nil {
			is.Logger.Debug("handlers - api - RawTransaction - sendTx - tx done", zap.Any("TX", res))
			mux.Unlock()
			goto done
		}

		is.Logger.Debug("handlers - api - RawTransaction - sendTx - pending, sleep 2 sec")
		time.Sleep(2 * time.Second)
	}

done:
	return hash, nil
}
