package internal

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"time"
)

func SetPosition(name string) int {
	switch name {
	case "acoustic_1":
		return -2
	case "acoustic_2":
		return -1
	case "acoustic":
		return 0
	case "electric":
		return 1
	case "encore":
		return 999
	}
	// extract the number from "set_1", "set_2", "set_3"
	var n int
	_, err := fmt.Sscanf(name, "set_%d", &n)
	if err != nil {
		log.Printf("warning: unrecognized set name %q, sorting last", name)
		return 1000
	}
	return n
}

func SortSetPositions(rawEntries []ShowSortInput) (ShowResponse, error) {
	if len(rawEntries) < 1 {
		return ShowResponse{}, errors.New("could not sort set positions: rawEntries slice is empty")
	}

	setsMap := map[string][]string{}
	var venue string
	var date time.Time

	for _, row := range rawEntries {
		venue = row.Venue
		date = row.Date
		setsMap[row.SetName] = append(setsMap[row.SetName], row.RawEntry)
	}

	setNames := make([]string, 0, len(setsMap))
	for k := range setsMap {
		setNames = append(setNames, k)
	}
	sort.Slice(setNames, func(i, j int) bool {
		return SetPosition(setNames[i]) < SetPosition(setNames[j])
	})

	sets := []SetResponse{}
	for _, set := range setNames {
		sets = append(sets, SetResponse{
			SetName: set,
			Songs:   setsMap[set],
		})
	}
	return ShowResponse{
		ShowMeta: ShowMeta{
			ShowID: rawEntries[0].ShowID,
			Date:   date.Format(time.DateOnly),
			Venue:  venue,
			City:   rawEntries[0].City,
			State:  rawEntries[0].State,
			Notes:  rawEntries[0].Notes,
		},
		Sets: sets,
	}, nil
}
