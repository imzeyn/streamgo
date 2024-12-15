package streamgo

import "bytes"

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