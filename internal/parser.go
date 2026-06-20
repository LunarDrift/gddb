package internal

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"time"
)

func SetPosition(name string) int {
	if name == "encore" {
		return 999
	}
	// extract the number from "set_1", "set_2", "set_3"
	var n int
	_, err := fmt.Sscanf(name, "set_%d", &n)
	if err != nil {
		log.Fatal(err)
	}
	return n
}

func SortSetPositions(shows []ShowSortInput) (ShowResponse, error) {
	if len(shows) < 1 {
		return ShowResponse{}, errors.New("shows slice is empty")
	}
	setsMap := map[string][]string{}
	var venue string
	var date time.Time

	for _, row := range shows {
		venue = row.Venue
		date = row.ShowDate
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
	for _, key := range setNames {
		sets = append(sets, SetResponse{
			SetName: key,
			Songs:   setsMap[key],
		})
	}
	return ShowResponse{
		Date:  date.Format("2006-01-02"),
		Venue: venue,
		City:  shows[0].City,
		State: shows[0].State,
		Notes: shows[0].Notes,
		Sets:  sets,
	}, nil
}
