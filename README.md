<!-- omit in toc -->
# Feather Core

**Feather Core** is a template blockchain built using Cosmos SDK and CometBFT and created with the Feather CLI. It is meant to be used together with **Feather 🪶**.

- [Features](#features)
- [Installing](#installing)
- [Developing](#developing)
- [Configuring](#configuring)
  - [For Development](#for-development)
  - [For Production](#for-production)
- [Publishing](#publishing)
- [Approving Join Requests](#approving-join-requests)
- [Interfaces](#interfaces)

## Features

- All default Cosmos SDK modules
- The Cosmos [IBC](https://ibc.cosmos.network/) module
- The Osmosis [Token Factory](https://github.com/CosmWasm/token-factory) module
- The Terra [Alliance](https://alliance.terra.money/) module

## Installing

Install dependencies:

- `go`
- `jq`
- `make`
- `gcc`
- `feather`

Build and install the chain binary:

```bash
# Run this in the repo root, where this `README.md` lives.
feather dev build
```

## Developing

Check for correctness by running chain simulations:

```bash
# Run in repo root
feather dev simulate
```

Initialize configs to start the chain (for local testing):

```bash
# Run in repo root
feather dev sandbox init

# Start the chain (your binary name may differ)
feather-cored start
```

## Configuring

### For Development

Configure the `config/localnet/config.json` file if you want to test run your chain before deploying to production with Feather.

```js
{
  // Naming convention: `[-a-zA-Z0-9]{3,47}`.
  "chain_id": "localnet-1",
  // List of genesis accounts, with their bank balances at genesis.
  "accounts": [
    {
      "name": "alice",
      "coins": [
        {
          "denom": "token",
          "amount": "20000"
        },
        {
          "denom": "stake",
          "amount": "200000000"
        }
      ]
    },
    {
      "name": "bob",
      "coins": [
        {
          "denom": "token",
          "amount": "10000"
        },
        {
          "denom": "stake",
          "amount": "100000000"
        }
      ]
    }
  ],
  // List of genesis validators, with their staked coins at genesis.
  "validators": [
    {
      "name": "alice",
      "bonded": {
        "denom": "stake",
        "amount": "100000000"
      }
    }
  ]
}
```

### For Production

Configure the `config/mainnet/config.json` if you would like to do any of the following:

1. Change the default bond denom of the chain
2. Change the name of the chain and chain binary
3. Change the genesis account balances
4. Change the address prefixes of user accounts, validator accounts or consensus accounts
5. Change parameters from the `x/staking` module when the chain is deployed by Feather
6. Configure the LCD/RPC/Prometheus endpoints when the chain is deployed by Feather

```js
{
  // Actual chain ID when chain is deployed by Feather.
  // Must uniquely identify your blockchain network.
  // Naming convention: `[-a-zA-Z0-9]{3,47}`.
  "chain_id": "feather-1",
  // Human readable name of the chain.
  "app_name": "feather-core",
  // Metadata registered in the cosmos sdk package (typically `app_name` suffixed with "d").
  "app_binary_name": "feather-cored",
  // Address prefixes for user, validator, and consensus accounts.
  "account_address_prefix": "pfeath",
  "account_pubkey_prefix": "pfeathpub",
  "validator_address_prefix": "pfeathvaloper",
  "validator_pubkey_prefix": "pfeathvaloperpub",
  "consensus_node_address_prefix": "pfeathvalcons",
  "consensus_node_pubkey_prefix": "pfeathvalconspub",
  // Staking bond denominator (i.e. coin denom used for staking).
  "bond_denom": "stake",
  // Amount of `bond_denom` used for staking at genesis.
  // This is split amongst approved genesis validators by Feather.
  "bond_supply": "1000000000",
  // Cooldown time after unbonding from a validator before an account can stake again.
  "unbonding_time": "1814400s",
  // Max number of validators the chain supports.
  "max_validators": 130,
  "max_entries": 7,
  // Minimum commission rate for validators.
  "min_commission_rate": "0",
  // List of genesis accounts, with their bank balances at genesis.
  // Unlike the `accounts` in `config/localnet/config.json`, the `name` field is replaced with the actual `address`.
  "accounts": [
    {
      "address": "pfeath1...aaa",
      "coins": [
        {
          "denom": "stake",
          "amount": "4000000000"
        },
        {
          "denom": "token",
          "amount": "3000000000"
        }
      ]
    }
  ],
  // Whether to start the LCD server with the chain.
  "lcd_enabled": true,
  // Exposes a swagger page documenting exposed API endpoints at IP:<lcd_port>.
  "lcd_swagger_enabled": true,
  // Configurable LCD server port.
  "lcd_port": 1317,
  // Enable this if querying the blockchain from a CORS-enabled app, like web browsers.
  "lcd_enable_unsafe_cors": false,
  // Whether to start the RPC server with the chain.
  "rpc_enabled": true,
  // Configurable RPC server port.
  "rpc_port": 26657,
  // Whether to enable Prometheus metrics.
  "prometheus_enabled": true,
  // Configurable Prometheus port.
  "prometheus_port": 26660
}
```

The total supply of a coin during chain genesis is the sum of the following:

1. The `bond_supply` field
2. The sum of the `amount` fields of the `accounts` array
3. If the coin is the `bond_denom`, the number of validators multiplied by `1000000` (the minimum amount of `bond_denom` required for a validator's self-delegation)

In the example `config.json` above, assuming we have exactly 5 validators, the total supply of `stake` is `5005000000` and the total supply of `token` is `3000000000` at chain genesis.

## Publishing

You must publish your chain with Feather for other Feather users to discover and deploy validators for your chain.

```bash
# Before publishing your chain, it must be publicly available
export REPO="https://github.com/<YourName>/<your-repo>"
git init
git remote add origin $REPO
git branch -M main
git push -u origin main

# Import your existing terra account using your seed phrase (must have LUNA in testnet)
# Request LUNA from testnet https://faucet.terra.money if needed.
export KEY='mykey'
feather config keys add $KEY --recover --hd-path="m/44'/330'/0'/0/0"

# Publish your chain
# export LAUNCH_TIME=$(date -d '+1 day' +%s) # Linux users
export LAUNCH_TIME=$(date -v +1d +%s)        # MacOS users
feather prod chains publish --key $KEY --launch-time $LAUNCH_TIME --repo $REPO

# Note down the printed launch id...

# Query your chain
export LAUNCH_ID="<launch_id>"
feather prod chains show $LAUNCH_ID
```

## Approving Join Requests

Feather validators may submit a [gentx](https://docs.cosmos.network/v0.46/run-node/run-node.html) request to join your chain as a validator. As a chain deployer, you will need to choose which validators to approve for your chain and run the following commands:

```bash
export LAUNCH_ID='<launch_id>'
export KEY='<key>'

# list requests to join a chain for a launch id
feather prod requests pull $LAUNCH_ID

# approve a validator for a given chain launch
export REQUEST_ID='<request_id>'
feather prod requests approve --key $KEY --launch-id $LAUNCH_ID --request-id $REQUEST_ID
```

## Interfaces

Feather Core **must** adhere to this interface to work correctly with Feather. **Feather will not launch your chain if it fails the interface verification.**

- Do not change the root `Makefile`
  - Use a non-standard makefile name with the `-f` or `--file` option if you need one for your own needs
- Do not change the location of the `main` package (must be in `cmd/<app_name>`)
- Do not change the **location** of the `simulation_test.go` file (must be `app/simulation_test.go`)
- Do not change the **contents** of the `simulation_test.go` file
- Do not change the location of the files in the root `config` directory
