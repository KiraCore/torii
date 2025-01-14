# Microservice mongo storage

## Ports
`8880`: **HTTP**  
`8881`: WebSocket   
`8882`: Socket

## How to run
`make build`: rebuild and start service  
`make up`: start service  
`make down`: stop service  
`make logs`: display service logs

## API requests
### CREATE
Create multiple documents in a collection

```curl
curl --request GET \
  --url http://localhost:8880/ \
  --header 'Content-Type: application/json' \
  --data '{
	"method": "create",
	"data": {
		"collection": "CollectionName",
		"documents": [{}], //<- array of structs to save in a DB
	}
}'
```

### READ
Return filtered documents. Can be paginated with skip\limit.

```curl
curl --request GET \
  --url http://localhost:8880/ \
  --header 'Content-Type: application/json' \
  --data '{
	"method": "read",
	"data": {
		"collection": "CollectionName",
		"select":{} //<-mongo select request format
	},
	"metadata": {
	   "skip": 10, //<- to skip first 10 elements
	   "limit": 10, //<- to limit result with 10 items
	   "count": 1 //<- to have total count in the result without pagination data
	}
}'
```

### UPDATE / UPSERT
Update: update all filtered documents.
Upsert: update all or create a new document if nothing to update.

```curl
curl --request GET \
  --url http://localhost:8880/ \
  --header 'Content-Type: application/json' \
  --data '{
	"method": "update", //<- Or "upsert"
	"data": {
		"collection": "CollectionName",
		"select":{} //<-mongo select request format,
		"document": {}, //<- struct to save in a DB
	}
}'
```

### DELETE
Delete all filtered documents.

```curl
curl --request GET \
  --url http://localhost:8880/ \
  --header 'Content-Type: application/json' \
  --data '{
	"method": "delete",
	"data": {
		"collection": "CollectionName",
		"select":{} //<-mongo select request format
	}
}'
```