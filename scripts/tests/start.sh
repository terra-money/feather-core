#!/bin/bash

BINARY=feather-cored
CHAIN_DIR=$(pwd)/.test-data
CHAINID_1=test-1
CHAINID_2=test-2

VAL_MNEMONIC_1="clock post desk civil pottery foster expand merit dash seminar song memory figure uniform spice circle try happy obvious trash crime hybrid hood cushion"
VAL_MNEMONIC_2="angry twist harsh drastic left brass behave host shove marriage fall update business leg direct reward object ugly security warm tuna model broccoli choice"
WALLET_MNEMONIC_1="banner spread envelope side kite person disagree path silver will brother under couch edit food venture squirrel civil budget number acquire point work mass"
WALLET_MNEMONIC_2="veteran try aware erosion drink dance decade comic dawn museum release episode original list ability owner size tuition surface ceiling depth seminar capable only"
RLY_MNEMONIC_1="alley afraid soup fall idea toss can goose become valve initial strong forward bright dish figure check leopard decide warfare hub unusual join cart"
RLY_MNEMONIC_2="record gift you once hip style during joke field prize dust unique length more pencil transfer quit train device arrive energy sort steak upset"

P2PPORT_1=16656
P2PPORT_2=26656
RPCPORT_1=16657
RPCPORT_2=26657
RESTPORT_1=1316
RESTPORT_2=1317
ROSETTA_1=8080
ROSETTA_2=8081
GRPCPORT_1=8090
GRPCPORT_2=9090
GRPCWEB_1=8091
GRPCWEB_2=9091

# Stop if it is already running 
if pgrep -x "$BINARY" >/dev/null; then
    echo "Terminating $BINARY..."
    killall $BINARY
fi

# Ensure go-rly is installed
if ! [ -x "$(command -v rly)" ]; then
    echo "rly is required to run this script..."
    echo "You can download at https://github.com/cosmos/relayer/releases/tag/v2.3.1"
    exit 1
fi

# Ensure screen manager with VT100/ANSI terminal emulation is installed
if ! [ -x "$(command -v screen)" ]; then
    echo "screen is required to run this script..."
    echo "You can download at https://git.savannah.gnu.org/cgit/screen.git"
    exit 1
fi
echo "Removing previous data..."
rm -rf $CHAIN_DIR/$CHAINID_1 &> /dev/null
rm -rf $CHAIN_DIR/$CHAINID_2 &> /dev/null

# Add directories for both chains, exit if an error occurs
if ! mkdir -p $CHAIN_DIR/$CHAINID_1 2>/dev/null; then
    echo "Failed to create chain folder. Aborting..."
    exit 1
fi

if ! mkdir -p $CHAIN_DIR/$CHAINID_2 2>/dev/null; then
    echo "Failed to create chain folder. Aborting..."
    exit 1
fi

echo "Initializing $CHAINID_1 & $CHAINID_2..."
$BINARY init test --home $CHAIN_DIR/$CHAINID_1 --chain-id=$CHAINID_1 &> /dev/null
$BINARY init test --home $CHAIN_DIR/$CHAINID_2 --chain-id=$CHAINID_2 &> /dev/null

echo "Adding genesis accounts..."
echo $VAL_MNEMONIC_1 | $BINARY keys add val1 --home $CHAIN_DIR/$CHAINID_1 --recover --keyring-backend=test
echo $VAL_MNEMONIC_2 | $BINARY keys add val2 --home $CHAIN_DIR/$CHAINID_2 --recover --keyring-backend=test
echo $WALLET_MNEMONIC_1 | $BINARY keys add wallet1 --home $CHAIN_DIR/$CHAINID_1 --recover --keyring-backend=test
echo $WALLET_MNEMONIC_2 | $BINARY keys add wallet2 --home $CHAIN_DIR/$CHAINID_2 --recover --keyring-backend=test
echo $RLY_MNEMONIC_1 | $BINARY keys add rly1 --home $CHAIN_DIR/$CHAINID_1 --recover --keyring-backend=test 
echo $RLY_MNEMONIC_2 | $BINARY keys add rly2 --home $CHAIN_DIR/$CHAINID_2 --recover --keyring-backend=test 

