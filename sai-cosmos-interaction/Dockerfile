# Build
FROM golang as BUILD

WORKDIR /src

COPY ./ /src/

RUN go build -o sai-cosmos-interaction -buildvcs=false

FROM ubuntu

WORKDIR /srv
RUN apt-get update && apt-get -y install ca-certificates

# Copy binary from build stage
COPY --from=BUILD /src/sai-cosmos-interaction /srv/

RUN chmod +x /srv/sai-cosmos-interaction

# Set command to run your binary
CMD /srv/sai-cosmos-interaction start
