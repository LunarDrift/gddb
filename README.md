# Deadabase

Deadabase is a structured database + API for browsing Grateful Dead show history, setlists, and venues. Still in early development.

## Features
- Search for shows by date, venue, or ID
- Endpoints to get a random show or view most played songs
- Search for shows between two dates

## Tech Stack
- Go: HTTP server
- PostgreSQL: persistent data storage
- goose: database migrations
- sqlc: type-safe SQL query generation

## API Endpoints (GET only)
| ENDPOINT            | DESCRIPTION                                                          |
| ------------------- | -------------------------------------------------------------------- |
| `/shows/:date`      | Search for a specific show by date (YYYY-MM-DD format)                                   |
| `/shows/:id`        | Search for a specific show by ID                                     |
|`/shows?song=` | Search for shows where a specific song was played |
| `/shows/random`     | Get details about a random show                                      |
| `/shows/between?startdate=YYYY-MM-DD&enddate=YYYY-MM-DD` | List of shows between two dates |
| `/venues?name=`     | Search for shows by venue. Returns a list of shows with their IDs    |
| `/songs/mostplayed` | Returns a list of all songs and the amount of times they were played |
| `/songs?played_lt=n`    | Songs played less than `n` times                        |
| `/stats/top-encores`    | Most common encore songs                                |


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
| `/songs/:name`          | Stats for a song (times played, first/last time played) |
| `/shows?state=&city=`  | Search for songs in a specific state/city               |
| `/venue/:name/songs`    | All songs played at a specific venue                    |
| `/stats/songs-per-city` | Unique songs per city                                   |
