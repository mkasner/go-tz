# go-tz

tz-lookup by lng and lat

[![GoDoc](https://godoc.org/github.com/ugjka/go-tz?status.svg)](https://godoc.org/github.com/ugjka/go-tz)
[![Go Report Card](https://goreportcard.com/badge/github.com/ugjka/go-tz)](https://goreportcard.com/report/github.com/ugjka/go-tz)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Donate](https://dl.ugjka.net/Donate-PayPal-green.svg)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=UVTCZYQ3FVNCY)

lookup timezone for a given location

```go
// Loading Zone for Line Islands, Kiritimati
zone, err := gotz.GetZone(gotz.Point{
    Lat: 1.74294, Lon: -157.21328,
})
if err != nil {
    panic(err)
}
fmt.Println(zone)
```

```bash
[ugjka@archee example]$ go run main.go
Pacific/Kiritimati
```

Uses simplified shapefile from [timezone-boundary-builder](https://github.com/evansiroky/timezone-boundary-builder/)

GeoJson Simplification done with [mapshaper](http://mapshaper.org/) and [shapefile-geo](https://github.com/foursquare/shapefile-geo)

:)
