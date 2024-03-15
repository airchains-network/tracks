// sample POST request:
// {"jsonrpc":"2.0","method":"tracks_getLatestPod","params":[],"id":1}

// Sample Response:
/*
	{
	"jsonrpc":"2.0",
	"id":"1",
	"error":{"code":200,"message":""},
	"result":[
		{
			"LatestPodHeight":417,
			"LatestPodHash":"IJN....",
			"LatestPodProof":"ey...",
			"LatestPublicWitness":"WyIx...",
			"Votes":{},
			"TracksAppHash":"IJ...",
			"Batch":{"from":["0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35","0x5f98dDB37bE09AC29e9D3649aE6E2a4dca2e2e35"],"to":["0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693","0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693"],"amounts":["1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1","1"],"tx_hashes":["0x25668c6bf0b2dbd57d1e53018816046365587107b0935a3872287644e5451e73","0x57262935a098635a2faad8f8f79b6b0a2d68b0e064a0cbe561b969beaa0f9a14","0xa4016bdb35e7f9dafcba8d4518abccb37ea1ebfd7e15829511043d84c25f42d0","0x971e3d8409903d8b933bb0dce5580c4bd0308f11bea486d0bb061374de77c734","0xb80b0a2ad00f83680623d23b36c727af3a69b34ea9b565977e8d708c28c80cd7","0xa5d075f8dee8ab9ebbea7ffc6f53cebd38375deff482985106e63ec042bf5fe1","0x53a8a3a424cc9413ba0c12f6668f1abb8a86b7216b8abb7f2399600b0798df2e","0x71e20e6facde034ec1b4aa1461a38752de6eb319575f1370b80e502724d88f8e","0x49cc6c5506deb90d61cb181abc05891650b5f789d495bc87b053d0e36726725b","0x7cb088b559e9f42fdfabba15b7e48e54bd518474dfaf7c8ca5eae2be5828b6dc","0x28cb38ed5bd7ecaf34cfdff037c0eb302e1e298e5ed7443d7a609173660bfba4","0xcd6fbfe260d83274ad296ff06129f21c81e8b4900b77218c6813bb39ea960c48","0x1c9501150e0366d673faddd7d17fd25a8f7add86ca3cbbd589c114d80f8c0c30","0xd6933a1e5cdf33817514c98dfbc5da15cebb0f0e426a445bc1acce4d820aafe1","0x42d3f7cc9c78215addd749d1d50a2f8d54079b8f91b0335a2caad7aa9ed1227f","0xdd3cc36a571efe2d70e3e149ba9a7d0b886e6d9779beabbf674bf5dfb9362f2d","0x2da92ba42959075f11942c45c29052fcc583175be426641a1933a0a54ea800e6","0x8c0a782bc22891f0da2543ff3806eb77bd6d3a97874aa9a9e6f6a28cdf3759e4","0x2592563dd3dc0d50cfdfe5f2a54e580e621ef13c806f08d477007eb226b72a5d","0x34594207baa0ec40287857b80ccd6a215fe5a9967c5c08cffcc052d9b243d53d","0x459207b637563429d41a6bc6c0a2fda331cbbb5717dd1ee1f3284dbf8334c9be","0x13cee26aacbc0da9a09c135bbcbd48fb826b4f37c93add70b22f3ee2f6c9a404","0xf6de6522bde6ff9d561d33e51d2f43b9f30521512a78c72a0c8c0d0b1f649e6b","0x91ef43ee6bc58aa1a95c00a134e8307b0de0fbc3af89438124e62643e44e635a","0x583788914dbcb0228231c15aa622abf7d215276447915bf7271b4b09ef118259"],"sender_balances":["99998999999999955312409620","99998999999999955312409620","99998999999999955312409620","99998999999999955312409620","99998999999999955312409620","99998999999999955312409620","99998999999999955312409620","99998999999999955312409620","99998999999999955312409620","99998999999999955312409620","99998999999999955312409620","99998999999999955312409620","99998999999999955288889588","99998999999999955288889588","99998999999999955288889588","99998999999999955288889588","99998999999999955288889588","99998999999999955288889588","99998999999999955288889588","99998999999999955288889588","99998999999999955288889588","99998999999999955288889588","99998999999999955288889588","99998999999999955288889588","99998999999999955288889588"],"receiver_balances":["10380","10380","10380","10380","10380","10380","10380","10380","10380","10380","10380","10380","10412","10412","10412","10412","10412","10412","10412","10412","10412","10412","10412","10412","10412"],"messages":["Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether ","Hello RSM bhaiya i am SlayeR, sending u 1ether "],"tx_nonces":["10401","10402","10403","10404","10405","10406","10407","10408","10409","10410","10411","10412","10413","10414","10415","10416","10417","10418","10419","10420","10421","10422","10423","10424","10425"],
			"account_nonces":["0x0","0x0","0x0","0x0","0x0","0x0","0x0","0x0","0x0","0x0","0x0","0x0","0x0","0x0","0x0","0x0","0x0","0x0","0x0","0x0","0x0","0x0","0x0","0x0","0x0"]},
			"MasterTrackAppHash":null
		}
	]
}
*/

package handler

import (
	"encoding/json"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	"github.com/airchains-network/decentralized-sequencer/types"
	"github.com/gin-gonic/gin"
)

func HandleGetLatestPod(c *gin.Context, Params []any) {

	// currently it have no use
	_ = Params

	var podStateData *types.PodState
	stateConnection := shared.Node.NodeConnections.GetStateDatabaseConnection()

	podStateDataByte, err := stateConnection.Get([]byte("podState"), nil)
	if err != nil {
		logs.Log.Error("error in getting pod state data from database")
		respondWithError(c, err.Error())
		return
	}
	err = json.Unmarshal(podStateDataByte, &podStateData)
	if err != nil {
		logs.Log.Error("error in unmarshal pod state data")
		respondWithError(c, err.Error())
		return
	}

	responseData := []any{podStateData}
	respondWithSuccess(c, responseData, "success")
	return

}
