package importer

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/LunarDrift/deadabase/internal"
)

func LoadFile(filename string) (internal.Dataset, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var data internal.Dataset
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	return data, nil
}
