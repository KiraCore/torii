# Step 1: Modules caching
FROM golang:1.21.4-alpine3.17 as modules
COPY go.mod go.sum /modules/
COPY /saiP2P-go /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Builder
FROM golang:1.21.4-alpine3.17 as builder
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build  -o /bin/app .

RUN chmod +x /bin/app

CMD /bin/app start

EXPOSE 8080
EXPOSE 9080
EXPOSE 9000