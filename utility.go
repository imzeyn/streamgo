package streamgo

import (
	"bytes"
)

func ClearURL(s string) string {

	var builder bytes.Buffer
	builder.Grow(len(s))
	builder.WriteByte('/')

	bytesSlice := []byte(s)
	start := 0

	for i := 0; i < len(bytesSlice); i++ {
		b := bytesSlice[i]

		if b == '/' || i == len(bytesSlice)-1 {

			if i == len(bytesSlice)-1 && b != '/' {
				i++
			}

			part := bytesSlice[start:i]
			if len(bytes.TrimSpace(part)) > 0 {
				builder.Write(part)
				builder.WriteByte('/')
			}

			start = i + 1
		}
	}

	return builder.String()
}

func SplitArray[T any](data []T, numParts int) [][]T {
	if numParts <= 0 || len(data) == 0 {
		return nil
	}

	if numParts == 1 {
		return [][]T{data}
	}

	length := len(data)
	chunkSize, remainder := length/numParts, length%numParts

	result := make([][]T, numParts)
	start, idx := 0, 0

	for ; idx < remainder; idx++ {
		end := start + chunkSize + 1
		result[idx] = data[start:end]
		start = end
	}

	for ; idx < numParts; idx++ {
		end := start + chunkSize
		result[idx] = data[start:end]
		start = end
	}

	for i := 0; i < numParts; i++ {
		if len(result[i]) == 0 {
			result = result[:i]
			break
		}
	}

	return result
}
