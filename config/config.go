package config

import (
	_ "embed"
	"encoding/json"
)

type Config struct {
	ChainID              string    `json:"chain_id"`
	AppName              string    `json:"app_name"`
	AccountAddressPrefix string    `json:"account_address_prefix"`
	BondDenom            string    `json:"bond_denom"`
	BondSupply           string    `json:"bond_supply"`
	UnbondingTime        string    `json:"unbonding_time"`
	MaxValidators        int       `json:"max_validators"`
	MaxEntries           int       `json:"max_entries"`
	MinCommissionRate    string    `json:"min_commission_rate"`
	Accounts             []Account `json:"accounts"`
}

type Account struct {
	Address string `json:"address"`
	Coins   []Coin `json:"coins"`
}

type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

//go:embed config.json
var configJson []byte

func Load() (Config, error) {
	cfg := Config{}
	err := json.Unmarshal(configJson, &cfg)
	return cfg, err
}
