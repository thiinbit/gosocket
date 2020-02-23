// Copyright 2020 @thiinbit. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file

package gosocket

// Codec
type Codec interface {
	// Encode body to bytes
	Encode(message interface{}) ([]byte, error)
	// Decode from bytes
	Decode(bytes []byte) (interface{}, error)
}
