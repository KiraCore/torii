package internal

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/saiset-co/sai-eth-interaction/models"
	"github.com/saiset-co/sai-eth-interaction/utils"
	"go.uber.org/zap"
)

// get contracts from contracts.json file when start app
func (s *InternalService) getInitialContracts(path string) error {
	f, err := os.OpenFile(contractsPath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		s.Logger.Error("main - getInitialContract - open file", zap.Error(err))
		return err
	}

	defer func() {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		s.Logger.Error("main - getInitialContract - read from file", zap.Error(err))
		return err
	}

	// file was created
	if len(b) == 0 {
		s.Logger.Debug("Contracts file was not found, created a new one")
		return nil
	}

	contracts := models.Contracts{}
	err = json.Unmarshal(b, &contracts)
	if err != nil {
		s.Logger.Error("main - init - getInitialContracts - unmarshal data from contracts file", zap.Error(err))
		return fmt.Errorf("contracts json unmarshal error: %w", err)
	}

	s.Contracts = contracts.Contracts
	s.Logger.Sugar().Debugf("FOUND CONTRACTS :%+v\n", s.Contracts) // DEBUG
	return nil
}

// check if some of incomingContracts already exists in contracts file
func (s *InternalService) filterUniqueContracts(incomingContracts []models.Contract) []models.Contract {
	checkedContracts := make([]models.Contract, 0)
LOOP:
	for _, incomingContract := range incomingContracts {
		b, err := json.Marshal(incomingContract)
		if err != nil {
			Service.Logger.Error("handlers - add - filterUniqueContracts - marshal incoming contract", zap.String("contract name", incomingContract.Name), zap.Error(err))
			continue
		}
		incomingHash := sha256.Sum256(b)
		for _, contract := range Service.Contracts {
			b1, err := json.Marshal(contract)
			if err != nil {
				Service.Logger.Error("handlers - add - filterUniqueContracts - marshal existing contract", zap.String("contract name", incomingContract.Name), zap.Error(err))
				continue
			}
			contractHash := sha256.Sum256(b1)
			if incomingHash == contractHash {
				Service.Logger.Debug("handlers - add - contract already exists", zap.Any("contract", contract))
				continue LOOP
			}
		}
		checkedContracts = append(checkedContracts, incomingContract)
	}
	return checkedContracts
}

// rewrite contracts file
func (s *InternalService) RewriteContractsConfig(contractsConfigPath string) error {
	data, err := json.MarshalIndent(models.Contracts{
		Contracts: s.Contracts,
	}, "", "	")
	if err != nil {
		s.Logger.Error("handlers - add - rewrite contracts config - marshal contracts", zap.Error(err))
		return err
	}

	f, err := os.OpenFile(contractsPath, os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		s.Logger.Error("handlers - add - rewrite contracts config - open contract file", zap.Error(err))
		return err
	}

	err = f.Truncate(0)
	if err != nil {
		s.Logger.Error("handlers - add - rewrite contracts config - truncate", zap.Error(err))
		return err
	}

	_, err = f.Seek(0, 0)
	if err != nil {
		s.Logger.Error("handlers - add - rewrite contracts config - seek", zap.Error(err))
		return err
	}

	_, err = f.Write(data)
	if err != nil {
		s.Logger.Error("handlers - add - rewrite contracts config - write", zap.Error(err))
		return err
	}

	return nil
}

func (s *InternalService) GetContractByName(name string) (contract *models.Contract, err error) {
	for _, c := range s.Contracts {
		if c.Name == name {
			contract = &c
			return contract, nil
		}
	}
	return nil, fmt.Errorf("Contract name = [%s] was not found in contract file", name)

}

func (s *InternalService) DeleteContracts(deleteContractName *models.DeleteData) {
	var notFoundNames []string

LOOP:
	for _, name := range deleteContractName.Names {
		for i, existingContract := range Service.Contracts {
			if name == existingContract.Name {
				Service.Contracts = utils.RemoveContract(Service.Contracts, i)
				continue LOOP
			}
		}
		notFoundNames = append(notFoundNames, name)
	}

	if len(notFoundNames) != 0 {
		Service.Logger.Info("not found names to delete", zap.Strings("names", notFoundNames), zap.Int("len", len(notFoundNames)))
	}
}
