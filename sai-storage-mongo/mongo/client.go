package mongo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"github.com/saiset-co/sai-storage-mongo/logger"
	"github.com/saiset-co/sai-storage-mongo/types"
)

type Client struct {
	Config *types.StorageConfig
	Host   *mongo.Client
	Ctx    context.Context
}

type IndexElement struct {
	Key   string      `json:"key" bson:"key"`
	Value interface{} `json:"value" bson:"value"`
}

type IndexData struct {
	Keys   []bson.M `bson:"keys" json:"keys"`
	Unique bool     `bson:"unique" json:"unique"`
}

type IndexesData []IndexData

func NewMongoClient(config *types.StorageConfig) (*Client, error) {
	var host *mongo.Client
	var hostErr error

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if config.ConnectionString != "" {
		host, _ = mongo.NewClient(options.Client().ApplyURI(config.ConnectionString))

		hostErr = host.Connect(ctx)
	} else {
		switch config.Atlas {
		case false:
			{
				host, _ = mongo.NewClient(options.Client().ApplyURI(
					"mongodb://" + config.Host + ":" + config.Port + "/" + config.Database,
				))

				hostErr = host.Connect(ctx)
			}
		default:
			{
				host, hostErr = mongo.Connect(ctx, options.Client().ApplyURI(
					"mongodb+srv://"+config.User+":"+config.Pass+"@"+config.Host+"/"+config.Database+"?ssl=true&authSource=admin&retryWrites=true&w=majority",
				))
			}
		}
	}

	client := &Client{
		Ctx:    ctx,
		Config: config,
		Host:   host,
	}

	if hostErr != nil {
		return client, hostErr
	}

	return client, nil
}

func (c Client) GetCollection(collectionName string) *mongo.Collection {
	return c.Host.Database(c.Config.Database).Collection(collectionName)
}

func (c Client) FindOne(collectionName string, selector map[string]interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	collection := c.GetCollection(collectionName)
	selector = c.preprocessSelector(selector)
	cur, err := collection.Find(context.TODO(), selector)

	if err != nil {
		logger.Logger.Error("FindOne", zap.Error(err))
		return result, err
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var elem map[string]interface{}
		decodeErr := cur.Decode(&elem)

		if decodeErr != nil {
			logger.Logger.Error("FindOne", zap.Error(decodeErr))
			return result, decodeErr
		}

		result = elem
		break
	}

	if cursorErr := cur.Err(); cursorErr != nil {
		logger.Logger.Error("FindOne", zap.Error(cursorErr))
		return result, cursorErr
	}

	return result, nil
}

func (c Client) Find(collectionName string, selector map[string]interface{}, inputOptions *types.Options, includeFields []string) (*types.FindResult, error) {
	findResult := &types.FindResult{}
	requestOptions := options.Find()

	if inputOptions != nil && inputOptions.Count != 0 {
		collection := c.GetCollection(collectionName)
		selector = c.preprocessSelector(selector)
		count, err := collection.CountDocuments(context.TODO(), selector)
		if err != nil {
			logger.Logger.Error("Find", zap.Error(err))
			return &types.FindResult{}, err
		}
		findResult.Count = count
	}

	if inputOptions != nil && inputOptions.Sort != nil {
		requestOptions.SetSort(inputOptions.Sort)
	}

	if inputOptions != nil && inputOptions.Skip != 0 {
		requestOptions.SetSkip(inputOptions.Skip)
	}

	if inputOptions != nil && inputOptions.Limit != 0 {
		requestOptions.SetLimit(inputOptions.Limit)
	}

	if includeFields != nil {
		projection := bson.D{}
		for _, v := range includeFields {
			projection = append(projection, bson.E{v, 1})
		}
		requestOptions.SetProjection(projection)
	}

	collection := c.GetCollection(collectionName)
	selector = c.preprocessSelector(selector)

	cur, err := collection.Find(context.TODO(), selector, requestOptions)

	if err != nil {
		logger.Logger.Error("Find", zap.Error(err))
		return &types.FindResult{}, err
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var elem map[string]interface{}
		decodeErr := cur.Decode(&elem)

		if decodeErr != nil {
			logger.Logger.Error("Find", zap.Error(decodeErr))
			return &types.FindResult{}, decodeErr
		}

		findResult.Result = append(findResult.Result, elem)
	}

	if cursorErr := cur.Err(); cursorErr != nil {
		logger.Logger.Error("Find", zap.Error(cursorErr))
		return findResult, cursorErr
	}

	return findResult, nil
}

func (c Client) Insert(collectionName string, doc interface{}) (*mongo.InsertOneResult, error) {
	collection := c.GetCollection(collectionName)

	//processedDoc := c.preprocessDoc(doc)
	result, err := collection.InsertOne(context.TODO(), doc)
	if err != nil {
		logger.Logger.Error("Insert", zap.Error(err))
		return nil, err
	}

	return result, nil
}

func (c Client) InsertMany(collectionName string, docs []interface{}) (*mongo.InsertManyResult, error) {
	collection := c.GetCollection(collectionName)

	//processedDoc := c.preprocessDoc(doc)
	result, err := collection.InsertMany(context.TODO(), docs)
	if err != nil {
		logger.Logger.Error("InsertMany", zap.Error(err))
		return nil, err
	}

	return result, nil
}

