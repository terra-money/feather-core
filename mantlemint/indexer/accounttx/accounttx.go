package accounttx

import (
	"encoding/json"
	"fmt"
	tmdb "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	tm "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	terra "github.com/terra-money/alliance/app"
	"github.com/terra-money/feather-core/app"

	"github.com/terra-money/feather-core/mantlemint/db/safe_batch"
	"github.com/terra-money/feather-core/mantlemint/indexer"
	"github.com/terra-money/feather-core/mantlemint/indexer/tx"
	"github.com/terra-money/feather-core/mantlemint/mantlemint"
)

var cdc = app.MakeEncodingConfig()

var IndexTx = indexer.CreateIndexer(func(db safe_batch.SafeBatchDB, block *tm.Block, blockID *tm.BlockID, evc *mantlemint.EventCollector, _ *app.App) error {
	for i, txByte := range block.Txs {
		txRecord := evc.ResponseDeliverTxs[i]
		addrsInTx, err := parseTxBytesForAddresses(txByte)
		if err != nil {
			return err
		}
		addrsInEvents, err := parseEventsForAddresses(txRecord.Events)
		if err != nil {
			return err
		}

		for k, _ := range addrsInEvents {
			addrsInTx[k] = true
		}

		for addr, _ := range addrsInTx {
			key := GetAccountTxKey(addr, uint64(block.Height), uint64(i))
			accountTx := AccountTx{
				TxHash:      fmt.Sprintf("%X", txByte.Hash()),
				BlockHeight: uint64(block.Height),
				Timestamp:   block.Time,
			}
			b, err := json.Marshal(accountTx)
			if err != nil {
				return err
			}
			err = db.Set(key, b)
			if err != nil {
				return err
			}
		}
	}
	return nil
})

func parseTxBytesForAddresses(txByte []byte) (addrs map[string]bool, err error) {
	// Use a map to collect unique addresses
	addrs = make(map[string]bool)

	// Decode to Tx struct
	tx, err := cdc.TxConfig.TxDecoder()(txByte)
	if err != nil {
		return addrs, err
	}

	wrappedTx, err := cdc.TxConfig.WrapTxBuilder(tx)
	if err != nil {
		return addrs, err
	}

	signers := wrappedTx.GetTx().GetSigners()
	for _, signer := range signers {
		addr, err := sdk.Bech32ifyAddressBytes(terra.AccountAddressPrefix, signer)
		if err != nil {
			return addrs, err
		}
		addrs[addr] = true
	}

	// Encode to JSON
	jsonByte, err := cdc.TxConfig.TxJSONEncoder()(tx)
	if err != nil {
		return addrs, err
	}
	// Decode to generic interface to find addresses
	var txRaw map[string]interface{}
	err = json.Unmarshal(jsonByte, &txRaw)
	if err != nil {
		return addrs, err
	}
	bodyRaw, found := txRaw["body"]
	if !found {
		return addrs, nil
	}

	body, ok := bodyRaw.(map[string]interface{})
	if !ok {
		return addrs, fmt.Errorf("unable to coerce tx body into map")
	}

	var findAddresses func(o interface{})
	findAddresses = func(o interface{}) {
		stringValue, isString := o.(string)
		if isString {
			_, err := sdk.GetFromBech32(stringValue, terra.AccountAddressPrefix)
			if err == nil {
				addrs[stringValue] = true
			}
			return
		}

		mapValue, isMap := o.(map[string]interface{})
		if isMap {
			for _, v := range mapValue {
				findAddresses(v)
			}
			return
		}

		arrayValue, isArray := o.([]interface{})
		if isArray {
			for _, a := range arrayValue {
				findAddresses(a)
			}
		}
	}

	msgsRaw, found := body["messages"]
	if !found {
		return addrs, nil
	}
	findAddresses(msgsRaw)
	return addrs, nil
}

func parseEventsForAddresses(events []abci.Event) (addrs map[string]bool, err error) {
	addrs = make(map[string]bool)
	for _, event := range events {
		attrs := event.GetAttributes()
		for _, attr := range attrs {
			addrStr := string(attr.GetValue())
			_, err := sdk.GetFromBech32(addrStr, terra.AccountAddressPrefix)
			if err == nil {
				addrs[addrStr] = true
			}
		}
	}
	return addrs, nil
}

func getTxnsByAccount(db tmdb.DB, account string, offset uint64, limit uint64) (txs []tx.TxByHeightRecord, err error) {
	key := GetAccountTxKeyByAddr(account)
	iter, err := db.Iterator(key, sdk.PrefixEndBytes(key))
	if err != nil {
		return txs, err
	}
	currentOffset := uint64(0)
	currentLimit := uint64(0)
	for iter.Valid() {
		if currentOffset < offset {
			currentOffset += 1
			continue
		}
		var accountTx AccountTx
		b := iter.Value()
		err = json.Unmarshal(b, &accountTx)
		if err != nil {
			return txs, err
		}
		txRecord, err := tx.GetTxByHash(db, accountTx.TxHash)
		if err != nil {
			return txs, err
		}

		var txResponse tx.ResponseDeliverTx
		err = tmjson.Unmarshal(txRecord.TxResponse, &txResponse)
		if err != nil {
			return txs, err
		}
		txRes := tx.TxByHeightRecord{
			Code:      txResponse.Code,
			Codespace: txResponse.Codespace,
			GasUsed:   txResponse.GasUsed,
			GasWanted: txResponse.GasWanted,
			Height:    int64(accountTx.BlockHeight),
			RawLog:    txResponse.Log,
			Logs: func() json.RawMessage {
				if txResponse.Code == 0 {
					return []byte(txResponse.Log)
				} else {
					out, _ := json.Marshal([]string{})
					return out
				}
			}(),
			TxHash:    accountTx.TxHash,
			Timestamp: accountTx.Timestamp,
			Tx:        txRecord.Tx,
		}

		txs = append(txs, txRes)
		currentLimit += 1
		if currentLimit >= limit {
			break
		}
		iter.Next()
	}
	return txs, nil
}
