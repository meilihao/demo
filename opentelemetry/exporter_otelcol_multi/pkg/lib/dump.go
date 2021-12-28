package lib

import (
	jsoniter "github.com/json-iterator/go"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

func DumpToJson(any interface{}) []byte {
	data, err := json.Marshal(any)
	if err != nil {
		panic(err)
	}

	return data
}
