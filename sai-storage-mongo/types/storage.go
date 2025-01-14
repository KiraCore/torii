package types

type StorageConfig struct {
	Atlas            bool    `json:"atlas" yaml:"atlas"`
	Host             string  `json:"host" yaml:"host"`
	Port             string  `json:"port" yaml:"port"`
	User             string  `json:"user" yaml:"user"`
	Pass             string  `json:"pass" yaml:"pass"`
	Database         string  `json:"database" yaml:"database"`
	ConnectionString string  `json:"connectionString" yaml:"connectionString"`
	Duplicate        bool    `json:"duplicate" yaml:"duplicate"`
	DuplicateURL     string  `json:"duplicateURL" yaml:"duplicateURL"`
	DuplicateTimeout float64 `json:"duplicateTimeout" yaml:"duplicateTimeout"`
	DuplicatePause   float64 `json:"duplicatePause" yaml:"duplicatePause"`
	DuplicateMethod  string  `json:"duplicateMethod" yaml:"duplicateMethod"`
}
