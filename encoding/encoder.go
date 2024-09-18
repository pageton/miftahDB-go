package encoding

import (
	"errors"

	"github.com/vmihailenco/msgpack/v5"
)

// concatUint8Arrays merges multiple byte arrays into a single byte array
func concatUint8Arrays(arrays ...[]byte) []byte {
	totalLength := 0
	for _, arr := range arrays {
		totalLength += len(arr)
	}
	result := make([]byte, totalLength)
	offset := 0
	for _, arr := range arrays {
		copy(result[offset:], arr)
		offset += len(arr)
	}
	return result
}

// EncodeValue encodes the given value into a msgpack format with a marker
func EncodeValue(value interface{}) ([]byte, error) {
	var marker []byte
	var encodedValue []byte
	var err error

	if byteArray, ok := value.([]byte); ok {
		marker = []byte{0x01}
		encodedValue = byteArray
	} else {
		marker = []byte{0x02}
		encodedValue, err = msgpack.Marshal(value)
		if err != nil {
			return nil, err
		}
	}
	return concatUint8Arrays(marker, encodedValue), nil
}

// DecodeValue decodes the msgpack-encoded byte array back into the original value
func DecodeValue[T any](buffer []byte) (T, error) {
	var result T
	marker := buffer[0]
	actualValue := buffer[1:]

	switch marker {
	case 0x01:
		// For byte array data
		if v, ok := any(actualValue).(T); ok {
			return v, nil
		} else {
			return result, errors.New("type assertion failed for []byte")
		}
	case 0x02:
		// For msgpack encoded data
		err := msgpack.Unmarshal(actualValue, &result)
		if err != nil {
			return result, err
		}
	default:
		return result, errors.New("unknown data marker")
	}
	return result, nil
}
