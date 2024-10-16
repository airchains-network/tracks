
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
stationAPI="http://127.0.0.1:8545"
stationType="evm" 
sequencerType="default"
daName="mock"
./build/tracks init --daRpc "$daRpc" --daKey "$daKey" --daType "$daType" --moniker "$moniker" --stationRpc "$stationRpc" --stationAPI "$stationAPI" --stationType "$stationType"  --sequencerType "$sequencerType" --daName "$daName"
```

Initialise the sequencer for Espresso 
```sh
sequencerType="espresso"
daRpc="mock-rpc"
daKey="mockKey"
daName="mocha"
daType="celestia"
moniker="monkey"
stationRpc="http://127.0.0.1:8545"
stationAPI="http://127.0.0.1:8545"
stationType="evm" 
sequencerRPC="https://query.decaf.testnet.espresso.network"
sequencerNamespace="2345678"

./build/tracks init --daRpc "$daRpc" --daKey "$daKey" --daName "$daName" --daType "$daType" --moniker "$moniker" --stationRpc "$stationRpc" --stationAPI "$stationAPI" --stationType "$stationType" --sequencerType "$sequencerType" --sequencerRpc "$sequencerRPC" --sequencerNamespace "$sequencerNamespace"
```

## Step 4: Initialize the Prover (not required if using external sequencer)

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
accountAddressArray="air16yhjt95p7eqyxm6wl3fmv2pdfv7qfx7m8mdyhv" #! replace it with your address
accountName="dummy"
accountPath=".tracks/junction-accounts/keys"
jsonRPC="http://0.0.0.0:26657" 
stationName="testStation"
bootstrapNode="/ip4/192.168.1.24/tcp/2300/p2p/12D3KooWFoN66sCWotff1biUcnBE2vRTmYJRHJqZy27x1EpBB6AM"
#info="EVM Track"
operators="air1e7l4nlsj8hww60y6kjas9ccq2v9x3ep5spaqlw,air16yhjt95p7eqyxm6wl3fmv2pdfv7qfx7m8mdyhv"

./build/tracks create-station --stationName "$stationName" --tracks "$accountAddressArray" --accountName "$accountName" --accountPath "$accountPath" --jsonRPC "$jsonRPC" --bootstrapNode "$bootstrapNode" --operators "$operators"
```

## Step 8: Start the Tracks

Finally, start the node to begin interacting with the Tracks blockchain.
```shell
./build/tracks start
```

### List Engagements (In case of Espresso)
```shell
./build/tracks query list-station-engagements --offset 0 --limit 2 --order "desc"
```

### List Schemas (In case of Espresso)
```shell
./build/tracks query list-station-schemas --offset 0 --limit 2 --reverse "true"
```

### List Station (In case of Espresso)
```shell
./build/tracks query list-station --offset 0 --limit 35 --reverse "true"
```


## Troubleshooting

If you encounter any issues during setup, refer to [official documentation](https://docs.airchains.io/rollups/evm-zk-rollup/system-requirements) or reach out [Airchains discord](https://discord.gg/airchains) for support.
