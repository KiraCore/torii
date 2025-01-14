package internal

import (
	"os"

	jsoniter "github.com/json-iterator/go"
)

func (is *InternalService) loadAddresses() error {
	fileBytes, err := os.ReadFile(filePathAddresses)
	if err != nil {
		return err
	}

	var addressArray []string
	err = jsoniter.Unmarshal(fileBytes, &addressArray)
	if err != nil {
		return err
	}

	is.mu.Lock()
	for _, address := range addressArray {
		is.addresses[address] = struct{}{}
	}
	is.mu.Unlock()

	return nil
}

func (is *InternalService) rewriteAddressesFile() error {
	addressArray := make([]string, len(is.addresses))
	is.mu.Lock()
	i := 0
	for k, _ := range is.addresses {
		addressArray[i] = k
		i++
	}
	is.mu.Unlock()

	jsonBytes, err := jsoniter.Marshal(&addressArray)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePathAddresses, jsonBytes, os.ModePerm)

	return err
}
