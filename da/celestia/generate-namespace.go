package celestia

import (
	"encoding/base64"
	"math/rand"
)

func GenerateNamespace() string {
	// Create the byte slice with initial leading zeros
	namespaceBytes := make([]byte, totalNamespaceBytes)
	namespaceBytes[leadingZeroBytes] = namespaceVersion // Set version byte

	// Generate random bytes for the user-specified part
	rand.Read(namespaceBytes[leadingZeroBytes+1:])

	encodedNamespace := base64.StdEncoding.EncodeToString(namespaceBytes)

	// Encode the full namespaceBytes to hex string
	return encodedNamespace
}
