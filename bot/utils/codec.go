package utils

import "encoding/json"

type JSONCodec struct{}

func (c JSONCodec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (c JSONCodec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
