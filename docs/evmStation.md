
### Remove old data if present
```shell
sudo rm -rf ~/.tracks
```

### Init sequencer
```shell
go run cmd/main.go init --daRpc "mock-rpc" --daKey "mockKey" --daType "mock"  --moniker "monkey" --stationRpc "http://127.0.0.1:26657/" --stationAPI "http://127.0.0.1:26657/" --stationType "evm"
```

### Create Keys for Junction
```shell
go run cmd/main.go keys junction --accountName dummy --accountPath ./.tracks/junction-accounts/keys
```

### Fund the wallet
- join [Airchains Discord](https://discord.com/invite/airchains)
- goto #switchyard-faucet-bot channel
- type `$faucet <your address>`
- you will get few AMF soon...

### Init Prover`
```shell
go run cmd/main.go prover v1EVM
```

### Create station on junction
NOTE: replace value of MyNewAddress with your newly created wallet address
```sh
MyNewAddress="air19afxdqx8nuc5lzfet5uel8gzdwejam8fw2l74f" # replace with your new address
go run cmd/main.go create-station --accountName dummy --accountPath ./.tracks/junction-accounts/keys --jsonRPC "https://junction-testnet-rpc.synergynodes.com/" --info "EVM Track" --tracks $MyNewAddress  --bootstrapNode "/ip4/192.168.1.24/tcp/2300/p2p/12D3KooWFoN66sCWotff1biUcnBE2vRTmYJRHJqZy27x1EpBB6AM"
```

### start  node
```shell
go run cmd/main.go start
```