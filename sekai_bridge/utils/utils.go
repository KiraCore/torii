package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/KiraCore/sekai-bridge/types"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
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

func SendHttp(url string, data []byte) error {
	r := bytes.NewReader(data)
	_, err := http.Post(url, "application/json", r)
	if err != nil {
		return err
	}

	return nil
}

func SaiQuerySender(body io.Reader, address, token string) ([]byte, error) {
	const failedResponseStatus = "NOK"

	type responseWrapper struct {
		Status string              `json:"Status"`
		Error  string              `json:"Error"`
		Result jsoniter.RawMessage `json:"result"`
		Count  int                 `json:"count"`
	}

	req, err := http.NewRequest(http.MethodPost, address, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Token", token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s", resBytes)
	}

	result := responseWrapper{}
	err = jsoniter.Unmarshal(resBytes, &result)
	if err != nil {
		return nil, err
	}

	if result.Status == failedResponseStatus {
		return nil, fmt.Errorf(result.Error)
	}

	return resBytes, nil
}

func GetConfig() (types.Config, error) {
	_config := types.Config{}
	yamlData, err := os.ReadFile("config.yml")

	if err != nil {
		return _config, fmt.Errorf("readfile : %w", err)
	}

	err = yaml.Unmarshal(yamlData, &_config)

	if err != nil {
		return _config, fmt.Errorf("unmarshal : %w", err)
	}
	return _config, nil
}
