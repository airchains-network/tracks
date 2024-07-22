package junction

import (
	"fmt"
	logs "github.com/airchains-network/tracks/log"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
)

func GetAddress() (string, error) {
	_, _, accountPath, accountName, addressPrefix, _, err := GetJunctionDetails()
	if err != nil {
		logs.Log.Error("can not get junctionDetails.json data: " + err.Error())
		return "", err
	}

	registry, err := cosmosaccount.New(cosmosaccount.WithHome(accountPath))
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error creating account registry: %v", err))
		return "", err
	}

	newTempAccount, err := registry.GetByName(accountName)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error getting account: %v", err))
		return "", err
	}

	newTempAddr, err := newTempAccount.Address(addressPrefix)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error getting address: %v", err))
		return "", err
	}

	return newTempAddr, nil
}
