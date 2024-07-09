# monitoring-tool

The service is using go routines and channels to send out http requests and read responses as soon as they become available to efficiently index data.

## Description

monitoring-tool responsibilities:
- monitors the availability of many websites over the network
- produces metrics about these
- stores the metrics into a SQLite database.

monitoring-tool should:
- perform the checks periodically
- collect the request timestamp, the response time, the HTTP status code
- optionally check the returned page contents for a regex pattern that is expected to be found on the page
- Each URL should be checked periodically, with the ability to configure the interval and the regexp on a per-URL basis
- The monitored URLs can be anything found online.

Checks:
- [x] Code should be simple and understandable. Anything unnecessarily complex, undocumented or untestable will be considered a minus.
- [x] Main design goal is maintainability.
- [x] The solution must work (we need to be able to run the solution)
- [x] Must be tested and have tests
- [x] Must handle errors
- [x] Should be production quality
- [x] Should work for at least some thousands of separate sites (preconfigured)


## Installation

```$ go mod tidy```

## Running the app

```$ go run main.go```

## Running the tests

```go test . -v -count=1```

## Displaying DB entries (sqlite3 required)

```sqlite3 simple.db```

```SELECT * FROM responses;```


Example output

```
SQLite version 3.37.0 2021-12-09 01:34:53
Enter ".help" for usage hints.
sqlite> SELECT * FROM responses;
1|2024-07-09 20:05:46.410391 +0000 UTC|https://www.reuters.com/|401|197|0
2|2024-07-09 20:05:46.410304 +0000 UTC|https://www.bbc.com/|200|193|1
3|2024-07-09 20:05:47.410413 +0000 UTC|https://www.bbc.com/|200|36|1
4|2024-07-09 20:05:47.410514 +0000 UTC|https://www.reuters.com/|401|60|0
5|2024-07-09 20:05:48.410858 +0000 UTC|https://www.reuters.com/|401|37|0
6|2024-07-09 20:05:48.410753 +0000 UTC|https://www.bbc.com/|200|21|1
7|2024-07-09 20:05:49.410572 +0000 UTC|https://www.bbc.com/|200|44|1
8|2024-07-09 20:05:49.410681 +0000 UTC|https://www.reuters.com/|401|64|0
9|2024-07-09 20:05:50.410343 +0000 UTC|https://www.bbc.com/|200|51|1
10|2024-07-09 20:05:50.410446 +0000 UTC|https://www.reuters.com/|401|72|0
11|2024-07-09 20:05:51.409628 +0000 UTC|https://www.bbc.com/|200|37|1
12|2024-07-09 20:05:51.409726 +0000 UTC|https://www.reuters.com/|401|66|0
13|2024-07-09 20:05:52.410077 +0000 UTC|https://www.bbc.com/|200|30|1
14|2024-07-09 20:05:52.41019 +0000 UTC|https://www.reuters.com/|401|54|0
15|2024-07-09 20:05:53.409443 +0000 UTC|https://www.reuters.com/|401|55|0
```

## Thanks for the review!