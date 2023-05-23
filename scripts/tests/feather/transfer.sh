#!/bin/bash

echo ""
echo "############################################################"
echo "# IBC: Transfer a token from test-1 to test-2 throught IBC #"
echo "############################################################"
echo ""

BINARY=feather-cored
CHAIN_DIR=$(pwd)/.test-data

VAL_WALLET_1=$($BINARY keys show val1 -a --keyring-backend test --home $CHAIN_DIR/test-1)
VAL_WALLET_2=$($BINARY keys show val2 -a --keyring-backend test --home $CHAIN_DIR/test-2)

echo "Sending tokens from chain test-1 to chain test-2"
IBC_TRANSFER=$($BINARY tx ibc-transfer transfer transfer channel-0 $VAL_WALLET_2 1stake --chain-id test-1 --from $VAL_WALLET_1 --home $CHAIN_DIR/test-1 --node tcp://localhost:16657 --keyring-backend test --trace -y -o json | jq -r '.raw_log' )

if [[ "$IBC_TRANSFER" == "failed to execute message"* ]]; then
    echo "Error: IBC transfer failed, with error: $IBC_TRANSFER"
    exit 1
fi

ACCOUNT_BALANCE=""
IBC_DENOM=""
while [ "$ACCOUNT_BALANCE" == "" ]; do
    IBC_DENOM=$($BINARY q bank balances $VAL_WALLET_2 --chain-id test-2 --node tcp://localhost:26657 -o json | jq -r '.balances[0].denom')
    if [ "$IBC_DENOM" != "stake" ]; then
        ACCOUNT_BALANCE=$($BINARY q bank balances $VAL_WALLET_2 --chain-id test-2 --node tcp://localhost:26657 -o json | jq -r '.balances[0].amount')
    fi
    sleep 2
done

echo ""
echo "################################################################"
echo "# SUCCESS: Transfer a token from test-1 to test-2 throught IBC #"
echo "################################################################"
echo ""