package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cast"
	"go.uber.org/zap"

	"github.com/saiset-co/saiCosmosIndexer/internal/model"
	"github.com/saiset-co/saiCosmosIndexer/logger"
)

func (is *InternalService) getLatestBlock() (*model.LatestBlock, error) {
	res, err := is.makeTendermintRPCRequest("/blockchain", "max_height=1")
	if err != nil {
		logger.Logger.Error("getLatestBlock", zap.Error(err))
		return nil, err
	}

	lb := new(model.LatestBlock)
	err = jsoniter.Unmarshal(res, lb)
	if err != nil {
		logger.Logger.Error("getLatestBlock", zap.Error(err))
		return nil, err
	}

	return lb, err
}

func (is *InternalService) getBlockInfo() (*model.BlockInfo, error) {
	var query = url.Values{}
	query.Add("height", fmt.Sprintf("\"%d\"", is.currentBlock))

	res, err := is.makeTendermintRPCRequest("/block", query.Encode())
	if err != nil {
		logger.Logger.Error("getBlockTxs", zap.Error(err))
		return nil, err
	}

	var blockInfo = new(model.BlockInfo)
	err = jsoniter.Unmarshal(res, blockInfo)
	if err != nil {
		logger.Logger.Error("getBlockTxs", zap.Error(err))
		return nil, err
	}

	return blockInfo, nil
}

func (is *InternalService) getBlockTxs() ([]model.Tx, error) {
	var query = url.Values{}
	if is.config.TxType != "" {
		query.Add("query", fmt.Sprintf("\"tx.height=%d AND message.action='%s'\"", is.currentBlock, is.config.TxType))
	} else {
		query.Add("query", fmt.Sprintf("\"tx.height=%d\"", is.currentBlock))
	}

	res, err := is.makeTendermintRPCRequest("/tx_search", query.Encode())
	if err != nil {
		logger.Logger.Error("getBlockTxs", zap.Error(err))
		return nil, err
	}

	blockInfo := model.BlockTransactions{}
	err = jsoniter.Unmarshal(res, &blockInfo)
	if err != nil {
		logger.Logger.Error("getBlockTxs", zap.Error(err))
		return nil, err
	}

	return blockInfo.Txs, nil
}

func (is *InternalService) makeTendermintRPCRequest(url string, query string) ([]byte, error) {
	nodeAddress := cast.ToString(is.Context.GetConfig("node_address", ""))
	endpoint := fmt.Sprintf("%s%s?%s", nodeAddress, url, query)

	resp, err := http.Get(endpoint)
	if err != nil {
		logger.Logger.Error("MakeTendermintRPCRequest - Unable to connect to server", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Logger.Error("CosmosGateway - Handle - gRPC gateway error response", zap.Error(err))
		return nil, err
	}

	var result = new(model.RPCResponse)
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		logger.Logger.Error("[query-network-properties] Invalid response format", zap.Error(err))
		return nil, err
	}

	if resp.StatusCode >= 400 {

		errMsg := fmt.Sprintf("gRPC gateway error: url=%s status=%d, code=%d, message=%s", endpoint, resp.StatusCode, result.ID, result.Error)
		logger.Logger.Error("CosmosGateway - Handle - gRPC gateway error response",
			zap.Int("status", resp.StatusCode),
			zap.Any("code", result.ID),
			zap.Any("message", result.Error),
		)
		return nil, fmt.Errorf(errMsg)
	}

	return result.Result, nil
}

func (is *InternalService) rewriteLastHandledBlock(blockHeight int64) error {
	return os.WriteFile(filePathLatestBlock, []byte(strconv.Itoa(int(blockHeight))), os.ModePerm)
}
