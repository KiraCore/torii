## make binaries and dirs for local test
local-test-build:
## build first local instance
	@rm -rf dist
	@mkdir dist
	GOOS=${OSFLAG} GOARCH=${OSHW} go build -o dist/grape *.go
	@cp config.yml dist/

## build second local instance
	@rm -rf dist2
	@mkdir dist2
	@cp config2.yml dist2/
	@cp dist/grape dist2/grape
	@ mv dist2/config2.yml dist2/config.yml

## build third local instance
	@rm -rf dist3
	@mkdir dist3
	@cp dist/grape dist3/grape
	@cp config3.yml dist3/
	@ mv dist3/config3.yml dist3/config.yml

## build forth local instance
	@rm -rf dist4
	@mkdir dist4
	@cp dist/grape dist4/grape
	@cp config4.yml dist4/
	@ mv dist4/config4.yml dist4/config.yml

