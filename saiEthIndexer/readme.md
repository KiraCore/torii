# saiEthIndexer

Utility for viewing transactions of specified addresses in ETH SDK based blockchains.
If added address found in the transaction, this transaction will be saved to the storage and sent to notification address.

## Configurations
**config.json** - config file.

### Common block
- `http_server` - http server section
    - `enabled` - enable or disable http handlers
    - `port`    - http server port

### Specific block
- `geth_server` - ETH server url
- `storage` - sai-storage http server
  - `url` - sai-storage http server address
  - `token` - sai-storage token
  - `collection` - sai-storage collection name
- `notifier` - bridge service url
  - `url` - bridge service url
  - `token` - bridge service token
  - `sender_id` - sender name: CYCLONE
- `start_block` - start block height
- `sleep` - sleep duration between loop iteration(in seconds)
- `skipFailedTransactions` - TRUE to skip not parsed transaction

## How to run
`make build`: rebuild and start service  
`make up`: start service  
`make down`: stop service  
`make logs`: display service logs

## API
### Add addresses <host:port>/v1/add_contract
```json lines
{
  "contracts": [
    {
      "address": "$address",
      "abi": "$abi",
      "start_block":$start_block
    },
    {
      "address": "$address",
      "abi": "$abi",
      "start_block":$start_block
    }
  ]
}
```
#### Params
`$address` <- any contract address to find in transaction
`$abi` <- abi string quoted
`$start_block` <- block height to start indexation

### Delete addresses <host:port>/v1/delete_contract
```json lines
{
  "addresses": [
    "$address","$address"
  ]
}
```
#### Params
`$address` <- any contract address to find in transaction
