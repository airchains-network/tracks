package da

import (
	"github.com/airchains-network/decentralized-sequencer/blocksync"
	"github.com/airchains-network/decentralized-sequencer/da/avail"
	"github.com/airchains-network/decentralized-sequencer/da/celestia"
	"github.com/airchains-network/decentralized-sequencer/da/eigen"
	mock "github.com/airchains-network/decentralized-sequencer/da/mockda"
	"github.com/airchains-network/decentralized-sequencer/types"
	"log"
)

func DALayer(daData []byte, batchNumber int) (string, error) {

	daConfig := types.DAConfigType{
		DALayer:    "mockdb",                                   // avail, celestia, eigen, mockdb
		DARpc:      "https://disperser-goerli.eigenda.xyz:443", // avail http://127.0.0.1:7000/,  celestia http://localhost:26658", eigen https://disperser-goerli.eigenda.xyz:443
		AccountKey: "9430d5ad8ea52329be63afe66a8c8d5e0ba75bf0de0cbd41aa30fadf5f575ec24cff557777e20a0578ec4fedc66274c37fe5d25ed4c4a09cb73b1ddc15349bb4",
		DAAuth:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJwdWJsaWMiLCJyZWFkIiwid3JpdGUiLCJhZG1pbiJdfQ.CiV6oPz7sWeez1LrCSquzxisuagHv2k9nZDii5kQs_8",
	}

	switch daConfig.DALayer {
	case "avail":
		successRes, err := avail.Avail(daData, daConfig.DARpc)
		if err != nil {
			return "", err
		}
		return successRes, nil
	case "celestia":
		successRes, err := celestia.Celestia(daData, daConfig.DARpc, daConfig.DAAuth)
		if err != nil {
			return "", err
		}
		return successRes, nil
	case "eigen":
		successRes, err := eigen.Eigen(daData,
			daConfig.DARpc, daConfig.AccountKey,
		)
		if err != nil {
			return "", err
		}
		return successRes, nil
	case "mockdb":
		mdb := blocksync.GetMockDbInstance()
		successRes, err := mock.MockDA(mdb, daData, batchNumber)
		if err != nil {
			return "", err
		}
		return successRes, nil
	default:
		log.Fatalln("Unknown layer. Please use 'avail' or 'celestia' as argument.")
		return "", nil
	}
}
