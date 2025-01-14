package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	jsoniter "github.com/json-iterator/go"
	"github.com/saiset-co/saiCosmosIndexer/internal/model"
)

func (is *InternalService) getLatestBlock() (int64, error) {
	res, err := is.client.Get(is.config.NodeAddress + "/cosmos/base/tendermint/v1beta1/blocks/latest")
	if err != nil {
		return 0, err
	}

	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("getLatestBlock res status: %v, %s", res.StatusCode, bodyBytes)
	}

	lb := model.LatestBlock{}
	err = jsoniter.Unmarshal(bodyBytes, &lb)
	if err != nil {
		return 0, err
	}

	blockHeight, err := strconv.Atoi(lb.Block.Header.Height)

	return int64(blockHeight), err
}

func (is *InternalService) getBlockTxs() ([]model.TxResponse, error) {
	const (
		urlTemplateGetTxs = "%s/cosmos/tx/v1beta1/txs?pagination.limit=100&pagination.offset=%v&events=tx.height=%v&events=message.action='%s'"
		paginationLimit   = 100
	)

	offset := 0
	total := 1
	txs := make([]model.TxResponse, 0, paginationLimit)
	for {
		url := fmt.Sprintf(urlTemplateGetTxs, is.config.NodeAddress, offset, is.currentBlock, is.config.TxType)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		res, err := is.client.Do(req)
		if err != nil {
			return nil, err
		}

		defer res.Body.Close()

		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("getBlockTxs res status: %v, %s", res.StatusCode, bodyBytes)
		}

		blockInfo := model.BlockTransactions{}
		err = jsoniter.Unmarshal(bodyBytes, &blockInfo)
		if err != nil {
			return nil, err
		}

		if total == 0 && blockInfo.Pagination.Total != "" {
			total, err = strconv.Atoi(blockInfo.Pagination.Total)
			if err != nil {
				return nil, err
			}
		}

		txs = append(txs, blockInfo.TxResponses...)

		if len(txs) == total || len(blockInfo.TxResponses) == 0 {
			break
		}

		offset += paginationLimit
	}

	return txs, nil
}

func (is *InternalService) rewriteLastHandledBlock(blockHeight int64) error {
	return os.WriteFile(filePathLatestBlock, []byte(strconv.Itoa(int(blockHeight))), os.ModePerm)
}
