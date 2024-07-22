
### Remove old data if present
```shell 
sudo rm -rf ~/.tracks
```

### Init sequencer
```shell
go run cmd/main.go init --daRpc "mock-rpc" --daKey "mockKey" --daType "mock"  --moniker "monkey" --stationRpc "http://127.0.0.1:26657" --stationAPI "http://127.0.0.1:1317" --stationType "wasm" 
```

### Create Keys for Junction
```shell
go run cmd/main.go keys junction --accountName dummy --accountPath ./.tracks/junction-accounts/keys
```

### Init Prover
```shell
go run cmd/main.go prover v1WASM
```

### Create station on junction
```sh
go run cmd/main.go create-station --accountName dummy --accountPath ./.tracks/junction-accounts/keys --jsonRPC "https://junction-testnet-rpc.synergynodes.com/" --info "Wasm Track" --tracks air1pkd0pg82d545xpnfyryfdya9xvhulenzwzvlsn  --bootstrapNode "/ip4/192.168.1.24/tcp/2300/p2p/12D3KooWFoN66sCWotff1biUcnBE2vRTmYJRHJqZy27x1EpBB6AM"
```

### start  node
```shell
go run cmd/main.go start
```