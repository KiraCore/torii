package eth

import (
	"github.com/onrik/ethrpc"
	"go.uber.org/zap"
)

func GetClient(address string, logger *zap.Logger) (*ethrpc.EthRPC, error) {
	ethClient := ethrpc.New(address)

	// got -32000 context cancelled error when get web3clientVersion

	// version, err := ethClient.Web3ClientVersion()
	// if err != nil {
	// 	logger.Error("get version", zap.Error(err))
	// 	return nil, err
	// }

	version, err := ethClient.NetVersion()
	if err != nil {
		logger.Error("get net version", zap.Error(err))
		return nil, err
	}

	logger.Sugar().Debugf("Connected to geth server  : %s,  client net version : %s", ethClient.URL(), version)

	return ethClient, nil
}