VAL1_ADDR=$($BINARY keys show val1 --home $CHAIN_DIR/$CHAINID_1 --keyring-backend test -a)
VAL2_ADDR=$($BINARY keys show val2 --home $CHAIN_DIR/$CHAINID_2 --keyring-backend test -a)
WALLET1_ADDR=$($BINARY keys show wallet1 --home $CHAIN_DIR/$CHAINID_1 --keyring-backend test -a)
WALLET2_ADDR=$($BINARY keys show wallet2 --home $CHAIN_DIR/$CHAINID_2 --keyring-backend test -a)
RLY1_ADDR=$($BINARY keys show rly1 --home $CHAIN_DIR/$CHAINID_1 --keyring-backend test -a)
RLY2_ADDR=$($BINARY keys show rly2 --home $CHAIN_DIR/$CHAINID_2 --keyring-backend test -a)

$BINARY genesis add-genesis-account $VAL1_ADDR 1000000000000stake --home $CHAIN_DIR/$CHAINID_1
$BINARY genesis add-genesis-account $VAL2_ADDR 1000000000000stake --home $CHAIN_DIR/$CHAINID_2
$BINARY genesis add-genesis-account $WALLET1_ADDR 1000000000000stake --home $CHAIN_DIR/$CHAINID_1
$BINARY genesis add-genesis-account $WALLET2_ADDR 1000000000000stake --home $CHAIN_DIR/$CHAINID_2
$BINARY genesis add-genesis-account $RLY1_ADDR 1000000000000stake --home $CHAIN_DIR/$CHAINID_1
$BINARY genesis add-genesis-account $RLY2_ADDR 1000000000000stake --home $CHAIN_DIR/$CHAINID_2

echo "Creating and collecting gentx..."
$BINARY genesis gentx val1 6500000000stake --home $CHAIN_DIR/$CHAINID_1 --chain-id $CHAINID_1 --keyring-backend test
$BINARY genesis gentx val2 6500000000stake --home $CHAIN_DIR/$CHAINID_2 --chain-id $CHAINID_2 --keyring-backend test
$BINARY genesis collect-gentxs --home $CHAIN_DIR/$CHAINID_1 &> /dev/null
$BINARY genesis collect-gentxs --home $CHAIN_DIR/$CHAINID_2 &> /dev/null

echo "Changing defaults and ports in app.toml and config.toml files..."
sed -i -e 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:'"$P2PPORT_1"'"#g' $CHAIN_DIR/$CHAINID_1/config/config.toml
sed -i -e 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:'"$RPCPORT_1"'"#g' $CHAIN_DIR/$CHAINID_1/config/config.toml
sed -i -e 's/timeout_commit = "5s"/timeout_commit = "1s"/g' $CHAIN_DIR/$CHAINID_1/config/config.toml
sed -i -e 's/timeout_propose = "3s"/timeout_propose = "1s"/g' $CHAIN_DIR/$CHAINID_1/config/config.toml
sed -i -e 's/enable = false/enable = true/g' $CHAIN_DIR/$CHAINID_1/config/app.toml
sed -i -e 's/swagger = false/swagger = true/g' $CHAIN_DIR/$CHAINID_1/config/app.toml
sed -i -e 's#"tcp://localhost:1317"#"tcp://0.0.0.0:'"$RESTPORT_1"'"#g' $CHAIN_DIR/$CHAINID_1/config/app.toml
sed -i -e 's#":8080"#":'"$ROSETTA_1"'"#g' $CHAIN_DIR/$CHAINID_1/config/app.toml

