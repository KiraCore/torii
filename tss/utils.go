package tss

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/binance-chain/tss-lib/tss"
	"go.uber.org/zap"
)

// convert simple pubkey (id) string to tss.PartyID
func PubkeyToPartyID(pubkey string) *tss.PartyID {
	mon := fmt.Sprintf("moniker_%s", pubkey)
	key, _ := new(big.Int).SetString(pubkey, 10)
	partyID := tss.NewPartyID(pubkey, mon, key)
	return partyID
}

func (t *TssKeyGen) GetParties(keys []string, localPartyKey string) ([]*tss.PartyID, *tss.PartyID, error) {
	var localPartyID *tss.PartyID
	var unSortedPartiesID []*tss.PartyID
	for _, item := range keys {
		// simple ids used
		mon := fmt.Sprintf("moniker_%s", item)
		key, _ := new(big.Int).SetString(item, 10)
		partyID := tss.NewPartyID(item, mon, key)

		if item == localPartyKey {
			localPartyID = partyID
		}

		unSortedPartiesID = append(unSortedPartiesID, partyID)

		// pk, err := sdk.UnmarshalPubKey(sdk.AccPK, item)
		// if err != nil {
		// 	return nil, nil, fmt.Errorf("fail to get account pub key address(%s): %w", item, err)
		// }
		// key := new(big.Int).SetBytes(pk.Bytes())
		// // Set up the parameters
		// // Note: The `id` and `moniker` fields are for convenience to allow you to easily track participants.
		// // The `id` should be a unique string representing this party in the network and `moniker` can be anything (even left blank).
		// // The `uniqueKey` is a unique identifying key for this peer (such as its p2p public key) as a big.Int.
		// partyID := tss.NewPartyID(strconv.Itoa(idx), "", key)
		// if item == localPartyKey {
		// 	localPartyID = partyID
		// }
		// unSortedPartiesID = append(unSortedPartiesID, partyID)
	}

	if localPartyID == nil {
		return nil, nil, errors.New("local party is not in the list")
	}

	partiesID := tss.SortPartyIDs(unSortedPartiesID)

	for _, partyID := range partiesID {
		t.PartiesMap[*partyID] = true
	}
	return partiesID, localPartyID, nil
}

// get party id address
func (t *TssKeyGen) GetPartyIDpeerAddr() []string {
	addrs := make([]string, 0)

	for partyID := range t.PartiesMap {
		id := partyID.Id
		addr, ok := t.ConnectionStorage[id]
		if !ok {
			t.Logger.Error("tss -> utils -> id from PartyID map was not found", zap.String("id", id))
			continue
		}
		addrs = append(addrs, addr)
	}
	return addrs
}
