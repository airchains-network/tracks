package junction

import (
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
)

func CheckIfAccountExists(accountName, accountPath, addressPrefix string) (addr string, err error) {

	registry, err := cosmosaccount.New(cosmosaccount.WithHome(accountPath))
	if err != nil {

		return "", err
	}

	account, err := registry.GetByName(accountName)
	if err != nil {
		return "", err
	}

	addr, err = account.Address(addressPrefix)
	if err != nil {
		return "", err
	}

	return addr, nil
}
