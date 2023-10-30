package utils

import (
	"encoding/base64"

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
