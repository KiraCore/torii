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

# Step 3: Final image
FROM alpine:3.17
# Install CA certificates
RUN apk --no-cache add ca-certificates && \
    update-ca-certificates

# Create directory for the application
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /bin/app /app/app
# Copy the config file
COPY --from=builder /app/config.yml /app/config.yml
# If you have additional config files, copy them too
# COPY --from=builder /app/additional-configs/ /app/additional-configs/

RUN chmod +x /app/app

CMD ["/app/app", "start"]