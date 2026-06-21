// Package importer used to import JSON data into PostgreSQL database
package importer

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/LunarDrift/deadabase/internal"
	"github.com/LunarDrift/deadabase/internal/database"
)

func Run(db *sql.DB, filename string) error {
	data, err := LoadFile(filename)
	if err != nil {
		return fmt.Errorf("error loading file: %w", err)
	}

	return ImportShows(db, data)
}

func ImportShows(db *sql.DB, data internal.Dataset) error {
	for _, shows := range data {
		for _, show := range shows {
			err := ImportShow(db, show)
			if err != nil {
				return fmt.Errorf("error importing show: %w", err)
			}
		}
	}
	return nil
}

func ImportShow(db *sql.DB, show internal.Show) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting SQL transaction: %w", err)
	}
	defer tx.Rollback()

	q := database.New(db).WithTx(tx)

	showLocation := strings.Split(show.Location, ",")
	city := showLocation[0]
	state := strings.TrimSpace(showLocation[1])

	parsedShowDate, err := time.Parse("2006-01-02", show.Date)
	if err != nil {
		return fmt.Errorf("show_id: %d: failed parsing show date %q: %w", show.ShowID, show.Date, err)
	}

	err = q.CreateShow(context.Background(), database.CreateShowParams{
		ShowID:   int32(show.ShowID),
		ShowDate: parsedShowDate,
		Day:      sql.NullString{String: show.Day, Valid: show.Day != ""},
		City:     city,
		State:    state,
		Venue:    show.Venue,
		Notes:    sql.NullString{String: show.Notes, Valid: show.Notes != ""},
	})
	if err != nil {
		return fmt.Errorf("error creating show: %w", err)
	}

	keys := make([]string, 0, len(show.Setlist))
	for k := range show.Setlist {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return internal.SetPosition(keys[i]) < internal.SetPosition(keys[j])
	})

	for i, setName := range keys {
		songs := show.Setlist[setName]
		setID, err := q.CreateSet(context.Background(), database.CreateSetParams{
			ShowID:   int32(show.ShowID),
			SetName:  setName,
			Position: int32(i + 1),
		})
		if err != nil {
			return fmt.Errorf("error creating set: %w", err)
		}

		for j, song := range songs {
			err = q.CreateSetEntry(context.Background(), database.CreateSetEntryParams{
				SetID:    setID,
				RawEntry: song,
				Position: int32(j + 1),
			})
			if err != nil {
				return fmt.Errorf("error creating set entry: %w", err)
			}
		}
	}

	for marker, noteText := range show.Footnotes {
		err = q.CreateFootnote(context.Background(), database.CreateFootnoteParams{
			ShowID:   int32(show.ShowID),
			Marker:   marker,
			NoteText: noteText,
		})
		if err != nil {
			return fmt.Errorf("error creating footnote: %w", err)
		}
	}

	return tx.Commit()
}