sed -i -e 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:'"$P2PPORT_2"'"#g' $CHAIN_DIR/$CHAINID_2/config/config.toml
sed -i -e 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:'"$RPCPORT_2"'"#g' $CHAIN_DIR/$CHAINID_2/config/config.toml
sed -i -e 's/timeout_commit = "5s"/timeout_commit = "1s"/g' $CHAIN_DIR/$CHAINID_2/config/config.toml
sed -i -e 's/timeout_propose = "3s"/timeout_propose = "1s"/g' $CHAIN_DIR/$CHAINID_2/config/config.toml
sed -i -e 's/enable = false/enable = true/g' $CHAIN_DIR/$CHAINID_2/config/app.toml
sed -i -e 's/swagger = false/swagger = true/g' $CHAIN_DIR/$CHAINID_2/config/app.toml
sed -i -e 's#"tcp://localhost:1317"#"tcp://0.0.0.0:'"$RESTPORT_2"'"#g' $CHAIN_DIR/$CHAINID_2/config/app.toml
sed -i -e 's#":8080"#":'"$ROSETTA_2"'"#g' $CHAIN_DIR/$CHAINID_2/config/app.toml

echo "Chaning genesis.json..."
sed -i -e 's/"voting_period": "172800s"/"voting_period": "10s"/g' $CHAIN_DIR/$CHAINID_1/config/genesis.json
sed -i -e 's/"voting_period": "172800s"/"voting_period": "10s"/g' $CHAIN_DIR/$CHAINID_2/config/genesis.json
sed -i -e 's/"reward_delay_time": "604800s"/"reward_delay_time": "0s"/g' $CHAIN_DIR/$CHAINID_1/config/genesis.json
sed -i -e 's/"reward_delay_time": "604800s"/"reward_delay_time": "0s"/g' $CHAIN_DIR/$CHAINID_2/config/genesis.json
sed -i -e 's/"base_denom": "uluna"/"base_denom": "stake"/g' $CHAIN_DIR/$CHAINID_1/config/genesis.json
sed -i -e 's/"base_denom": "uluna"/"base_denom": "stake"/g' $CHAIN_DIR/$CHAINID_2/config/genesis.json
sed -i -e 's/"base_chain_id": "phoenix-1"/"base_chain_id": "test-1"/g' $CHAIN_DIR/$CHAINID_2/config/genesis.json
sed -i -e 's/"alliance_bond_height": "1000"/"alliance_bond_height": "65"/g' $CHAIN_DIR/$CHAINID_1/config/genesis.json
sed -i -e 's/"alliance_bond_height": "1000"/"alliance_bond_height": "65"/g' $CHAIN_DIR/$CHAINID_2/config/genesis.json
sed -i -e 's/"halt_if_no_channel": false/"halt_if_no_channel": true/g' $CHAIN_DIR/$CHAINID_1/config/genesis.json
sed -i -e 's/"halt_if_no_channel": false/"halt_if_no_channel": true/g' $CHAIN_DIR/$CHAINID_2/config/genesis.json

echo "Starting $CHAINID_1 in $CHAIN_DIR..."
echo "Creating log file at $CHAIN_DIR/$CHAINID_1.log"
$BINARY start --log_level trace --log_format json --home $CHAIN_DIR/$CHAINID_1 --pruning=nothing --grpc.address="0.0.0.0:$GRPCPORT_1" --grpc-web.address="0.0.0.0:$GRPCWEB_1" > $CHAIN_DIR/$CHAINID_1.log 2>&1 &

echo "Starting $CHAINID_2 in $CHAIN_DIR..."
echo "Creating log file at $CHAIN_DIR/$CHAINID_2.log"
$BINARY start --log_level trace --log_format json --home $CHAIN_DIR/$CHAINID_2 --pruning=nothing --grpc.address="0.0.0.0:$GRPCPORT_2" --grpc-web.address="0.0.0.0:$GRPCWEB_2" > $CHAIN_DIR/$CHAINID_2.log 2>&1 &
