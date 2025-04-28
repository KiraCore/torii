package internal

import (
	"net/http"

	saiService "github.com/saiset-co/sai-service/service"
)

func (is *InternalService) NewHandler() saiService.Handler {
	return saiService.Handler{
		"add_address": saiService.HandlerElement{
			Name:        "add_address",
			Description: "Add new address for scan transactions",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				return is.addAddress(data)
			},
		},
		"delete_address": saiService.HandlerElement{
			Name:        "delete_address",
			Description: "Delete address from addresses list",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				return is.deleteAddress(data)
			},
		},
	}
}

func (is *InternalService) addAddress(data interface{}) (string, int, error) {
	address, ok := data.(string)
	if !ok {
		return "address should be a string", http.StatusBadRequest, nil
	}
	_, exist := is.addresses[address]
	if exist {
		return "addAddress", http.StatusOK, nil
	}

	is.mu.Lock()
	is.addresses[address] = struct{}{}
	is.mu.Unlock()

	err := is.rewriteAddressesFile()
	if err != nil {
		is.mu.Lock()
		delete(is.addresses, address)
		is.mu.Unlock()
		return "addAddress", http.StatusInternalServerError, err
	}

	return "addAddress", http.StatusOK, nil
}

func (is *InternalService) deleteAddress(data interface{}) (string, int, error) {
	address, ok := data.(string)
	if !ok {
		return "address should be a string", http.StatusBadRequest, nil
	}

	_, exist := is.addresses[address]
	if !exist {
		return "deleteAddress", http.StatusOK, nil
	}

	is.mu.Lock()
	delete(is.addresses, address)
	is.mu.Unlock()

	err := is.rewriteAddressesFile()
	if err != nil {
		return "deleteAddress", http.StatusInternalServerError, err
	}

	return "deleteAddress", http.StatusOK, nil
}
