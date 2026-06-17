package internal

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/LunarDrift/deadabase/internal/database"
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
		Sets:  sets,
	}, nil
}

func GroupByVenue(rows []database.SearchByVenueRow) []VenueResult {
	// Made by Claude :)
	// TODO: Git gud and understand wtf this is doing
	// Read notes in Obsidian, that might help

	var results []VenueResult

	// These maps let us find an existing entry by key
	// without scanning the slice every time
	venueIndex := map[string]int{}
	showIndex := map[string]int{}
	setIndex := map[string]int{}

	for _, row := range rows {
		venueKey := row.Venue + "|" + row.City + "|" + row.State
		showKey := venueKey + "|" + row.Date.Format("2006-01-02")
		setKey := showKey + "|" + row.SetName

		// Venue level
		if _, exists := venueIndex[venueKey]; !exists {
			results = append(results, VenueResult{
				Venue: row.Venue,
				City:  row.City,
				State: row.State,
			})
			venueIndex[venueKey] = len(results) - 1
		}
		vi := venueIndex[venueKey]

		// Show level
		if _, exists := showIndex[showKey]; !exists {
			results[vi].Shows = append(results[vi].Shows, ShowResult{
				Date:  row.Date.Format("2006-01-02"),
				Notes: row.Notes.String,
			})
			showIndex[showKey] = len(results[vi].Shows) - 1
		}
		si := showIndex[showKey]

		// Set level
		if _, exists := setIndex[setKey]; !exists {
			results[vi].Shows[si].Sets = append(results[vi].Shows[si].Sets, SetResult{
				SetName: row.SetName,
			})
			setIndex[setKey] = len(results[vi].Shows[si].Sets) - 1
		}
		setI := setIndex[setKey]

		// Song level (always just append, no dedup needed)
		results[vi].Shows[si].Sets[setI].Songs = append(results[vi].Shows[si].Sets[setI].Songs, row.Song)
	}

	return results
}
