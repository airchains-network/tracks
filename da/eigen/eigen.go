//package eigen
//
//func Eigen(daData []byte, rpcUrl string, accountKey string) (string, error) {
//	//encodedString := base64.StdEncoding.EncodeToString(daData)
//	//daDataString := string(encodedString)
//	//input := C.CString(daDataString)
//	//defer C.free(unsafe.Pointer(input))
//	//
//	//// Convert Go string length to C size_t
//	//inputLen := C.size_t(len(daData))
//	//
//	//// Convert endpoint and accountID strings to C strings
//	//endpointCStr := C.CString(rpcUrl)
//	//defer C.free(unsafe.Pointer(endpointCStr))
//	//accountIDCStr := C.CString(accountKey)
//	//defer C.free(unsafe.Pointer(accountIDCStr))
//	//
//	//// Call the Rust function
//	//resultCStr := C.eigen_da_sync((*C.uchar)(unsafe.Pointer(input)), inputLen, endpointCStr, accountIDCStr)
//	//defer C.eigen_da_free_return_string(resultCStr)
//	//
//	//// Convert C string back to Go string
//	//resultGoStr := C.GoString(resultCStr)
//	//
//	//// Check if the result string starts with "Error:".
//	//if resultGoStr != "" && resultGoStr[:6] == "Error:" {
//	//	// Extract the error message and return it as an error.
//	//	errMsg := resultGoStr[7:]
//	//	return "", fmt.Errorf("%s", errMsg)
//	//}
//	//
//	//// Print the result
//	//return resultGoStr, nil
//	return "", nil
//}

package eigen

import (
	"context"
	"crypto/tls"
	"fmt"
	disperserGrpc "github.com/airchains-network/decentralized-sequencer/da/eigen/grpc"
	"github.com/airchains-network/decentralized-sequencer/da/eigen/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"time"
)

func Eigen(daData []byte, rpcUrl string, accountKey string) (string, error) {
	ctx := context.Background()

	credential := credentials.NewTLS(&tls.Config{})
	addr := fmt.Sprintf("%v:%v", rpcUrl, 443)
	dialOptions := grpc.WithTransportCredentials(credential)
	conn, err := grpc.Dial(addr, dialOptions)
	if err != nil {
		fmt.Println(err)
		return "nil", err
	}
	defer func() { _ = conn.Close() }()

	d, de := DisperserBlob(ctx, conn, daData, accountKey)
	if de != nil {
		fmt.Printf("failed to disperse blob: %v", de)
		return "nil", err
	}

	blobKey := string(d[:])

	//TODO Add DA check Status
	fmt.Println("Eigen DA Blob KEY", blobKey)
	return blobKey, nil

}

func DisperserBlob(ctx context.Context, conn *grpc.ClientConn, daData []byte, accountKey string) ([]byte, error) {
	disperserClient := disperserGrpc.NewDisperserClient(conn)
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	quorums := []uint8{}

	quorumNumbers := make([]uint32, len(quorums))
	for i, q := range quorums {
		quorumNumbers[i] = uint32(q)
	}

	data := utils.ConvertByPaddingEmptyByte(daData)
	request := &disperserGrpc.DisperseBlobRequest{
		Data:                data,
		CustomQuorumNumbers: quorumNumbers,
		AccountId:           accountKey,
	}

	reply, err := disperserClient.DisperseBlob(ctxTimeout, request)
	if err != nil {
		return nil, err
	}

	return reply.GetRequestId(), nil
}

func GetStatus(conn *grpc.ClientConn, requestID []byte, ctx context.Context) (*disperserGrpc.BlobStatus, error) {
	startTime := time.Now()
	timeoutDuration := 5 * time.Minute

	for time.Since(startTime) < timeoutDuration {
		disperserClient := disperserGrpc.NewDisperserClient(conn)
		ctxTimeout, cancel := context.WithTimeout(ctx, timeoutDuration)
		defer cancel()

		request := &disperserGrpc.BlobStatusRequest{
			RequestId: requestID,
		}

		reply, err := disperserClient.GetBlobStatus(ctxTimeout, request)
		if err != nil {
			return nil, err
		}

		check := disperserGrpc.BlobStatus_CONFIRMED
		if reply.Status.Enum() == &check {
			return reply.Status.Enum(), nil
		}

		time.Sleep(2 * time.Second)
		continue
	}

	return nil, fmt.Errorf("timeout")
}
