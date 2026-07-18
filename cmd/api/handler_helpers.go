package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/LunarDrift/deadabase/internal"
)

// fuzzyPattern wraps and inserts a '%' between every character of `input` to be used during SQL query searches
func fuzzyPattern(input string) string {
	words := strings.Fields(input)
	if len(words) == 0 {
		return ""
	}
	return "%" + strings.Join(words, "%") + "%"
}

func respondWithJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.SetLogLoggerLevel(slog.LevelError)
		log.Println("respondWithJSON - could not encode payload:", err)
	}
}

func respondWithError(w http.ResponseWriter, status int, message string, err error) {
	if err != nil {
		slog.SetLogLoggerLevel(slog.LevelError)
		log.Println(err)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, status, errorResponse{Error: message})
}

// buildShowResponse takes a slice of ShowSortInput rows for a single show and returns
// a 'no setlist available' message if the show has no songs, otherwise sorting the
// sets + songs and attaching any footnotes, returning the completed show response.
func (s *server) buildShowResponse(r *http.Request, parsedShowRows []internal.ShowSortInput) (any, error) {
	if len(parsedShowRows) > 0 && parsedShowRows[0].RawEntry == "" {
		row := parsedShowRows[0]
		return internal.ShowWithNoSetlist{
			ShowMeta: internal.ShowMeta{
				ShowID:   row.ShowID,
				Date:     row.Date.Format(time.DateOnly),
				Venue:    row.Venue,
				City:     row.City,
				Location: row.Location,
				Notes:    row.Notes,
			},
			Message: "No setlist available for this show",
		}, nil
	}

	showResp, err := internal.SortSetPositions(parsedShowRows)
	if err != nil {
		return nil, err
	}

	footnoteRows, err := s.queries.GetFootnotesFromShowID(r.Context(), parsedShowRows[0].ShowID)
	if err != nil {
		return nil, err
	}
	showResp.Footnotes = make(map[string]string)
	for _, f := range footnoteRows {
		showResp.Footnotes[f.Marker] = f.NoteText
	}

	return showResp, nil
}
