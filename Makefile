## Build
build:
	GOOS=${OSFLAG} GOARCH=${OSHW} go build -o sekai-bridge *.go

## Run
run:
	./sekai-bridge start	

## make binaries and dirs for local test
local-test-build:
## build first local instance
	@rm -rf dist
	@mkdir dist
	GOOS=${OSFLAG} GOARCH=${OSHW} go build -o dist/sekai-bridge *.go
	@cp config.yml dist/
	@cp key.json dist/ || true

## build second local instance
	@rm -rf dist2
	@mkdir dist2
	@cp config2.yml dist2/
	@ mv dist2/config2.yml dist2/config.yml
	@cp dist/sekai-bridge dist2/sekai-bridge
	@cp key2.json dist2/ || true
	@mv dist2/key2.json dist2/key.json

## build third local instance
	@rm -rf dist3
	@mkdir dist3
	@cp dist/sekai-bridge dist3/sekai-bridge
	@cp config3.yml dist3/
	@ mv dist3/config3.yml dist3/config.yml
	@cp key3.json dist3/ || true
	@mv dist3/key3.json dist3/key.json

## build forth local instance
	@rm -rf dist4
	@mkdir dist4
	@cp dist/sekai-bridge dist4/sekai-bridge
	@cp config4.yml dist4/
	@ mv dist4/config4.yml dist4/config.yml
	@cp key4.json dist4/ || true
	@mv dist4/key4.json dist4/key.json


## check by golangci linter
linter: 
	golangci-lint run --config=.golangci.yml ./...


### docker build
docker-build:
	docker build --tag 'sekai-bridge' .

### run docker image
docker-run:
	docker run sekai-bridge

### stop docker image
docker-run:
	docker run sekai-bridge