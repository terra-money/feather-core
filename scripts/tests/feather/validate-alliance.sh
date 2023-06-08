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

    if [ $CHAIN_1_HEIGHT -eq 65 ] && [ $CHAIN_2_HEIGHT -eq 65  ]; then
        ALLIANCES_ON_CHAIN_2=$($BINARY q alliance alliances --home $CHAIN_DIR/test-2 --node tcp://localhost:26657 -o json | jq -r ".alliances | length")

        if [ $ALLIANCES_ON_CHAIN_2 -ne 1 ]; then
            echo "Error: Alliance not created on chain-1"
            exit 1
        fi

        if [ $CHAIN_1_HEIGHT -ne 65 ]; then
            echo "Error: Alliance not created on chain-2"
            exit 1
        fi
        
        echo "Chain test-1 reached block $CHAIN_1_HEIGHT and halted successfully"
        echo "Chain test-2 reached block $CHAIN_2_HEIGHT and created $ALLIANCES_ON_CHAIN_2 alliance successfully"
        break
    fi

    echo "Waiting for chains test-1 to reach block 65 (current height $CHAIN_1_HEIGHT) and test-2 to get over block 65 (current height $CHAIN_2_HEIGHT)"
    if [ $CHAIN_1_HEIGHT -gt 70 ]; then
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