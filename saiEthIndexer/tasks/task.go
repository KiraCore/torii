package tasks

import (
	"encoding/json"
	"fmt"
	"github.com/adam-lavrik/go-imath/ix"
	"os"
	"sync"
	"time"

	"github.com/onrik/ethrpc"
	"github.com/saiset-co/saiEthIndexer/config"
	"github.com/saiset-co/saiEthIndexer/pkg/eth"
	"github.com/saiset-co/saiEthIndexer/utils"
	"go.uber.org/zap"
)

const (
	configPath    = "./config.json"
	contractsPath = "./contracts.json"
)

type TaskManager struct {
	Config       *config.Configuration
	EthClient    *ethrpc.EthRPC
	Logger       *zap.Logger
	BlockManager *BlockManager
	resultChan   chan error
}

var StopLoop bool

func NewManager(config *config.Configuration, logger *zap.Logger) (*TaskManager, error) {
	ethClient, err := eth.GetClient(config.Specific.GethServer, logger)
	if err != nil {
		return nil, err
	}

	blockManager := NewBlockManager(*config, logger)

	return &TaskManager{
		Config:       config,
		EthClient:    ethClient,
		Logger:       logger,
		BlockManager: blockManager,
		resultChan:   make(chan error),
	}, nil
}

// Process blocks, which got from geth-server
func (t *TaskManager) ProcessBlocks() {

	for {
		StopLoop = false
		blockID, err := t.EthClient.EthBlockNumber()
		if err != nil {
			t.Logger.Error("tasks - ProcessBlocks - get block number from eth client", zap.Error(err))
			continue
		}

		t.Logger.Sugar().Debugf("get most recent block from geth-server : %d", blockID)

		blk, err := t.BlockManager.GetLastBlock(blockID)
		if err != nil {
			continue
		}

		t.Logger.Sugar().Debugf("get most recent block from storage : %d", blockID)

		for i := blk.ID; i <= blockID; i++ {
			blkInfo, err := t.EthClient.EthGetBlockByNumber(i, true)
			if err != nil || blkInfo == nil {
				blk.ID = i
				t.BlockManager.SetLastBlock(blk)
				t.Logger.Error("tasks - ProcessBlocks - get block by number from server", zap.Error(err))
				continue
			}

			if len(blkInfo.Transactions) == 0 {
				blk.ID = i
				t.BlockManager.SetLastBlock(blk)
				t.Logger.Info("tasks - ProcessBlocks - get block by number from server - transactions - no transactions found", zap.Int("current block id in for cycle", i), zap.Int("current block id from eth server", blockID))
				continue
			}

			t.Logger.Sugar().Debugf("block %d from %d analyzed, %d total transactions", i, blockID, len(blkInfo.Transactions))
			receipts := map[string]*ethrpc.TransactionReceipt{}

			for _, tr := range blkInfo.Transactions {
				receipt, err := t.EthClient.EthGetTransactionReceipt(tr.Hash)

				if err != nil {
					t.Logger.Error("tasks - ProcessBlocks - get transaction receipt from server", zap.Error(err))
					continue
				}

				receipts[tr.Hash] = receipt
			}

			t.BlockManager.HandleTransactions(blkInfo.Transactions, receipts)

			if StopLoop {
				t.Logger.Sugar().Debug("Got new contract, will continue with the new lowest block.")
				break
			} else {
				blk.ID = i
				t.BlockManager.SetLastBlock(blk)
			}
		}

		time.Sleep(time.Duration(t.Config.Specific.Sleep) * time.Second)
	}
}

