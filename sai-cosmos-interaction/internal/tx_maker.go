package internal

import (
	"bytes"
	"encoding/base64"
	"fmt"
	types2 "github.com/KiraCore/sekai/x/bridge/types"
	"io"
	"net/http"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/saiset-co/saiCosmosInteraction/internal/model"
	"github.com/spf13/cast"
)

type TransactionMaker struct {
	nodeAddress   string
	cli           http.Client
	codec         *codec.ProtoCodec
	txConfig      client.TxConfig
	txBuilder     client.TxBuilder
	senderAcc     types.AccAddress
	senderAccInfo model.AccountInfo
	fromAddr      string
	receiverAddr  types.AccAddress
	chainID       string
	kRing         keyring.Keyring
	kRingUUID     string
}

func NewTransactionMaker(senderAddress, nodeAddress, chainID, fromAddr, toAddr, passphrase string, privateKey []byte) (*TransactionMaker, error) {
	tm := new(TransactionMaker)
	tm.nodeAddress = nodeAddress
	tm.cli = http.Client{Timeout: time.Second * 5}

	var err error
	tm.senderAcc, err = types.AccAddressFromBech32(senderAddress)
	if err != nil {
		return nil, err
	}

	tm.senderAccInfo, err = tm.GetAccountInfo(tm.senderAcc.String())
	if err != nil {
		return nil, err
	}

	tm.receiverAddr, err = types.AccAddressFromBech32(toAddr)
	if err != nil {
		return nil, err
	}

	tm.fromAddr = fromAddr

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	interfaceRegistry.RegisterInterface("types.PubKey", (*cryptotypes.PubKey)(nil), &secp256k1.PubKey{})
	interfaceRegistry.RegisterInterface("types.PrivKey", (*cryptotypes.PrivKey)(nil), &secp256k1.PrivKey{})
	interfaceRegistry.RegisterInterface("types.Msg", (*types.Msg)(nil), &banktypes.MsgSend{})
	tm.codec = codec.NewProtoCodec(interfaceRegistry)
	tm.txConfig = authtx.NewTxConfig(tm.codec, authtx.DefaultSignModes)
	tm.txBuilder = tm.txConfig.NewTxBuilder()
	tm.chainID = chainID

	tm.kRing = keyring.NewInMemory(tm.codec)
	tm.kRingUUID = uuid.New().String()
	err = tm.kRing.ImportPrivKey(tm.kRingUUID, string(privateKey), "12345678")
	if err != nil {
		return nil, err
	}

	return tm, nil
}

func (tm *TransactionMaker) BuildTx(gasLimit uint64, amount, feeAmount int64, memo string) error {
	message := types2.NewMsgChangeEthereumCosmos(
		tm.senderAcc,
		tm.fromAddr,
		tm.receiverAddr,
		types.NewCoins(types.NewInt64Coin("ukex", amount)),
	)
	err := tm.txBuilder.SetMsgs(message)
	if err != nil {
		return err
	}

	tm.txBuilder.SetGasLimit(gasLimit)
	tm.txBuilder.SetMemo(memo)
	tm.txBuilder.SetFeeAmount(types.NewCoins(types.NewInt64Coin("ukex", feeAmount)))

	return nil
}

func (tm *TransactionMaker) SignTx() error {
	rec, err := tm.kRing.Key(tm.kRingUUID)
	if err != nil {
		return err
	}

	pubKey, err := rec.GetPubKey()
	if err != nil {
		return err
	}

	seq := cast.ToUint64(tm.senderAccInfo.Account.Sequence)
	num := cast.ToUint64(tm.senderAccInfo.Account.AccountNumber)
	err = tm.txBuilder.SetSignatures(signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  tm.txConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: seq,
	})

	if err != nil {
		return err
	}

	signerData := xauthsigning.SignerData{
		Address:       tm.senderAcc.String(),
		ChainID:       tm.chainID,
		AccountNumber: num,
		Sequence:      seq,
		PubKey:        pubKey,
	}

	var sigV2 signing.SignatureV2

	signMode := tm.txConfig.SignModeHandler().DefaultMode()
	signBytes, err := tm.txConfig.SignModeHandler().GetSignBytes(signMode, signerData, tm.txBuilder.GetTx())
	if err != nil {
		return err
	}

	sig, _, err := tm.kRing.SignByAddress(tm.senderAcc, signBytes)
	if err != nil {
		return err
	}

	sigData := signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: sig,
	}

	sigV2 = signing.SignatureV2{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: signerData.Sequence,
	}

	err = tm.txBuilder.SetSignatures(sigV2)

	return err
}

func (tm *TransactionMaker) BroadcastTx() (string, error) {
	const urlTemplate = "%s/cosmos/tx/v1beta1/txs"

	txBytes, err := tm.txConfig.TxEncoder()(tm.txBuilder.GetTx())
	if err != nil {
		return "", err
	}

	broadcastReq := model.TxBroadcastReq{
		TxBytes: base64.StdEncoding.EncodeToString(txBytes),
		Mode:    tx.BroadcastMode_BROADCAST_MODE_SYNC.String(),
	}

	txReqBytes, err := jsoniter.Marshal(broadcastReq)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost,
		fmt.Sprintf(urlTemplate, tm.nodeAddress),
		bytes.NewReader(txReqBytes),
	)
	if err != nil {
		return "", err
	}

	res, err := tm.cli.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s", bodyBytes)
	}

	broadcastRes := model.TxBroadcastRes{}
	err = jsoniter.Unmarshal(bodyBytes, &broadcastRes)
	if err != nil {
		return "", err
	}

	if broadcastRes.TxResponse.Code != 0 {
		return "", fmt.Errorf("%s", bodyBytes)
	}

	return broadcastRes.TxResponse.Txhash, nil
}

func (tm *TransactionMaker) GetAccountInfo(address string) (model.AccountInfo, error) {
	const urlTemplate = "%s/cosmos/auth/v1beta1/accounts/%s"

	res, err := tm.cli.Get(fmt.Sprintf(urlTemplate, tm.nodeAddress, address))
	if err != nil {
		return model.AccountInfo{}, err
	}

	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return model.AccountInfo{}, err
	}

	if res.StatusCode != http.StatusOK {
		return model.AccountInfo{}, fmt.Errorf("%s", bodyBytes)
	}

	ai := model.AccountInfo{}
	err = jsoniter.Unmarshal(bodyBytes, &ai)

	return ai, err
}
