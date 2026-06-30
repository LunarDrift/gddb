# Deadabase

Deadabase is a structured database + API for browsing Grateful Dead show history, setlists, and venues. Still in early development.

## Features

- Search for shows by date, venue, or ID
- Endpoints to get a random show or view most played songs
- Search for shows between two dates
- Search for songs played less than `n` times

## Tech Stack

- **Go:** HTTP server
- **PostgreSQL:** persistent data storage
- **goose:** database migrations
- **sqlc:** type-safe SQL query generation

## API Endpoints (GET only)

| ENDPOINT            | DESCRIPTION                                                          |
| ------------------- | -------------------------------------------------------------------- |
| `/shows/:value`      | Search for a specific show by show ID or date (YYYY-MM-DD format)                                   |
|`/shows?song=` | Search for shows where a specific song was played. Returns a list of shows |
|`/shows?set_name=` | Search for shows by set name (set_1, set_2, set_3, encore, acoustic, electric) |
| `/shows?venue=`     | Search for shows by venue. Returns a list of shows with their IDs    |
|`/shows?has_notes=true/false`| Search for shows with/without notes attached |
| `/shows?startdate=&enddate=` | List of shows between two dates (YYYY-MM-DD format)|
| `/shows/random`     | Get details for a random show                                      |
| `/songs?sort=mostplayed` | Returns a list of all songs and the amount of times they were played |
| `/songs?venue=`    | All songs played at a specific venue                    |
| `/songs?played_lt=n`    | Songs played less than `n` times                        |
| `/songs?sort=most_played&set_name=`    | Most played songs by set name                                |
| `/songs/:name`          | Stats for a song (times played, first/last time played) |
| `/stats/songs-per-city` | Unique song count per city                                   |

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

## Planned Endpoints

| ENDPOINT                | DESCRIPTION                                             |
| ----------------------- | ------------------------------------------------------- |
| `/shows?state=&city=`  | Search for songs in a specific state/city               |
