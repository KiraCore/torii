# Build stage
ARG SERVICE=sai-eth-interaction

FROM golang as BUILD

ARG SERVICE

WORKDIR /src/

COPY ./ /src/

RUN go build -o $SERVICE -buildvcs=false

FROM ubuntu

ARG SERVICE

WORKDIR /srv

# Copy binary from build stage
COPY --from=BUILD /src/$SERVICE /srv/$SERVICE

# Copy other files
COPY ./config.yml /srv/config.yml

RUN chmod +x /srv/$SERVICE

# Set command to run your binary
CMD /srv/$SERVICE start
