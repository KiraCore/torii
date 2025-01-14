# Build
FROM golang as BUILD

WORKDIR /src/cmd/app

COPY ./ /src/

RUN go build -o sai-eth-indexer -buildvcs=false

FROM ubuntu

WORKDIR /srv
RUN apt-get update && apt-get -y install ca-certificates

# Copy binary from build stage
COPY --from=BUILD /src/cmd/app/sai-eth-indexer /srv/

RUN chmod +x /srv/sai-eth-indexer

# Set command to run your binary
CMD /srv/sai-eth-indexer --debug
