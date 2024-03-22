package utilis

import (
	"fmt"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
)

func CreateAccount(accountName string, accountPath string) {
	registry, err := cosmosaccount.New(cosmosaccount.WithHome(accountPath))
	if err != nil {
		fmt.Println(err)
		return
	}

	account, mnemonic, err := registry.Create(accountName)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("new account created: ", account, mnemonic)

}
