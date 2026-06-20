package importer

import (
	"encoding/json"
	"os"

	"github.com/LunarDrift/deadabase/internal"
)

func LoadFile(filename string) (internal.Dataset, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var data internal.Dataset
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
