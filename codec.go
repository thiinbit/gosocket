// Copyright 2020 @thiinbit. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file

package gosocket

import "context"

// Codec
type Codec interface {
	// Encode body to bytes
	Encode(ctx context.Context, message interface{}, session *Session) ([]byte, error)
	// Decode from bytes
	Decode(ctx context.Context, bytes []byte, session *Session) (interface{}, error)
}

// ClientCodec
type ClientCodec interface {
	// Encode body to bytes
	Encode(ctx context.Context, message interface{}, cli *TCPClient) ([]byte, error)
	// Decode from bytes
	Decode(ctx context.Context, bytes []byte, cli *TCPClient) (interface{}, error)
}
