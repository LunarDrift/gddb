// Package importer used to import JSON data into PostgreSQL database
package importer

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
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
	var songCounter, showCounter int64 = 0, 0
	for _, shows := range data {
		for _, show := range shows {
			err := ImportShow(db, show, &songCounter, &showCounter)
			if err != nil {
				return fmt.Errorf("error importing show: %w", err)
			}
		}
	}
	return nil
}

func ImportShow(db *sql.DB, show internal.Show, songCounter, showCounter *int64) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting SQL transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	q := database.New(db).WithTx(tx)

	showLocation := strings.Split(show.Location, ",")
	city := showLocation[0]
	state := strings.TrimSpace(showLocation[1])

	parsedShowDate, err := time.Parse(time.DateOnly, show.Date)
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
	*showCounter++
	currentShowProgress := *showCounter
	const totalShows = 2309

	if currentShowProgress%231 == 0 || currentShowProgress == totalShows {
		log.Printf("shows: %d/%d\t(%.1f%%)", currentShowProgress, totalShows, float64(currentShowProgress)/float64(2309)*100)
	}

	var songBatch []database.InsertEntrySongParams
	const totalSongs = 23035

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

		for j, rawEntry := range songs {
			entryID, err := q.CreateSetEntry(context.Background(), database.CreateSetEntryParams{
				SetID:    setID,
				RawEntry: rawEntry,
				Position: int32(j + 1),
			})
			if err != nil {
				return fmt.Errorf("error creating set entry: %w", err)
			}

			if err := q.DeleteEntrySongs(context.Background(), int64(entryID)); err != nil {
				return fmt.Errorf("error clearing entry songs: %w", err)
			}

			for _, name := range parseEntry(rawEntry) {
				songBatch = append(songBatch, database.InsertEntrySongParams{
					SetEntryID: int64(entryID),
					SongName:   name,
				})
				// err = q.InsertEntrySong(context.Background(), database.InsertEntrySongParams{
				// 	SetEntryID: int64(entryID),
				// 	SongName:   name,
				// })
				// if err != nil {
				// 	return fmt.Errorf("error creating entry song: %w", err)
				// }
			}
			*songCounter++
			currentProgress := *songCounter

			if currentProgress%2304 == 0 || currentProgress == totalSongs {
				log.Printf("songs: %d/%d\t(%.1f%%)", currentProgress, totalSongs, float64(currentProgress)/float64(totalSongs)*100)
			}
		}
	}

	if len(songBatch) > 0 {
		err = InsertEntrySongsBulk(context.Background(), tx, songBatch)
		if err != nil {
			return fmt.Errorf("failed to bulk insert song batch: %w", err)
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

var markerRe = regexp.MustCompile(`\s*(\*+)\s*$`)

func parseEntry(raw string) []string {
	parts := strings.Split(raw, ">")
	songs := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(markerRe.ReplaceAllString(part, ""))
		if part != "" {
			songs = append(songs, part)
		}
	}
	return songs
}

// InsertEntrySongsBulk The pq SQL driver doesn't support :copyfrom sqlc code generation.
// Takes an entire show's worth of rawEntrys and inserts them into the database as a batch
// instead of a single rawEntry at a time.
func InsertEntrySongsBulk(ctx context.Context, tx *sql.Tx, batch []database.InsertEntrySongParams) error {
	if len(batch) == 0 {
		return nil
	}

	var query strings.Builder
	fmt.Fprint(&query, "INSERT INTO set_entry_songs (set_entry_id, song_name) VALUES ")

	args := make([]any, 0, len(batch)*2)

	for i, song := range batch {
		if i > 0 {
			fmt.Fprint(&query, ", ")
		}
		p1 := i*2 + 1
		p2 := i*2 + 2
		fmt.Fprintf(&query, "($%d, $%d)", p1, p2)

		args = append(args, song.SetEntryID, song.SongName)
	}

	_, err := tx.ExecContext(ctx, query.String(), args...)
	return err
}
