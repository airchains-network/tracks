package eigen

/*
#cgo LDFLAGS: -L../../lib -ldemo_rust
#include <stdlib.h>
extern char* eigen_da_sync(const unsigned char* da_data, size_t da_data_len, const char* endpoint, const char* account_id);
extern void eigen_da_free_return_string(char* s);
*/
import "C"
import (
	"encoding/base64"
	"fmt"
	"unsafe"
)

func Eigen(daData []byte, rpcUrl string, accountKey string) (string, error) {
	encodedString := base64.StdEncoding.EncodeToString(daData)
	daDataString := string(encodedString)
	input := C.CString(daDataString)
	defer C.free(unsafe.Pointer(input))

	// Convert Go string length to C size_t
	inputLen := C.size_t(len(daData))

	// Convert endpoint and accountID strings to C strings
	endpointCStr := C.CString(rpcUrl)
	defer C.free(unsafe.Pointer(endpointCStr))
	accountIDCStr := C.CString(accountKey)
	defer C.free(unsafe.Pointer(accountIDCStr))

	// Call the Rust function
	resultCStr := C.eigen_da_sync((*C.uchar)(unsafe.Pointer(input)), inputLen, endpointCStr, accountIDCStr)
	defer C.eigen_da_free_return_string(resultCStr)

	// Convert C string back to Go string
	resultGoStr := C.GoString(resultCStr)

	// Check if the result string starts with "Error:".
	if resultGoStr != "" && resultGoStr[:6] == "Error:" {
		// Extract the error message and return it as an error.
		errMsg := resultGoStr[7:]
		return "", fmt.Errorf("%s", errMsg)
	}

	// Print the result
	return resultGoStr, nil
}