func (t *TaskManager) AddContract(contracts []config.Contract) error {
	t.Config.EthContracts.Mutex.Lock()
	defer t.Config.EthContracts.Mutex.Unlock()
	t.Config.EthContracts.Contracts = append(t.Config.EthContracts.Contracts, contracts...)

	var blocks []int
	for i := 0; i < len(contracts); i++ {
		blocks = append(blocks, contracts[i].StartBlock)
	}

	lblk, glbErr := t.BlockManager.GetLastBlock(0)
	if glbErr != nil {
		t.Logger.Error("task - addContract - get last block", zap.Error(glbErr))
		return glbErr
	}

	blocks = append(blocks, lblk.ID)

	blk := &Block{ID: ix.MinSlice(blocks)}
	t.BlockManager.SetLastBlock(blk)
	StopLoop = true

	go t.RewriteContractsConfig(contractsPath)

	err := <-t.resultChan
	if err != nil {
		return err
	}
	t.Logger.Sugar().Debugf("contracts")

	return nil
}

func (t *TaskManager) RewriteContractsConfig(contractsConfigPath string) {
	data, err := json.MarshalIndent(&t.Config.EthContracts, "", "	")
	if err != nil {
		t.Logger.Error("task - addContract - rewrite contracts config - marshal contracts", zap.Error(err))
		t.resultChan <- err
		return
	}

	f, err := os.OpenFile(contractsPath, os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		t.Logger.Error("task - addContract - rewrite contracts config - open contract config file", zap.Error(err))
		t.resultChan <- err
		return
	}

	err = f.Truncate(0)
	if err != nil {
		t.Logger.Error("task - addContract - rewrite contracts config - truncate", zap.Error(err))
		t.resultChan <- err
		return
	}

	_, err = f.Seek(0, 0)
	if err != nil {
		t.Logger.Error("task - addContract - rewrite contracts config - seek", zap.Error(err))
		t.resultChan <- err
		return
	}

	_, err = f.Write(data)
	if err != nil {
		t.Logger.Error("task - addContract - rewrite contracts config - write", zap.Error(err))
		t.resultChan <- err
		return
	}

	err = t.ReloadContracts()
	if err != nil {
		t.Logger.Error("task - addContract - rewrite contracts config - reload contract", zap.Error(err))
		t.resultChan <- err
		return
	}

	t.resultChan <- nil

	t.Logger.Sugar().Infof("active config : %+v\n", t.Config)
	return
}

// reload config after contracts was added via http add_contracts endpoint
func (t *TaskManager) ReloadContracts() error {
	contracts := config.EthContracts{}
	b, err := os.ReadFile(contractsPath)
	if err != nil {
		return fmt.Errorf("contracts json read error: %w", err)
	}
	err = json.Unmarshal(b, &contracts)
	if err != nil {
		return fmt.Errorf("contracts json unmarshal error: %w", err)
	}

	t.Config.EthContracts = contracts
	t.Config.EthContracts.Mutex = new(sync.RWMutex)
	t.BlockManager.config.EthContracts = contracts
	t.BlockManager.config.EthContracts.Mutex = new(sync.RWMutex)
	return nil
}

func (t *TaskManager) DeleteContract(contracts []string) error {
	t.Config.EthContracts.Mutex.Lock()
	defer t.Config.EthContracts.Mutex.Unlock()
	var notFoundAddresses []string

LOOP:
	for _, incomingAddress := range contracts {
		for i, existingContract := range t.Config.EthContracts.Contracts {
			t.Logger.Sugar().Infof("address : %s, index : %d", existingContract.Address, i)
			if incomingAddress == existingContract.Address {
				t.Config.EthContracts.Contracts = utils.RemoveContract(t.Config.EthContracts.Contracts, i)
				continue LOOP
			}
		}
		notFoundAddresses = append(notFoundAddresses, incomingAddress)
	}

	if len(notFoundAddresses) != 0 {
		t.Logger.Info("Not found addresses to delete", zap.Strings("addresses", notFoundAddresses), zap.Int("len", len(notFoundAddresses)))
	}

	go t.RewriteContractsConfig(contractsPath)

	err := <-t.resultChan
	if err != nil {
		return err
	}

	return nil
}
