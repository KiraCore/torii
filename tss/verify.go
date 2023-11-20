package tss

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/binance-chain/tss-lib/common"
)

func (t *TssServer) VerifySignature(signature *common.ECSignature, msg string) bool {
	pkX, pkY := t.Key.ECDSAPub.X(), t.Key.ECDSAPub.Y()
	pk := ecdsa.PublicKey{
		Curve: t.Key.ECDSAPub.ToECDSAPubKey().Curve,
		X:     pkX,
		Y:     pkY,
	}
	r := new(big.Int).SetBytes(signature.GetR())
	s := new(big.Int).SetBytes(signature.GetS())

	return ecdsa.Verify(&pk, []byte(msg), r, s)
}
