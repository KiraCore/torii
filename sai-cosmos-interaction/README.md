# saiCosmosInteraction


Utility for creating transactions in Cosmos SDK based blockchains.

==========================================

### handlers

- `make_tx` - Make new transaction. You need to have a file with a private key on the server side named the sender's address


Example:

`curl --location 'localhost:8080' \
--header 'Content-Type: application/json' \
--data '{
"method": "make_tx",
"data": {
"node_address": "https://rest.sentry-01.theta-testnet.polypore.xyz",
"from": "sender",
"to": "recipient",
"chain_id": "theta-testnet-001",
"memo": "my first transasction test memo",
"amount": 100000,
"gas_limit": 100000,
"fee_amount": 750,
"passphrase": "your_passphrase"
}
}'`
