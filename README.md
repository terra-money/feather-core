# Feather Core

**Feather Core** is a blockchain built using Cosmos SDK and CometBFT and created with the Feather CLI. It is mean to be used together with **Feather CLI**.

## Installation

Build and install the chain binary:

```bash
# Run this in the repo root, where README.md lives.
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

# Then, start the chain
feather-cored start
```

## Configuration

### For Development

Configure the `config/localnet/config.json` file if you want to test run your chain before deploying to production with Feather.

```json
{
    // Should uniquely identify your blockchain network.
    // Naming convention: `[-a-zA-Z0-9]{3,47}`
    "chain_id": "test-chain",
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

1. Change the default bond denom of the chain.
2. Change the name of the chain and chain binary.
3. Change the address prefixes of user accounts, validator accounts or consensus accounts.
4. Change parameters from the `x/staking` module when the chain is deployed by Feather.
5. Configure the LCD/RPC/Prometheus endpoints when the chain is deployed by Feather.

```json
{
    // Actual chain id when chain is deployed by Feather.
    "chain_id": "feather-1",
    // Name of the chain and chain binary (suffixed with 'd').
    "app_name": "feather-core",
    // Metadata registered in the cosmos sdk package.
    "app_binary_name": "feather-cored",
    // Address prefixes
    "account_address_prefix": "pfeath",
    "account_pubkey_prefix": "pfeathpub",
    "validator_address_prefix": "pfeathvaloper",
    "validator_pubkey_prefix": "pfeathvaloperpub",
    "consensus_node_address_prefix": "pfeathvalcons",
    "consensus_node_pubkey_prefix": "pfeathvalconspub",
    // Staking bond denominator (i.e. coin currency used for staking)
    "bond_denom": "stake",
    // Amount of bond denom for staking available at genesis.
    // This is split among approved genesis validators by Feather.
    "bond_supply": "1000000000",
    // Cooldown time after unbonding from a validator before an account can stake again.
    "unbonding_time": "1814400s",
    // Max number of validators the chain supports.
    "max_validators": 130,
    "max_entries": 7,
    "min_commission_rate": "0",
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
