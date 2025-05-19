package model

type ServiceConfig struct {
	NodeAddress    string
	TxType         string
	SkipFailedTxs  bool
	HandleBlocks   bool
	CollectionName string
}

type StorageConfig struct {
	Token      string
	Url        string
	Email      string
	Password   string
	Collection string
}
