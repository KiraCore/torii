FROM golang as BUILD
WORKDIR /src/
COPY ./ /src/
RUN go build -o sai-eth-interaction -buildvcs=false

FROM ubuntu
WORKDIR /srv

# Install and update CA certificates
RUN apt-get update && \
    apt-get install -y ca-certificates && \
    update-ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Copy binary from build stage
COPY --from=BUILD /src/sai-eth-interaction /srv/sai-eth-interaction

# Copy other files (uncomment if needed)
#COPY ./config.yml /srv/config.yml

RUN chmod +x /srv/sai-eth-interaction

# Set command to run your binary
CMD /srv/sai-eth-interaction start