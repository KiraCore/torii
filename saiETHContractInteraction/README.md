## saiETHInteraction

Http proxy to create transaction to the ETH contracts.

## Configurations
**config.yml** - common saiService config file.

### Common block
- `http_server` - http server section
    - `enabled` - enable or disable http handlers
    - `port`    - http server port

### Specific block
- `eth_server` - ETH server url

## How to run
`make build`: rebuild and start service   
`make up`: start service  
`make down`: stop service  
`make logs`: display service logs

## API
### Contract interaction
```json lines
{
  "method": "api",
  "data": {
    "contract":"$name",
    "method":"$contract_method_name", 
    "value": "$value", 
    "params":[
      {
        "type":"$(int|string|float...)",
        "value":"$some_value"
      }
    ]
  }
}
```
#### Params
`$name` <- contract name  
`$contract_method_name` <- contract method name  
`$value` <- string, value  
`$some_value` <- string, value  

### Add contracts
```json lines
{
  "contracts": [
    {
      "name":"$name", 
      "server": "$server",
      "address":"$address",
      "abi":"$abi", 
      "private": "$private", 
      "gas_limit":$gas_limit
    }
  ]
}
```

#### Params
`$name` <- contract name  
`$server` <- ETH server url  
`$address` <- contract address  
`$abi` <- abi encoded  
`$private` <- private key to sign transactions  
`$gas_limit` <- gas limit value  

### Delete contracts
```json lines
{
  "names": [
    "$name"
  ]
}
```

#### Params
`$name` <- contract name
