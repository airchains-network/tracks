package junction

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
)

func CheckBalance(jsonRPC string, accountAddress string) (bool, int64, error) {

	ctx := context.Background()
	client, err := cosmosclient.New(ctx, cosmosclient.WithNodeAddress(jsonRPC))
	pageRequest := &query.PageRequest{} // Add this line to create a new PageRequest

	balances, err := client.BankBalances(ctx, accountAddress, pageRequest) // Add pageRequest as the third argument
	if err != nil {
		fmt.Printf("Error querying bank balances: %v\n", err)
		return false, 0, err
	}

	for _, balance := range balances {
		if balance.Denom == "amf" {
			return true, balance.Amount.Int64(), nil
		}
	}

	return false, 0, nil
}
