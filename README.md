# Satellite Tracker

Satellite Tracker is a terminal-based application written in Go that lets you search for satellites and explore their orbital data using CelesTrak. You can compare satellites, save them to a watchlist, and view saved satellites offline.

## Features

- Search by satellite name or NORAD ID
- Explore satellite orbital characteristics and measurements
- Compare two satellites
- Save satellites to a watchlist
- Offline mode for saved satellites

## Requirements

- Go 1.22 or later
- An internet connection for live satellite data
- A Linux terminal

## How to Run

Clone or download this repository, then open a terminal in the project directory.

Build the application:

```bash
go build
```

Run it:

```bash
./satellite-tracker
```

You can also run it directly without building a binary:

```bash
go run .
```

The app creates a `watchlist.json` file to store your saved satellites.