// Copyright 2020 @thiinbit. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file

package gosocket

import (
	"context"
	"fmt"
)

type DefaultCodec struct {
}

func (d DefaultCodec) Encode(ctx context.Context, message interface{}, session *Session) ([]byte, error) {
	return []byte(fmt.Sprintf("%v", message)), nil
}

func (d DefaultCodec) Decode(ctx context.Context, bytes []byte, session *Session) (interface{}, error) {
	return string(bytes), nil
}

type ClientDefaultCodec struct {
}

func (d ClientDefaultCodec) Encode(ctx context.Context, message interface{}, cli *TCPClient) ([]byte, error) {
	return []byte(fmt.Sprintf("%v", message)), nil
}

func (d ClientDefaultCodec) Decode(ctx context.Context, bytes []byte, cli *TCPClient) (interface{}, error) {
	return string(bytes), nil
}
