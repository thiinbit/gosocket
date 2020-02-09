package gosocket

// Codec
type Codec interface {
	// Encode body to bytes
	Encode(message interface{}) ([]byte, error)
	// Decode from bytes
	Decode(bytes []byte) (interface{}, error)
}
