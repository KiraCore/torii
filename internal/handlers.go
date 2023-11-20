package internal

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/KiraCore/sekai-bridge/tss"
	"github.com/KiraCore/sekai-bridge/types"
	"github.com/binance-chain/tss-lib/common"
	"github.com/saiset-co/saiService"
	"go.uber.org/zap"
)

func (is *InternalService) NewHandler() saiService.Handler {
	return saiService.Handler{
		"keygen": saiService.HandlerElement{
			Name:        "keygen",
			Description: "Start threshold keys generation",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				return is.keygen()
			},
		},
		"sign": saiService.HandlerElement{
			Name:        "sign",
			Description: "Sign the data",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				return is.sign(data)
			},
		},
		"verify": saiService.HandlerElement{
			Name:        "verify",
			Description: "verify signature",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				return is.verify(data)
			},
		},
		"stats": saiService.HandlerElement{
			Name:        "stats",
			Description: "p2p stats",
			Function: func(data, meta interface{}) (interface{}, int, error) {
				return is.stats()
			},
		},
	}
}

// Keygen handler
func (is *InternalService) keygen() (interface{}, int, error) {
	response, err := is.Tss.Keygen(is.Tss.Parties, is.Tss.Threshold)
	if err != nil {
		return "keygen error", 500, err
	}

	return response.Key.ECDSAPub.Y(), 200, nil
}

// Keysign handler
func (is *InternalService) sign(data interface{}) (interface{}, int, error) {
	var request tss.SignMessageRequest

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return "marshaling error", 500, err
	}

	err = json.Unmarshal(dataJSON, &request)
	if err != nil {
		return "unmarshaling error", 500, err
	}

	response, err := is.Tss.Sign(&request)
	if err != nil {
		return "sign error", 500, err
	}

	return response, 200, nil
}

// stats handler (for debugging)
func (is *InternalService) stats() (interface{}, int, error) {
	stats := &types.SekaiBridgeStats{
		HTTPPort:             is.P2P.Config.Http.Port,
		P2PPort:              is.P2P.Config.P2P.Port,
		P2PSlot:              is.P2P.Config.P2P.Slot,
		P2pPeers:             is.P2P.Config.Peers,
		P2PConnectionStorage: is.P2P.ConnectionStorage,
		P2PSavedMessages:     make(map[string]bool),
		TssPartyID:           *is.Tss.LocalPartyID,
		TssConnectionStorage: is.Tss.ConnectionStorage,
		TssPartiesMap:        make(map[string]bool),
		TssKeygenMsgStorage:  make(map[string]bool),
	}
	if is.Tss.PartiesMap != nil {
		for id := range is.Tss.PartiesMap {
			stats.TssPartiesMap[id.Id+"|"+id.Moniker+"|"+fmt.Sprintf("index = %d", id.Index)] = true
		}
	}
	if is.Tss.KeygenInstance != nil {
		for _, msg := range is.Tss.KeygenInstance.KeygenMsgsStorage.M {
			stats.TssKeygenMsgStorage[msg.From.GetId()+"|"+msg.Type] = false
		}
	}

	it := is.P2P.Cache.Iterator()
	for it.SetNext() {
		entryInfo, err := it.Value()
		if err != nil {
			is.P2P.Logger.Error("p2p -> server -> GetStats -> it.Value", zap.Error(err))
			continue
		}
		stats.P2PSavedMessages[entryInfo.Key()] = true
	}

	return stats, 200, nil
}

// Keysign handler
func (is *InternalService) verify(data interface{}) (interface{}, int, error) {
	var request verifyRequest

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return "marshaling error", 500, err
	}

	err = json.Unmarshal(dataJSON, &request)
	if err != nil {
		return "unmarshaling error", 500, err
	}

	signatureData, err := base64.StdEncoding.DecodeString(request.Signature)
	if err != nil {
		return "DecodeString error", 500, err
	}

	signature := common.ECSignature{}

	err = json.Unmarshal(signatureData, &signature)
	if err != nil {
		return "Unmarshal error", 500, err
	}

	isValid := is.Tss.VerifySignature(&signature, request.Msg)

	return verifyResponse{Valid: isValid}, 200, nil
}

type verifyRequest struct {
	Msg       string `json:"msg"`
	Signature string `json:"signature"`
}

type verifyResponse struct {
	Valid bool `json:"is_valid"`
}
