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
	if err != nil {
		errorMsg := fmt.Errorf("Junction client connection failed, make sure jsonRPC is correct and active. JsonRPC: " + jsonRPC)
		return false, 0, errorMsg
	}
	pageRequest := &query.PageRequest{} // Add this line to create a new PageRequest

	balances, err := client.BankBalances(ctx, accountAddress, pageRequest) // Add pageRequest as the third argument
	if err != nil {
		errorMsg := fmt.Errorf("error querying bank balances for RPC=%s Address=%s", jsonRPC, accountAddress)
		return false, 0, errorMsg
	}

	if len(balances) == 0 {
		errorMsg := fmt.Errorf("account do not have balance. Address: %s", accountAddress)
		return false, 0, errorMsg
	}

	for _, balance := range balances {
		if balance.Denom == "amf" {
			return true, balance.Amount.Int64(), nil
		}
	}

	return false, 0, nil
}
