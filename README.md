
### Remove old data if present 
```shell 
sudo rm -rf ~/.tracks
```

### Init sequencer
```shell
go run cmd/main.go init --daRpc "mock-rpc" --daKey "mockKey" --daType "mock"  --moniker "monkey" --stationRpc "http://127.0.0.1:8545" --stationAPI "http://127.0.0.1:8545" --stationType "evm" 
```
### Create Keys for Junction
```shell
go run cmd/main.go keys junction --accountName dummy --accountPath ./accounts/keys
```

### Init Prover
```shell
go run cmd/main.go prover v1EVM
```

### Create station on junction
```sh
go run cmd/main.go create-station --accountName dummy --accountPath ./accounts/keys --jsonRPC "http://localhost:26667" --info "basic info" --tracks air1gzyukqnjzs4j07vmwf9fvfageeer62t0zqvx0x  --bootstrapNode "/ip4/192.168.1.24/tcp/2300/p2p/12D3KooWFoN66sCWotff1biUcnBE2vRTmYJRHJqZy27x1EpBB6AM"
```

### start  node 
```shell
go run cmd/main.go start
```
