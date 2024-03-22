package config

import (
	"context"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
)

func JunctionConnect(accountPath, jsonRPC, addressPrefix string) (client cosmosclient.Client, ctx context.Context) {

	ctx = context.Background()
	gasLimit := "70000000"

	client, err := cosmosclient.New(ctx, cosmosclient.WithGas(gasLimit), cosmosclient.WithAddressPrefix(addressPrefix), cosmosclient.WithNodeAddress(jsonRPC), cosmosclient.WithKeyringDir(accountPath))
	if err != nil {
		panic(err)
	}

	return client, ctx
}
