## make binaries and dirs for local test
local-test-build:
## build first local instance
	@rm -rf dist
	@mkdir dist
	GOOS=${OSFLAG} GOARCH=${OSHW} go build -o dist/sekai-bridge *.go
	@cp config.yml dist/

## build second local instance
	@rm -rf dist2
	@mkdir dist2
	@cp config2.yml dist2/
	@ mv dist2/config2.yml dist2/config.yml
	@cp dist/sekai-bridge dist2/sekai-bridge

## build third local instance
#	@rm -rf dist3
#	@mkdir dist3
#	@cp dist/sekai-bridge dist3/sekai-bridge
#	@cp config3.yml dist3/
#	@ mv dist3/config3.yml dist3/config.yml


### check by golangci linter
linter: 
	golangci-lint run --config=.golangci.yml ./...