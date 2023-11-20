package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func GeneratePrivateKey() secp256k1.PrivKey {
	return secp256k1.GenPrivKey()
}

func GetPubkeyString(privKey secp256k1.PrivKey) string {
	return base64.StdEncoding.EncodeToString(privKey.PubKey().Bytes())
}

func GetPubkeyFromString(pubkeyStr string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(pubkeyStr)
}

func LenSyncMap(m *sync.Map) int {
	var i int
	m.Range(func(k, v interface{}) bool {
		i++
		return true
	})
	return i
}

func LoadKeyFile() (*keygen.LocalPartySaveData, error) {
	b, err := os.ReadFile("key.json")
	if err != nil {
		return nil, fmt.Errorf("load key file : %w", err)
	}

	var key = new(keygen.LocalPartySaveData)

	if err = json.Unmarshal(b, &key); err != nil {
		return nil, fmt.Errorf("unmarshal : %w", err)
	}

	return key, nil
}

func SaveKeyFile(end *keygen.LocalPartySaveData) error {
	jsonStr, err := json.Marshal(end)
	if err != nil {
		return fmt.Errorf("marshal : %w", err)
	}
	path := "key.json"
	err = os.WriteFile(path, jsonStr, 0644)
	if err != nil {
		return fmt.Errorf("create file error : %w", err)
	}
	return nil
}
