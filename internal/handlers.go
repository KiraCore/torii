package internal

import (
	"encoding/json"

	"github.com/KiraCore/sekai-bridge/types"
	"github.com/saiset-co/saiService"
)

func (is *InternalService) NewHandler() saiService.Handler {
	return saiService.Handler{
		"keygen": saiService.HandlerElement{
			Name:        "keygen",
			Description: "Start threshold keys generation",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				return is.keygen(data)
			},
		},
		"sign": saiService.HandlerElement{
			Name:        "sign",
			Description: "Sign the data",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				return is.sign(data)
			},
		},
		"stats": saiService.HandlerElement{
			Name:        "stats",
			Description: "p2p stats",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				return is.stats(data)
			},
		},
	}
}

func (is *InternalService) keygen(data interface{}) (interface{}, int, error) {
	var request types.GenerateKeysRequest

	dataJson, err := json.Marshal(data)
	if err != nil {
		return "marshaling error", 500, err
	}

	err = json.Unmarshal(dataJson, &request)
	if err != nil {
		return "unmarshaling error", 500, err
	}

	err = request.Validate()
	if err != nil {
		return "validate error", 500, err
	}

	response, err := is.Tss.Keygen(&request)
	if err != nil {
		return "keygen error", 500, err
	}

	return response.Key, 200, nil
}

func (is *InternalService) sign(data interface{}) (string, int, error) {
	var request types.SignMessageRequest

	dataJson, err := json.Marshal(data)
	if err != nil {
		return "marshaling error", 500, err
	}

	err = json.Unmarshal(dataJson, &request)
	if err != nil {
		return "unmarshaling error", 500, err
	}

	response := is.Tss.Sign(request)

	return response.Signature, 200, nil
}

func (is *InternalService) stats(data interface{}) (interface{}, int, error) {
	stats := &types.SekaiBridgeStats{
		HttpPort:             is.P2P.Config.Http.Port,
		P2PPort:              is.P2P.Config.P2P.Port,
		P2PSlot:              is.P2P.Config.P2P.Slot,
		P2pPeers:             is.P2P.Config.Peers,
		P2PConnectionStorage: is.P2P.ConnectionStorage,
		P2PSavedMessages:     is.P2P.SavedMessages,
		TssPartyID:           *is.Tss.LocalPartyID,
		TssConnectionStorage: is.Tss.ConnectionStorage,
		TssKeygenStarted:     is.Tss.KeygenStarted,
		TssPartiesMap:        make(map[string]bool),
	}
	if is.Tss.TssKeygen.PartiesMap != nil {
		for id := range is.Tss.TssKeygen.PartiesMap {
			stats.TssPartiesMap[id.Id+"|"+id.Moniker] = true
		}
	}

	return stats, 200, nil
}
