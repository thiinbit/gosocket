// Copyright 2020 @thiinbit. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file

package gosocket

import "fmt"

type DefaultCodec struct {
}

func (d DefaultCodec) Encode(message interface{}) ([]byte, error) {
	return []byte(fmt.Sprintf("%v", message)), nil
}

func (d DefaultCodec) Decode(bytes []byte) (interface{}, error) {
	return string(bytes), nil
}
