
# Tracks Setup Guide
This guide will help you set up and initialize the Tracks environment.

## Step 1: Remove Old Data

If you have previously set up Tracks, remove any old data to ensure a clean environment.

```shell
sudo rm -rf ~/.tracks
```
## Step 2: Build  the Tracks

```bash
make build
```

## Step 3: Initialize the Tracks

Initialize the sequencer with the necessary parameters.

```shell
daRpc="mock-rpc"
daKey="mockKey"
daType="mock"
moniker="monkey"
stationRpc="http://127.0.0.1:8545"
stationAPI="http://127.0.0.1:26657"
stationType="evm" 

./build/tracks init --daRpc "$daRpc" --daKey "$daKey" --daType "$daType" --moniker "$moniker" --stationRpc "$stationRpc" --stationAPI "$stationAPI" --stationType "$stationType"
```

## Step 4: Initialize the Prover

Initialize the prover. Ensure you specify the correct version.

```shell
./build/tracks prover v1EVM
```

## Step 5: Create Keys for Junction (If not already created)

Create keys for the junction account. If the keys are not already created, use the following command:

```shell
accountName="dummy"
accountPath=".tracks/junction-accounts/keys"

./build/tracks keys junction --accountName "$accountName" --accountPath "$accountPath"
```

Alternatively, you can import an account using a mnemonic:

```shell 
accountName="dummy"
accountPath=".tracks/junction-accounts/keys"
mnemonic="huge bounce thing settle diet mobile fruit skill call roast offer soap other upset toward sand dress moral pole smile limb round vacant ecology"

./build/tracks keys import --accountName "$accountName" --accountPath "$accountPath" --mnemonic "$mnemonic"
```

## Step 6: Fund the wallet 
- Join [Airchains Discord ](https://discord.gg/airchains) 
- Goto `switchyard-faucet-bot` channel
- Type `$faucet <your_address>`, you will get 2AMF soon.

## Step 7: Create a Station on Junction

Create a station on the junction with the necessary parameters.
> NOTE: don't forget to replace `accountAddressArray` with the addresses you want to make track member. Replace it with  your new address 

```shell
accountAddressArray="air1knf2an5efl8d9t5w75uds4ty8uj0scxx9yg783" #! replace it with your address
accountName="dummy"
accountPath=".tracks/junction-accounts/keys"
#jsonRPC="http://0.0.0.0:26657" # localhost testing
jsonRPC="https://junction-testnet-rpc.synergynodes.com/" # junction testnet 
bootstrapNode="/ip4/192.168.1.24/tcp/2300/p2p/12D3KooWFoN66sCWotff1biUcnBE2vRTmYJRHJqZy27x1EpBB6AM"
info="EVM Track"

./build/tracks create-station --tracks "$accountAddressArray" --accountName "$accountName" --accountPath "$accountPath" --jsonRPC "$jsonRPC" --info "$info" --bootstrapNode "$bootstrapNode"
```

## Step 8: Start the Tracks

Finally, start the node to begin interacting with the Tracks blockchain.

```shell
./build/tracks start
```

## Troubleshooting

If you encounter any issues during setup, refer to [official documentation](https://docs.airchains.io/rollups/evm-zk-rollup/system-requirements) or reach out [Airchains discord](https://discord.gg/airchains) for support.
