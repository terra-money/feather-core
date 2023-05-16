make install
feather-cored init test --overwrite --chain-id test
yes | feather-cored keys add test --keyring-backend os
yes | feather-cored genesis add-genesis-account test 100000000stake --keyring-backend os
export CHAIN_ID=$(cat ~/.feather-core/config/genesis.json | jq -r '.chain_id')
yes | feather-cored genesis gentx test 1000000stake --chain-id $CHAIN_ID --keyring-backend os
feather-cored genesis collect-gentxs
feather-cored start