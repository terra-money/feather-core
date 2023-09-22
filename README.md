<!-- omit in toc -->
# Feather Core

**Feather Core** is a template blockchain built using Cosmos SDK and CometBFT and created with the Feather CLI. It is meant to be used together with **Feather 🪶**.

- [Features](#features)
- [Installing](#installing)
- [Developing](#developing)
- [Configuring](#configuring)
  - [Genesis Validator Delegations](#genesis-validator-delegations)
  - [Genesis Account Balances](#genesis-account-balances)
  - [Genesis Coins Total Supply](#genesis-coins-total-supply)
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
# Run in repo root:
feather dev simulate
```

Initialize and run a localnet with a single local validator for testing purposes:

```bash
# Run in repo root:
feather dev sandbox serve

# OR, if you prefer to be explicit:
feather dev sandbox init  # initialize all genesis files
feather-cored start       # start the localnet (your built binary name may differ)
```

## Configuring

Edit the `config/config.json` file if you would like to do any of the following:

1. Change the default bond denom of the chain
2. Change the chain ID
3. Change the genesis account balances
4. Change the address prefixes of user accounts, validator accounts, and consensus accounts

```js
{
  // Actual chain ID when chain is deployed by Feather.
  // Must uniquely identify your blockchain network.
  // Naming convention: `[-a-zA-Z0-9]{3,47}`.
  "chain_id": "feather-1",
  // Human readable name of the chain.
  "app_name": "feather-core",
  // Prefix for all addresses on the chain.
  "address_prefix": "feath",
  // Staking bond denominator (i.e. coin denom used for staking).
  "bond_denom": "stake",
  // Amount of `bond_denom` used for staking at genesis.
  // This is split amongst approved genesis validators by Feather.
  "bond_supply": "1000000000",
  // Cooldown time after unbonding from a validator before an account can stake again.
  "unbonding_time": "1814400s",
  // Max number of validators the chain supports.
  "max_validators": 130,
  // Minimum commission rate for validators. Must be a number between 0 and 1.
  "min_commission_rate": "0",
  // List of genesis accounts, with their balances at genesis.
  "accounts": [
    {
      "address": "feath1...aaa",
      "coins": [
        {
          "denom": "stake",
          "amount": "4000000000"
        },
        {
          "denom": "token",
          "amount": "3000000000"
        }
      ],
      // Optional vesting configurations
      "vesting": {
        "coins": [
          {
            "denom": "stake",      // Must already exist in the root `coins` section
            "amount": "1000000000" // Must be less than the amount defined in the root `coins` section
          }
        ],
        "start": 1690000000, // Unix time; optional
        "end": 1700000000    // Unix time; must be larger than `start`
      }
    }
  ]
}
```

### Genesis Validator Delegations

At genesis, all validators will have an initial *self-delegation* of exactly `1000000` of the `bond_denom` (ie. `stake`). This is to satisfy Cosmos SDK's requirement of having at least one validator with a total delegation of at least `1000000` of the `bond_denom` during chain genesis.

Additionally, all `bond_supply` (owned by the address of the chain deployer) will be split according to the stake distribution setting specified and delegated to genesis validators. There are currently two stake distribution settings:

1. `equal`: All genesis validators receive an equal amount of delegations (ie. $\frac{\texttt{bond\_supply}}{\texttt{num\_genesis\_validators}}$) from the chain deployer
2. `terravp`: All validators receive delegations proportional to the validator's voting power in the Terra chain from the chain deployer
   - A validator that has no voting power in the Terra chain will thus NOT receive any delegations from the chain deployer

In summary, the total delegation or voting power of a validator at chain genesis when using the `equal` stake distribution strategy is:

$$
1000000 + \left( \frac{\texttt{bond\_supply}}{\texttt{num\_genesis\_validators}}\right)
$$

And when using the `terravp` stake distribution strategy:

$$
1000000 + \left( {\texttt{bond\_supply}} \times \frac{\texttt{validator\_voting\_power}}{\texttt{total\_voting\_power}} \right)
$$

### Genesis Account Balances

The `accounts` array in `config.json` specifies the genesis accounts and their balances at chain genesis. These balances are **in addition** of the `bond_supply` field or any validator's self-delegations. In other words, a chain deployer can have a `bond_supply` of `1000000000` and still have an additional genesis account with a balance of `5000000000` of the `bond_denom` (ie. `stake`) specified, meaning they will own a total of `6000000000` of the `bond_denom` at chain genesis. Likewise, a validator will own a total of `1000000` of the `bond_denom` (used for self-delegation) and whatever is specified in the `accounts` array.

An account can also contain an optional `vesting` key, which defines the vesting strategy of the various coins of that account. Note that the coins defined in here are a subset of the main coins of the account. In other words, the coins defined in the `vesting` key must already exist, and must be less than the coins defined directly outside of the `vesting` key. The `start` key is optional (defaults to the start time of the chain), but the `end` key is compulsory, and they both accept Unix time only.

### Genesis Coins Total Supply

The total supply of a coin at chain genesis is the sum of the following:

1. The sum of the `amount` fields of the `accounts` array whose `denom` field matches the coin's denom
2. If the coin is the `bond_denom`, the `bond_supply` field
3. If the coin is the `bond_denom`, the number of genesis validators multiplied by `1000000` (the minimum amount of `bond_denom` required for a validator's self-delegation)

In the example `config.json` above, assuming we have exactly 8 validators, the total supply of `stake` is `5008000000` and the total supply of `token` is `3000000000` at chain genesis.

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
