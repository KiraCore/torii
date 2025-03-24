# Sekai Bridge API Documentation

## Overview

Sekai Bridge is a service that provides threshold signature scheme (TSS) functionality for cross-chain transactions between Cosmos and Ethereum blockchains. This API allows for key generation, message signing, signature verification, and transaction handling between chains.

All API endpoints require authentication with a valid token.

## Authentication

All requests must include a token in the metadata field. The token must match the configured token in the service.

Example:
```json
{
  "method": "keygen",
  "metadata": {
    "token": "your-access-token"
  }
}
```

## Endpoints

### 1. Keygen

Initiates the generation of threshold keys for the TSS (Threshold Signature Scheme).

- **Endpoint**: `keygen`
- **Method**: POST
- **Description**: Start threshold keys generation
- **Authentication**: Required

**Response:**
- Returns the Y-coordinate of the ECDSA public key on success
- Status code: 200 on success, 500 on error

### 2. Sign

Signs data using the threshold signature scheme.

- **Endpoint**: `sign`
- **Method**: POST
- **Description**: Sign the data
- **Authentication**: Required

**Request Body:**
```json
{
  "method": "sign",
  "data": {
    // SignMessageRequest as defined by the TSS package
  },
  "metadata": {
    "token": "your-access-token"
  }
}
```

**Response:**
- Returns the signature response object on success
- Status code: 200 on success, 500 on error

### 3. Verify

Verifies a signature against a message.

- **Endpoint**: `verify`
- **Method**: POST
- **Description**: Verify signature
- **Authentication**: Required

**Request Body:**
```json
{
  "method": "verify",
  "data": {
    "msg": "message-to-verify",
    "signature": "base64-encoded-signature"
  },
  "metadata": {
    "token": "your-access-token"
  }
}
```

**Response:**
```json
{
  "is_valid": true|false
}
```
- Status code: 200 on success, 500 on error

### 4. Stats

Returns debug statistics about the P2P network and TSS status.

- **Endpoint**: `stats`
- **Method**: POST
- **Description**: P2P stats
- **Authentication**: Required

**Response:**
- Returns detailed statistics about the TSS and P2P components, including:
    - HTTP and P2P port configurations
    - P2P slot and peer information
    - P2P connection storage status
    - TSS party ID details
    - TSS connection storage
    - Party mapping information
    - Keygen message storage status
- Status code: 200 on success, 500 on error

### 5. Notify

Webhook for handling cross-chain transaction notifications.

- **Endpoint**: `notify`
- **Method**: POST
- **Description**: Notification webhook
- **Authentication**: Required

**Request Body:**
```json
{
  "method": "notify",
  "data": {
    "from": "Cosmos|Ethereum",
    "tx": {
      // Transaction data specific to the source chain
    },
    "signature": "signature-data"
  },
  "metadata": {
    "token": "your-access-token"
  }
}
```

**Cosmos Transaction Format:**
```json
{
  "from": "cosmos-address",
  "to": "ethereum-address",
  "hash": "transaction-hash",
  "amount": "transaction-amount",
  "signature": "transaction-signature"
}
```

**Ethereum Transaction Format:**
```json
{
  "Amount": "big-integer-amount",
  "From": "ethereum-address",
  "Hash": "transaction-hash",
  "Input": {
    "amount": "integer-amount",
    "cyclAddress": "cosmos-address",
    "hash": "input-hash"
  },
  "Number": "block-number",
  "Status": true|false,
  "To": "contract-address",
  "Signature": "transaction-signature"
}
```

**Response:**
- Returns the original data on success
- Status code: 200 on success, 500 on error

### 6. Logs

Retrieves transaction logs for a specific address.

- **Endpoint**: `logs`
- **Method**: POST
- **Description**: Get logs from storage
- **Authentication**: Required

**Request Body:**
```json
{
  "method": "logs",
  "data": {
    "address": "blockchain-address"
  },
  "metadata": {
    "token": "your-access-token"
  }
}
```

**Response:**
- Returns transaction logs from both Cosmos and Ethereum chains for the specified address
- Status code: 200 on success, 500 on error

## Error Handling

All endpoints return appropriate HTTP status codes and error messages:

- 200: Successful operation
- 500: Internal server error, with details in the response

Common error messages include:
- "token doe not valid" (Authentication error)
- "marshaling error" (Request format error)
- "un-marshaling error" (Request parsing error)
- Operation-specific errors (e.g., "keygen error", "sign error")

## Implementation Details

The service uses:
- Threshold Signature Scheme (TSS) for secure cross-chain signatures
- P2P network for node communication
- MongoDB for transaction storage
- JSON for request/response serialization