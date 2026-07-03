# Deadabase
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev/)[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-4169E1?logo=postgresql&logoColor=white)](https://www.postgresql.org/)[![Docker](https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=white)](https://www.docker.com/)[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Deadabase is a REST API for browsing and exploring Grateful Dead concert history. Search through thousands of shows by date, venue, song, or set name - with full setlists and show footnotes. A data importer script populates the database from a JSON source. The API server then provides read-only access to the data.

Show data was sourced from [Grateful Sets](https://www.gratefulsets.net/). Many thanks to them for this wonderful data!

## Features

- **Fuzzy search** for shows by venue or song name - partial names work
- Browse full setlists with **set positions** (Set 1, Set 2, Encore, etc.)
- **Song footnotes** - annotations for notable moments like song debuts or first-time performances
- Filter shows by **date range**, **set name**, **year**, **state/country**, or whether they include **show notes**
- **Song statistics** - times played, first and last performance dates
- Most played songs across all shows or filtered by **set name**
- Unique song counts **per city**

## Tech Stack

- **Go:** HTTP server
- **PostgreSQL:** persistent data storage
- **goose:** database migrations
- **sqlc:** type-safe SQL query generation
- **Docker:** package everything together and make setup a simple command

## Setup
```bash
git clone https://github.com/LunarDrift/gddb
cd gddb
docker compose up --build
```
Use `--build` on first run or after any change to the Go source or Dockerfile, so `docker-compose` rebuilds the `deadabase:local` app image before starting the containers.
*The Postgres credentials are hardcoded in `docker-compose.yml` for local/demo purposes*.

## API Endpoints (GET only)

| ENDPOINT | DESCRIPTION|
| ------------------- | --------------------------------------------------------------------|
| `/shows/{id}` | Search for a specific show by its ID |
| `/shows/{date}` | Search for a show by date (YYYY-MM-DD format) |
| `/shows?song=` | Search for shows where a specific song was played. Returns a list of shows |
| `/shows?set_name=` | Search for shows by set name (set_1, set_2, set_3, encore, acoustic, electric) |
| `/shows?venue=` | Search for shows by venue. Returns a list of shows with their IDs |
| `/shows?has_notes=true/false` | Search for shows with/without notes attached |
| `/shows?start_date=&end_date=` | List of shows between two dates (YYYY-MM-DD format) |
| `/shows?year=` | List of shows filtered by year |
| `/shows?state=` | List of shows filtered by state/country (states should be the standard 2 letter abbreviations) |
| `/shows?year=&state=` | List of shows filtered by year and state/country (states should be the standard 2 letter abbreviations) |
| `/shows/random` | Get details for a random show |
| `/songs?sort=most_played` | Returns a list of all songs and the amount of times they were played |
| `/songs?venue=` | All songs played at a specific venue |
| `/songs?played_lt=n` | Songs played less than `n` times |
| `/songs?sort=most_played&set_name=` | Most played songs by set name (set_1, set_2, set_3, encore, acoustic, electric) |
| `/songs/{name}` | Stats for a song (times played, first/last time played) |
| `/stats/songs-per-city` | Unique song count per city |

For further details, take a look at the [wiki](https://github.com/LunarDrift/gddb/wiki).

## Rate Limiting
Requests are limited per IP address to **2 requests/second** (burst up to 10). Exceeding this returns a `429 Too Many Requests`

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

## License
This project is licensed under the MIT license - see the [LICENSE](LICENSE) file for details.