func (c Client) Update(collectionName string, selector map[string]interface{}, update interface{}) (*mongo.UpdateResult, error) {
	collection := c.GetCollection(collectionName)
	selector = c.preprocessSelector(selector)

	result, err := collection.UpdateMany(context.TODO(), selector, update)
	if err != nil {
		logger.Logger.Error("UpdateMany", zap.Error(err))
		return nil, err
	}

	return result, nil
}

func (c Client) Upsert(collectionName string, selector map[string]interface{}, update interface{}) error {
	collection := c.GetCollection(collectionName)
	requestOptions := options.Update().SetUpsert(true)
	selector = c.preprocessSelector(selector)

	_, err := collection.UpdateMany(context.TODO(), selector, update, requestOptions)
	if err != nil {
		logger.Logger.Error("Upsert", zap.Error(err))
		return err
	}

	return nil
}

func (c Client) Remove(collectionName string, selector map[string]interface{}) error {
	collection := c.GetCollection(collectionName)
	selector = c.preprocessSelector(selector)

	_, err := collection.DeleteMany(context.TODO(), selector)
	if err != nil {
		logger.Logger.Error("Remove", zap.Error(err))
		return err
	}

	return nil
}

func (c Client) Aggregate(collectionName string, pipeline interface{}) (*types.FindResult, error) {
	findResult := &types.FindResult{}
	collection := c.GetCollection(collectionName)

	cur, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		logger.Logger.Error("Aggregate", zap.Error(err))
		return nil, err
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var elem map[string]interface{}
		decodeErr := cur.Decode(&elem)

		if decodeErr != nil {
			logger.Logger.Error("Aggregate", zap.Error(decodeErr))
			return &types.FindResult{}, decodeErr
		}

		findResult.Result = append(findResult.Result, elem)
	}

	if cursorErr := cur.Err(); cursorErr != nil {
		logger.Logger.Error("Aggregate", zap.Error(cursorErr))
		return findResult, cursorErr
	}

	return findResult, nil
}

func (c Client) CreateIndexes(collectionName string, data interface{}) ([]string, error) {
	collection := c.GetCollection(collectionName)
	var _data []IndexData

	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Logger.Error("CreateIndexes", zap.Error(err))
		return nil, err
	}

	err = json.Unmarshal(jsonData, &_data)
	if err != nil {
		logger.Logger.Error("CreateIndexes", zap.Error(err))
		return nil, err
	}

	var indexes []mongo.IndexModel

	for _, indexValue := range _data {
		doc := bson.D{}

		for _, v := range indexValue.Keys {
			for _i, _v := range v {
				value, ok := _v.(float64)
				if !ok {
					return nil, errors.New("index value not an integer")
				}

				doc = append(doc, bson.E{
					Key:   _i,
					Value: int64(value),
				})
			}
		}

		indexModel := mongo.IndexModel{
			Keys: doc,
		}

		if indexValue.Unique {
			indexModel.Options = options.Index().SetUnique(true)
		}

		indexes = append(indexes, indexModel)
	}

	result, err := collection.Indexes().CreateMany(context.TODO(), indexes)
	if err != nil {
		logger.Logger.Error("CreateIndexes", zap.Error(err))
		return nil, err
	}

	return result, nil
}

func (c Client) GetIndexes(collectionName string) ([]interface{}, error) {
	var result []interface{}
	collection := c.GetCollection(collectionName)

	cur, err := collection.Indexes().List(context.TODO())
	if err != nil {
		logger.Logger.Error("GetIndexes", zap.Error(err))
		return result, err
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var index interface{}
		decodeErr := cur.Decode(&index)

		if decodeErr != nil {
			return result, decodeErr
		}

		result = append(result, index)
		break
	}

	if cursorErr := cur.Err(); cursorErr != nil {
		logger.Logger.Error("GetIndexes", zap.Error(err))
		return result, cursorErr
	}

	return result, nil
}

func (c Client) DropIndexes(collectionName string) ([]interface{}, error) {
	var result []interface{}
	collection := c.GetCollection(collectionName)

	_, err := collection.Indexes().DropAll(context.TODO())
	if err != nil {
		logger.Logger.Error("DropIndexes", zap.Error(err))
		return result, err
	}

	return result, nil
}

func (c Client) preprocessSelector(selector map[string]interface{}) map[string]interface{} {
	if selector["_id"] != nil {
		switch selector["_id"].(type) {
		case string:
			objID, err := primitive.ObjectIDFromHex(selector["_id"].(string))
			if err != nil {
				logger.Logger.Error("preprocessSelector", zap.Error(err))
				return selector
			}
			selector["_id"] = objID
		case map[string]interface{}:
			objIDslice := make([]primitive.ObjectID, 0)
			m := selector["_id"].(map[string]interface{})
			for k, v := range m {
				switch v.(type) {
				case []interface{}:
					for _, s := range v.([]interface{}) {
						objID, err := primitive.ObjectIDFromHex(s.(string))
						if err != nil {
							continue
						}
						objIDslice = append(objIDslice, objID)
					}
					m[k] = objIDslice
				default:
					return selector
				}

			}
			selector["_id"] = m
		default:
			return selector
		}
	}
	return selector
}

// preprocess doc (insert method)
func (c Client) preprocessDoc(doc interface{}) primitive.M {
	m, ok := doc.(primitive.M)
	if !ok {
		return nil
	}
	if m["type"] == "refresh_token" {
		accessToken := m["access_token"].(map[string]interface{})
		id := accessToken["_id"].(string)
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			logger.Logger.Error("preprocessDoc", zap.Error(err))
			return nil
		}
		accessToken["_id"] = objID
		return m

	} else {
		return m
	}
}
