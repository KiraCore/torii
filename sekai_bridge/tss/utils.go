package tss

import (
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/binance-chain/tss-lib/tss"
	"github.com/btcsuite/btcd/btcec"
)

// convert simple pubkey (id) string to tss.PartyID
func PubkeyToPartyID(pubkey string) *tss.PartyID {
	mon := fmt.Sprintf("moniker_%s", pubkey)
	key, _ := new(big.Int).SetString(pubkey, 10)
	partyID := tss.NewPartyID(pubkey, mon, key)
	return partyID
}

func (t *TssServer) GetParties(localPartyKey string) ([]*tss.PartyID, *tss.PartyID, error) {
	var localPartyID *tss.PartyID
	var unSortedPartiesID []*tss.PartyID

	keys := make([]string, 0)
	t.RWMutex.RLock()
	for pubkey := range t.ConnectionStorage {
		keys = append(keys, pubkey)
	}
	t.RWMutex.RUnlock()

	keys = append(keys, t.Pubkey)
	for _, item := range keys {
		// simple ids used
		mon := fmt.Sprintf("moniker_%s", item)
		key, _ := new(big.Int).SetString(item, 10)
		partyID := tss.NewPartyID(item, mon, key)

		if item == localPartyKey {
			localPartyID = partyID
		}

		unSortedPartiesID = append(unSortedPartiesID, partyID)
	}

	if localPartyID == nil {
		return nil, nil, errors.New("local party is not in the list")
	}

	partiesID := tss.SortPartyIDs(unSortedPartiesID)

	//	t.Logger.Info("sorted order", zap.Any("keys", partiesID.Keys()))

	for _, partyID := range partiesID {
		t.PartiesMap[*partyID] = true
	}
	return partiesID, localPartyID, nil
}

// get party id address
func GetPeersAddresses(connStorage map[string]string) []string {
	mu := new(sync.RWMutex)

	addrs := make([]string, 0)
	mu.RLock()
	for _, addr := range connStorage {
		addrs = append(addrs, addr)
	}
	mu.RUnlock()
	// for partyID := range t.PartiesMap {
	// 	id := partyID.Id
	// 	addr, ok := t.ConnectionStorage[id]
	// 	if !ok {
	// 		t.Logger.Error("tss -> utils -> id from PartyID map was not found", zap.String("id", id))
	// 		continue
	// 	}
	// 	addrs = append(addrs, addr)
	return addrs
}

func MsgToHashInt(msg []byte) (*big.Int, error) {
	return hashToInt(msg, btcec.S256()), nil
}

func MsgToHashString(msg []byte) (string, error) {
	if len(msg) == 0 {
		return "", errors.New("empty message")
	}
	h := sha256.New()
	_, err := h.Write(msg)
	if err != nil {
		return "", fmt.Errorf("fail to caculate sha256 hash: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func hashToInt(hash []byte, c elliptic.Curve) *big.Int {
	orderBits := c.Params().N.BitLen()
	orderBytes := (orderBits + 7) / 8
	if len(hash) > orderBytes {
		hash = hash[:orderBytes]
	}

	ret := new(big.Int).SetBytes(hash)
	excess := len(hash)*8 - orderBits
	if excess > 0 {
		ret.Rsh(ret, uint(excess))
	}
	return ret
}

func StringToBigInt(s string) *big.Int {
	i := new(big.Int)
	i.SetString(s, 16)
	return i
}
