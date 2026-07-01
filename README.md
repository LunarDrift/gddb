# Deadabase

Deadabase is a REST API for browsing and exploring Grateful Dead concert history. Search through thousands of shows by date, venue, song, or set name - with full setlists and show footnotes. A data importer script populates the database from a JSON source. The API server then provides read-only access to the data.

Show data was sourced from [Grateful Sets](https://www.gratefulsets.net/). Many thanks to them for this wonderful data!

## Features

- **Fuzzy search** for shows by venue or song name - partial names work
- Browse full setlists with **set positions** (Set 1, Set 2, Encore, etc.)
- **Song footnotes** - annotations for notable moments like song debuts or first-time performances
- Filter shows by **date range**, set name, or whether they include show notes
- **Song statistics** - times played, first and last performance dates
- Most played songs across all shows or filtered by **set name**
- Unique song counts **per city**

## Tech Stack

- **Go:** HTTP server
- **PostgreSQL:** persistent data storage
- **goose:** database migrations
- **sqlc:** type-safe SQL query generation

## Planned
- Docker support with `docker compose up` as the only required setup step

## API Endpoints (GET only)

| ENDPOINT | DESCRIPTION|
| ------------------- | --------------------------------------------------------------------|
| `/shows/{value}` | Search for a specific show by show ID or date (YYYY-MM-DD format) |
| `/shows?song=` | Search for shows where a specific song was played. Returns a list of shows |
| `/shows?set_name=` | Search for shows by set name (set_1, set_2, set_3, encore, acoustic, electric) |
| `/shows?venue=` | Search for shows by venue. Returns a list of shows with their IDs |
| `/shows?has_notes=true/false` | Search for shows with/without notes attached |
| `/shows?start_date=&end_date=` | List of shows between two dates (YYYY-MM-DD format) |
| `/shows/random` | Get details for a random show |
| `/songs?sort=most_played` | Returns a list of all songs and the amount of times they were played |
| `/songs?venue=` | All songs played at a specific venue |
| `/songs?played_lt=n` | Songs played less than `n` times |
| `/songs?sort=most_played&set_name=` | Most played songs by set name (set_1, set_2, set_3, encore, acoustic, electric) |
| `/songs/{name}` | Stats for a song (times played, first/last time played) |
| `/stats/songs-per-city` | Unique song count per city |

## Example Show Response

```json
{
  "date": "1993-03-14",
  "venue": "Richfield Coliseum",
  "city": "Richfield",
  "state": "OH",
  "notes": "",
  "sets": [
    {
      "set_name": "set_1",
      "songs": [
        "Cold Rain And Snow",
        "Walkin' Blues",
        "Brown Eyed Women",
        "Just Like Tom Thumb's Blues",
        "Lazy River Road",
        "Eternity > Don't Ease Me In"
      ]
    },
    {
      "set_name": "set_2",
      "songs": [
        "Touch Of Grey",
        "Samson And Delilah",
        "Way To Go Home",
        "Corrina > Terrapin Station > drums > space > I Need A Miracle > Stella Blue > Throwing Stones > Turn On Your Lovelight"
      ]
    },
    {
      "set_name": "encore",
      "songs": [
        "I Fought The Law *"
      ]
    }
  ],
  "footnotes": {
    "*": "First time played"
  }
}
```

## Notes
- Venue and song name searches use fuzzy matching - partial names work
- Shows without a recorded setlist return a custom `message` field instead of `sets`
- `footnotes` in show responses are keyed by marker symbol (e.g. `"*": "First time played"`)

## Credits
Show and setlist data sourced from [Grateful Sets](https://www.gratefulsets.net/)
