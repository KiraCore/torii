package types

type FindResult struct {
	Count  int64         `json:"count,omitempty"`
	Result []interface{} `json:"result,omitempty"`
}

type Options struct {
	Limit int64       `json:"limit"`
	Skip  int64       `json:"skip"`
	Sort  interface{} `json:"sort"`
	Count int64       `json:"count"`
}
