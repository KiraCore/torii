package model

type ServiceConfig struct {
	NodeAddress    string
	TxType         string
	SkipFailedTxs  bool
	CollectionName string
}

type StorageConfig struct {
	Token      string
	Url        string
	Email      string
	Password   string
	Collection string
}
