package utils

const BYTES_PER_SYMBOL = 32

func ConvertByPaddingEmptyByte(data []byte) []byte {
	dataSize := len(data)
	parseSize := BYTES_PER_SYMBOL - 1
	putSize := BYTES_PER_SYMBOL

	dataLen := (dataSize + parseSize - 1) / parseSize

	validData := make([]byte, dataLen*putSize)
	validEnd := len(validData)

	for i := 0; i < dataLen; i++ {
		start := i * parseSize
		end := (i + 1) * parseSize
		if end > len(data) {
			end = len(data)
			// 1 is the empty byte
			validEnd = end - start + 1 + i*putSize
		}

		// with big endian, set first byte is always 0 to ensure data is within valid range of
		validData[i*BYTES_PER_SYMBOL] = 0x00
		copy(validData[i*BYTES_PER_SYMBOL+1:(i+1)*BYTES_PER_SYMBOL], data[start:end])

	}
	return validData[:validEnd]
}
