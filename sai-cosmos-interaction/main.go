package main

import (
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/saiset-co/sai-service-crud-plus/logger"
	"github.com/saiset-co/saiCosmosInteraction/internal"
	"github.com/saiset-co/saiService"
)

var (
	AccountAddressPrefix   = "kira"
	AccountPubKeyPrefix    = "kirapub"
	ValidatorAddressPrefix = "kiravaloper"
	ValidatorPubKeyPrefix  = "kiravaloperpub"
	ConsNodeAddressPrefix  = "kiravalcons"
	ConsNodePubKeyPrefix   = "kiravalconspub"
)

func main() {
	svc := saiService.NewService("saiCosmosInteraction")
	svc.RegisterConfig("config.yml")

	logger.Logger = svc.Logger

	is := internal.InternalService{Context: svc.Context}

	svc.RegisterInitTask(is.Init)

	config := types.GetConfig()
	config.SetBech32PrefixForAccount(AccountAddressPrefix, AccountPubKeyPrefix)
	config.SetBech32PrefixForValidator(ValidatorAddressPrefix, ValidatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(ConsNodeAddressPrefix, ConsNodePubKeyPrefix)
	config.Seal()

	svc.RegisterHandlers(
		is.NewHandler(),
	)

	svc.Start()
}
