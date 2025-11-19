package main

import (
	"encoding/json"
	"os"
)

func loadData(path string) (map[string]interface{}, error) {
	__tmp0, __err0 := os.ReadFile(path)

	if __err0 != nil {
		return nil, __err0
	}
	// dingo:e:1
	var data = __tmp0
	var result map[string]interface{}
	__tmp1, __err1 := json.Unmarshal(data, &result)

	if __err1 != nil {
		return nil, __err1
	}
	// dingo:e:1
	var err = __tmp1
	return result, nil
}
