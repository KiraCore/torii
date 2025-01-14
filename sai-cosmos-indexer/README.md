# saiCosmosIndexer

Utility for viewing transactions of specified addresses
in Cosmos SDK based blockchains.

==========================================

### configurations

**config.yml** - common saiService config file.
You can specific port for http or ws protocol if you need.

**service_config.json** - the `storage` section
is intended to indicate storage authorization data and the name
of the collection for storing information about transactions.

- `node_address` - node address for API calls
- `start_block` - start block height
- `tx_type` - transactions type for scanning
- `sleep_duration` - sleep duration between loop iteration(in seconds)

**latest_handled_block** - file to save latest handled block for reboot cases
(created and overwritten automatically)

### handlers

- `add_address` - Add new address for scan transactions (save address to **./addresses.json**)
- `delete_address` - Delete address from addresses list (delete address from **./addresses.json**)

Example:

`curl --location 'localhost:8080' --header 'Content-Type: application/json' --data '{
"method": "delete_address",
"data": "some_address"
}'`
