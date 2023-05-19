echo ""
echo "##########################################################"
echo "# Feather Module: validate alliance created successfully #"
echo "##########################################################"
echo ""

BINARY=feather-cored
CHAIN_DIR=$(pwd)/.test-data

CHAIN_1_HEIGHT=""
CHAIN_2_HEIGHT=""
while true; do
    CHAIN_1_HEIGHT=$($BINARY status --home $CHAIN_DIR/test-1 --node tcp://localhost:16657 | jq -r ".SyncInfo.latest_block_height")
    CHAIN_2_HEIGHT=$($BINARY status --home $CHAIN_DIR/test-2 --node tcp://localhost:26657 | jq -r ".SyncInfo.latest_block_height")

    if [ $CHAIN_1_HEIGHT == 25 ] && [ $CHAIN_2_HEIGHT -gt 25  ]; then
        sleep 5
        ALLIANCES_ON_CHAIN_2=$($BINARY q alliance alliances --home $CHAIN_DIR/test-2 --node tcp://localhost:26657 -o json | jq -r ".alliances | length")

        if [ $ALLIANCES_ON_CHAIN_2 != 1 ]; then
            echo "Error: Alliance not created on chain-1"
            exit 1
        fi

        if [ $CHAIN_1_HEIGHT != 25 ]; then
            echo "Error: Alliance not created on chain-2"
            exit 1
        fi
        
        echo "Chain test-1 reached block $CHAIN_1_HEIGHT and halted successfully"
        echo "Chain test-2 reached block $CHAIN_2_HEIGHT and created $ALLIANCES_ON_CHAIN_2 alliance successfully"
        break
    fi

    echo "Waiting for chains test-1 to reach block 25 (current height $CHAIN_1_HEIGHT) and test-2 to get over block 25 (current height $CHAIN_2_HEIGHT)"
        if [ $CHAIN_1_HEIGHT == 30 ]; then
        echo "Chan test-1 didn't halted!"
        exit 1
    fi
    sleep 1
done

echo ""
echo "#################################################"
echo "# Feather Module: alliance created successfully #"
echo "#################################################"
echo ""