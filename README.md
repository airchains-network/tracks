
### Clear old data if present 
```shell 
sh clear-data.sh;
sudo rm -rf ~/.tracks
```

### Init sequencer
```shell
go run cmd/main.go init --daRpc "mock-rpc" --daKey "mockKey" --daType "mock"  --moniker "monkey" --stationRpc "http://192.168.1.24:8545" --stationAPI "http://192.168.1.24:8545" --stationType "evm" 
```

### create station on junction
```sh
go run cmd/main.go create-station --accountName noob --accountPath ./accounts/keys --jsonRPC "http://34.131.189.98:26657" --info "some info" --tracks air1dqf8xx42e8tlcwpd4ucwf60qeg4k6h7mzpnkf7,air1h25pqnxkv8g50n5nlrdv94wktjupfu4ujevsc8
```

### start single node 
note: it will stick on submit pod, if u put more then one track in create-track in above code..
```shell
go run cmd/main.go start
```

### Start multi node
```shell
go run cmd/main.go start

go run cmd/main.go start /ip4/192.168.1.24/tcp/2300/p2p/12D3KooWPi96exciLFcjnfN73dFD9GfynKYva3iibCYDTTaStKdM
go run cmd/main.go start /ip4/192.168.1.25/tcp/2300/p2p/12D3KooWB1CgEXF97AMga3xDdSggpfPbm7Npx2LPdWdLJt678tLY
```
