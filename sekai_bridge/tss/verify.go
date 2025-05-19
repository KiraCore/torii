package tss

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"
	"github.com/binance-chain/tss-lib/common"
	"go.uber.org/zap"
	"math/big"
)

func (t *TssServer) VerifySignature(signature *common.ECSignature, msg string) bool {
	if signature == nil {
		t.Logger.Error("tss -> VerifySignature -> signature is nil")
		return false
	}

	if t.Key == nil || t.Key.ECDSAPub == nil {
		t.Logger.Error("tss -> VerifySignature -> key or ECDSAPub is nil")
		return false
	}

	// Создаем публичный ключ ECDSA
	pkX, pkY := t.Key.ECDSAPub.X(), t.Key.ECDSAPub.Y()
	pk := ecdsa.PublicKey{
		Curve: t.Key.ECDSAPub.ToECDSAPubKey().Curve,
		X:     pkX,
		Y:     pkY,
	}

	// Хэшируем сообщение так же, как при создании подписи
	msgHash := sha256.Sum256([]byte(msg))

	// Используем непосредственно R и S из signature, без SetBytes
	r := new(big.Int).SetBytes(signature.R)
	s := new(big.Int).SetBytes(signature.S)

	t.Logger.Info("tss -> VerifySignature -> values",
		zap.String("message", msg),
		zap.String("hash_hex", fmt.Sprintf("%x", msgHash)),
		zap.String("r", r.String()),
		zap.String("s", s.String()),
		zap.String("pubkey_x", pkX.String()),
		zap.String("pubkey_y", pkY.String()))

	// Проверяем подпись с использованием хэша сообщения
	result := ecdsa.Verify(&pk, msgHash[:], r, s)

	t.Logger.Info("tss -> VerifySignature -> result", zap.Bool("verified", result))

	return result
}
