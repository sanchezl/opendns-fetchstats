# OpenDNS Stats Fetcher

## Usage

Usage of `fetchstats`:
```
  -date string
        YYYY-MM-DD or YYYY-MM-DDtoYYYY-MM-DD for range
  -network-id string
        Network ID. (default "all")
  -password string
        OpenDNS user password.
  -user string
        OpenDNS email or username.
```

## Description
Automatically fetch your OpenDNS Top Domains data for the given
date range in CSV format.  Fetches all pages and combines them
into one CSV file, which is printed to STDOUT.

## Installation
```
go install github.com/sanchezl/opendns-fetchstats/cmd/fetchstats
```

## Thanks!
Based on scripts written by Richard Crowley <richard at opendns.com> and Brian Hartvigsen <brian.hartvigsen at opendns.com>.
