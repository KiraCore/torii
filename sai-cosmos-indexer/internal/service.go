package internal

import (
	"bytes"
	"errors"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/saiset-co/sai-storage-mongo/external/adapter"
	"github.com/saiset-co/saiCosmosIndexer/internal/model"
	"github.com/saiset-co/saiCosmosIndexer/utils"
	"github.com/saiset-co/saiService"
	"github.com/spf13/cast"
)

const (
	filePathAddresses   = "./addresses.json"
	filePathLatestBlock = "./latest_handled_block"
)

type InternalService struct {
	mu            *sync.Mutex
	Context       *saiService.Context
	config        model.ServiceConfig
	currentBlock  int64
	addresses     map[string]struct{}
	storageConfig model.StorageConfig
	notifier      Notifier
	client        http.Client
}

func (is *InternalService) Init() {
	is.mu = &sync.Mutex{}
	is.config = model.ServiceConfig{}
	is.client = http.Client{}

	is.addresses = make(map[string]struct{})
	is.config.TxType = cast.ToString(is.Context.GetConfig("tx_type", ""))
	is.config.NodeAddress = cast.ToString(is.Context.GetConfig("node_address", ""))
	is.config.CollectionName = cast.ToString(is.Context.GetConfig("storage.mongo_collection_name", ""))
	is.config.SkipFailedTxs = cast.ToBool(is.Context.GetConfig("skip_failed_tx", false))
	is.storageConfig = model.StorageConfig{
		Token:      cast.ToString(is.Context.GetConfig("storage.token", "")),
		Url:        cast.ToString(is.Context.GetConfig("storage.url", "")),
		Email:      cast.ToString(is.Context.GetConfig("storage.email", "")),
		Password:   cast.ToString(is.Context.GetConfig("storage.password", "")),
		Collection: cast.ToString(is.Context.GetConfig("storage.mongo_collection_name", "")),
	}

	is.notifier = NewNotifier(
		cast.ToString(is.Context.GetConfig("notifier.sender_id", "")),
		cast.ToString(is.Context.GetConfig("notifier.email", "")),
		cast.ToString(is.Context.GetConfig("notifier.password", "")),
		cast.ToString(is.Context.GetConfig("notifier.token", "")),
		cast.ToString(is.Context.GetConfig("notifier.url", "")),
	)

	err := is.loadAddresses()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		//logger.Logger.Error("loadAddresses", zap.Error(err))
	}

	fileBytes, err := os.ReadFile(filePathLatestBlock)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			//logger.Logger.Error("can't read "+filePathLatestBlock, zap.Error(err))
		}
	} else {
		latestHandledBlock, err := strconv.Atoi(string(fileBytes))
		if err != nil {
			//logger.Logger.Error("strconv.Atoi", zap.Error(err))
		}

		is.currentBlock = int64(latestHandledBlock)
	}

	startBlock := cast.ToInt64(is.Context.GetConfig("start_block", 0))
	if is.currentBlock < startBlock {
		is.currentBlock = startBlock
	}
}

func (is *InternalService) Process() {
	sleepDuration := cast.ToDuration(is.Context.GetConfig("sleep_duration", 2))

	for {
		select {
		case <-is.Context.Context.Done():
			//logger.Logger.Debug("saiCosmosIndexer loop is done")
			return
		default:
			//if len(is.addresses) == 0 {
			//	time.Sleep(time.Second * sleepDuration)
			//	continue
			//}

			latestBlockHeight, err := is.getLatestBlock()
			if err != nil {
				//logger.Logger.Error("getLatestBlock", zap.Error(err))
				time.Sleep(time.Second * sleepDuration)
				continue
			}

			if is.currentBlock >= latestBlockHeight {
				time.Sleep(time.Second * sleepDuration)
				continue
			}

			err = is.handleBlockTxs()
			if err != nil {
				//logger.Logger.Error("handleBlockTxs", zap.Error(err))
				time.Sleep(time.Second * sleepDuration)
				continue
			}

			is.currentBlock += 1
		}
	}
}

func (is *InternalService) handleBlockTxs() error {
	blockTxs, err := is.getBlockTxs()
	if err != nil {
		return err
	}

	var txArray []interface{}
	for _, txRes := range blockTxs {
		if is.config.SkipFailedTxs && txRes.Code != 0 {
			continue
		}

		if len(txRes.Tx.Body.Messages) < 1 {
			continue
		}

		to := txRes.Tx.Body.Messages[0].ToAddress
		from := txRes.Tx.Body.Messages[0].FromAddress
		amount := txRes.Tx.Body.Messages[0].Amount[0].Amount
		hash := txRes.Tx.Body.Messages[0].Hash

		//_, isReceiver := is.addresses[to]
		//_, isSender := is.addresses[from]
		//if !isReceiver && !isSender {
		//	continue
		//}

		txTmp := map[string]interface{}{
			"Number": txRes.Height,
			"Hash":   hash,
			"From":   from,
			"To":     to,
			"Amount": amount,
			"Events": txRes.Events,
			"Status": txRes.Code,
			"Ts":     txRes.Timestamp,
		}

		txArray = append(txArray, txTmp)

		go is.sendTxNotification(txTmp)
	}

	//err = is.sendTxsToStorage(txArray)
	//if err != nil {
	//	return err
	//}

	err = is.rewriteLastHandledBlock(is.currentBlock)

	return err
}

func (is *InternalService) sendTxsToStorage(txs []interface{}) error {
	storageRequest := adapter.Request{
		Method: "create",
		Data: adapter.CreateRequest{
			Collection: is.storageConfig.Collection,
			Documents:  txs,
		},
	}

	bodyBytes, err := jsoniter.Marshal(&storageRequest)
	if err != nil {
		return err
	}

	_, err = utils.SaiQuerySender(bytes.NewBuffer(bodyBytes), is.storageConfig.Url, is.storageConfig.Token)

	return err
}

func (is *InternalService) sendTxNotification(tx interface{}) {
	err := is.notifier.SendTx(tx)
	if err != nil {
		//logger.Logger.Error("is.notifier.SendTx", zap.Error(err))
	}
}
