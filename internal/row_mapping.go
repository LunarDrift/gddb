package internal

import (
	"time"

	"github.com/LunarDrift/deadabase/internal/database"
)

// These are the mapping functions for the row interfaces defined in database/rows

func RowToShowSortInput(r database.ShowRow) ShowSortInput {
	return ShowSortInput{
		ShowID:   r.GetShowID(),
		Date:     r.GetShowDate(),
		Venue:    r.GetVenue(),
		City:     r.GetCity(),
		State:    r.GetState(),
		Notes:    r.GetNotes().String,
		SetName:  r.GetSetName().String,
		RawEntry: r.GetRawEntry().String,
	}
}

func RowToSongsTimesPlayed(r database.SongCountRow) SongsTimesPlayed {
	return SongsTimesPlayed{
		Song:        r.GetSongName().String,
		TimesPlayed: int(r.GetTimesPlayed()),
	}
}

func RowToShowMeta(r database.ShowSummaryRow) ShowMeta {
	return ShowMeta{
		ShowID: r.GetShowID(),
		Date:   r.GetShowDate().Format(time.DateOnly),
		Venue:  r.GetVenue(),
		City:   r.GetCity(),
		State:  r.GetState(),
		Notes:  r.GetNotes().String,
	}
}
