**sekai-bridge** uses tss-lib (`https://gitlab.com/thorchain/tss/tss-lib`) to provide Multi-Party Threshold Signature Scheme ECDSA (Elliptic Curve Digital Signature Algorithm) based on Gennaro and Goldfeder 2020 and EdDSA (Edwards-curve Digital Signature Algorithm) 

# SYNOPSIS

## Docker
### Build
`make docker-build`
### Run
`make docker-run`

## Build standalone application
`make build` 

## Run as standalone application
`make run` 


# CONFIGURATION

## config/config.json (application configuration)
- common(http_server,socket_server, web_socket) - common server options for http,socket and web socket servers (sai-service)
- p2p - saiP2P-go settings (port, slot count...)
- udp - udp settings (expected to remain unchanged)
- peers - peer to connect to
- tss - tss settings (pubkey - id for node, parties - parties count, threshold - threshold for keygen, quorum - quorum for signing)
- http - http port
- debug - debug mode
- cache - cache settings for saiP2P-go


# API

## Get stats for node (for debugging purposes)
curl --location --request GET 'http://<host:port>' \
--header 'Content-Type: application/json' \
--data-raw '{"method":"stats"}'

## Keygen
curl --location --request GET 'http://<host:port>' \
--header 'Content-Type: application/json' \
--data-raw '{"method": "keygen", "data": {}}'

## Keysign
curl --location --request GET 'http://<host:port>' \
--header 'Content-Type: application/json' \
--data-raw '{"method": "sign", "data": {"msg":"test"}}'

## Keysign one round
curl --location --request GET '<host:port>' \
--header 'Content-Type: application/json' \
--data-raw '{"method": "sign", "data": {"msg":"test","one_round_signing":true}}'

## Verify signature
curl --location --request GET 'http://<host:port>' \
--header 'Content-Type: application/json' \
--data-raw '{"method": "verify", "data": {"msg":"test",
"signature":"eyJzaWduYXR1cmUiOiJmK013d0NscTl1OGpmNzJnWEFjTnVqTjU2OWdkMTNod0x6QzQwTFB3RzdnOGVjZU50VFUzY1lmQzFRekNwQ0xGZWxjM0Nkcy9OajRqTzNmL0E3R043UT09Iiwic2lnbmF0dXJlX3JlY292ZXJ5IjoiQVE9PSIsInIiOiJmK013d0NscTl1OGpmNzJnWEFjTnVqTjU2OWdkMTNod0x6QzQwTFB3RzdnPSIsInMiOiJQSG5IamJVMU4zR0h3dFVNd3FRaXhYcFhOd25iUHpZK0l6dDMvd094amUwPSIsIm0iOiJNVEF4TUE9PSJ9"}}'