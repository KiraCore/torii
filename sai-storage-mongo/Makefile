SERVICE_NAME=sai-storage
EXTERNAL_PORT=8880
PORT=8880

build:
	cp config.yml config-template.yml
	set -a && . ./.env.dev && set +a && envsubst < config-template.yml > config.yml
	docker-compose up -d --build
	cp config-template.yml config.yml
	rm config-template.yml

up:
	cp config.yml config-template.yml
	set -a && . ./.env.dev && set +a && envsubst < config-template.yml > config.yml
	docker-compose up -d
	cp config-template.yml config.yml
	rm config-template.yml

sh:
	docker-compose exec ${SERVICE_NAME} bash

log:
	docker-compose logs -f ${SERVICE_NAME}

down:
	docker-compose down

test:
	go test ./tests -run TestStart -count=1

docker:
	docker build -t ${SERVICE_NAME} .
	docker stop ${SERVICE_NAME} || true
	docker rm ${SERVICE_NAME} || true
	docker run -d -p ${EXTERNAL_PORT}:${PORT} --restart unless-stopped --name ${SERVICE_NAME} ${SERVICE_NAME}
