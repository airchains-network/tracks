
### Remove old data if present 
```shell 
sudo rm -rf ~/.tracks
```

### Init sequencer
```shell
./bin/tracks init --daRpc "mock-rpc" --daKey "mockKey" --daType "mock"  --moniker "monkey" --stationRpc "http://127.0.0.1:8545" --stationAPI "http://127.0.0.1:8545" --stationType "evm" 
```
### Create Keys for Junction
```shell
./bin/tracks keys junction --accountName dummy --accountPath ./accounts/keys
```

### Init Prover
```shell
./bin/tracks  prover v1EVM
```

### Create station on junction
```sh
./bin/tracks create-station --accountName dummy --accountPath ./accounts/keys --jsonRPC "http://localhost:1213" --info "basic info" --tracks air1dqf8xx42e8tlcwpd4ucwf60qeg4k6h7mzpnkf7  --bootstrapNode "/ip4/192.168.1.24/tcp/2300/p2p/12D3KooWFoN66sCWotff1biUcnBE2vRTmYJRHJqZy27x1EpBB6AM"
```

### start  node 
```shell
./bin/tracks start
```



