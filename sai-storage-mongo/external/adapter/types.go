package adapter

type IWrapperRequest interface {
	GetMethod() string
	GetData() IRequest
}

type IRequest interface {
	GetCollection() string
	GetSelect() map[string]interface{}
	GetData() []interface{}
	GetOptions() *Options
	GetIncludeFields() []string
}

type Request struct {
	Method string   `json:"method"`
	Data   IRequest `json:"data"`
}

type Options struct {
	Limit int64       `json:"limit"`
	Skip  int64       `json:"skip"`
	Sort  interface{} `json:"sort"`
	Count int64       `json:"count"`
}

func (r Request) GetMethod() string {
	return r.Method
}

func (r Request) GetData() IRequest {
	return r.Data
}

type ReadRequest struct {
	Collection    string                 `json:"collection" validate:"required"`
	Select        map[string]interface{} `json:"select,omitempty" validate:"required"`
	Options       *Options               `json:"options"`
	IncludeFields []string               `json:"include_fields"`
}

type CreateRequest struct {
	Collection    string        `json:"collection" validate:"required"`
	Documents     []interface{} `json:"documents,omitempty" validate:"required"`
	Options       *Options      `json:"options"`
	IncludeFields []string      `json:"include_fields"`
}

type UpdateRequest struct {
	Collection    string                 `json:"collection" validate:"required"`
	Select        map[string]interface{} `json:"select,omitempty" validate:"required"`
	Document      interface{}            `json:"document,omitempty" validate:"required"`
	Options       *Options               `json:"options"`
	IncludeFields []string               `json:"include_fields"`
}

type UpsertRequest struct {
	Collection    string                 `json:"collection" validate:"required"`
	Select        map[string]interface{} `json:"select,omitempty" validate:"required"`
	Document      interface{}            `json:"document,omitempty" validate:"required"`
	Options       *Options               `json:"options"`
	IncludeFields []string               `json:"include_fields"`
}

type AggregateRequest struct {
	Collection string        `json:"collection" validate:"required"`
	Pipeline   []interface{} `json:"pipeline,omitempty" validate:"required"`
}

type DeleteRequest struct {
	Collection string                 `json:"collection" validate:"required"`
	Select     map[string]interface{} `json:"select,omitempty" validate:"required"`
}

func (r ReadRequest) GetCollection() string {
	return r.Collection
}

func (r ReadRequest) GetSelect() map[string]interface{} {
	return r.Select
}

func (r ReadRequest) GetData() []interface{} {
	return nil
}

func (r ReadRequest) GetOptions() *Options {
	return r.Options
}

func (r ReadRequest) GetIncludeFields() []string {
	return r.IncludeFields
}

func (r CreateRequest) GetCollection() string {
	return r.Collection
}

func (r CreateRequest) GetSelect() map[string]interface{} {
	return nil
}

func (r CreateRequest) GetData() []interface{} {
	return r.Documents
}

func (r CreateRequest) GetOptions() *Options {
	return r.Options
}

func (r CreateRequest) GetIncludeFields() []string {
	return r.IncludeFields
}

func (r UpdateRequest) GetCollection() string {
	return r.Collection
}

func (r UpdateRequest) GetSelect() map[string]interface{} {
	return r.Select
}

func (r UpdateRequest) GetData() []interface{} {
	return []interface{}{r.Document}
}

func (r UpdateRequest) GetOptions() *Options {
	return r.Options
}

func (r UpdateRequest) GetIncludeFields() []string {
	return r.IncludeFields
}

func (r UpsertRequest) GetCollection() string {
	return r.Collection
}

func (r UpsertRequest) GetSelect() map[string]interface{} {
	return r.Select
}

func (r UpsertRequest) GetData() interface{} {
	return r.Document
}

func (r UpsertRequest) GetOptions() *Options {
	return r.Options
}

func (r UpsertRequest) GetIncludeFields() []string {
	return r.IncludeFields
}

func (r AggregateRequest) GetCollection() string {
	return r.Collection
}

func (r AggregateRequest) GetSelect() map[string]interface{} {
	return nil
}

func (r AggregateRequest) GetData() interface{} {
	return r.Pipeline
}

func (r AggregateRequest) GetOptions() *Options {
	return nil
}

func (r AggregateRequest) GetIncludeFields() []string {
	return nil
}

func (r DeleteRequest) GetCollection() string {
	return r.Collection
}

func (r DeleteRequest) GetSelect() map[string]interface{} {
	return r.Select
}

func (r DeleteRequest) GetData() []interface{} {
	return nil
}

func (r DeleteRequest) GetOptions() *Options {
	return nil
}

func (r DeleteRequest) GetIncludeFields() []string {
	return nil
}
