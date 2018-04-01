# carbonintensity

```go
import "github.com\AlexCrane\uk-grid-carbon-intensity"
```

Package carbonintensity provides a wrapper around the national grid carbon
intensity API - see https://api.carbonintensity.org.uk/

## Usage

#### type APIHandler

```go
type APIHandler struct {
}
```

APIHandler is the struct which provides functions for querying the carbon
intensity API

#### func  NewCarbonIntensityAPIHandler

```go
func NewCarbonIntensityAPIHandler() *APIHandler
```
NewCarbonIntensityAPIHandler returns an APIHandler ready to make queries of the
national grid carbon intensity API server

#### func (*APIHandler) GetCurrentIntensity

```go
func (ah *APIHandler) GetCurrentIntensity() (*Intensity, error)
```
GetCurrentIntensity returns an Intensity object, for the current 30 minute
settlement period

I strongly considered implementing this as GetIntensityForTimePeriod(time.Now())
but I will use the dedicated /intensity resource provided by the API. I would be
very interested if the behaviour of these would ever differ (presumably round
trip delay could cause this)

#### func (*APIHandler) GetIntensityBetween

```go
func (ah *APIHandler) GetIntensityBetween(from time.Time, to time.Time) ([]*Intensity, error)
```
GetIntensityBetween returns an array of Intensity objects, for all 30 minute
settlement periods between from and to

The maximum date range is limited to 30 days

#### func (*APIHandler) GetIntensityFactors

```go
func (ah *APIHandler) GetIntensityFactors() (*IntensityFactors, error)
```
GetIntensityFactors gets an IntensityFactors struct

#### func (*APIHandler) GetIntensityForDay

```go
func (ah *APIHandler) GetIntensityForDay(date time.Time) ([]*Intensity, error)
```
GetIntensityForDay returns an array of Intensity objects, for all 30 minute
settlement periods in day represented by date

#### func (*APIHandler) GetIntensityForDayAndSettlementPeriod

```go
func (ah *APIHandler) GetIntensityForDayAndSettlementPeriod(date time.Time, settlementPeriod int) (*Intensity, error)
```
GetIntensityForDayAndSettlementPeriod returns an Intensity object, for the given
30 minute settlement period (settlementPeriod) in the day represented by date

National grid split the day into 48 half-hour settlement periods. The periods of
the day follow UK local time. The settlement periods are 1-index (numbered 1 to
48 inclusive).

#### func (*APIHandler) GetIntensityForTimePeriod

```go
func (ah *APIHandler) GetIntensityForTimePeriod(time time.Time) (*Intensity, error)
```
GetIntensityForTimePeriod returns an Intensity object, for the 30 minute
settlement period containing time

#### func (*APIHandler) GetNext24HourIntensity

```go
func (ah *APIHandler) GetNext24HourIntensity(from time.Time) ([]*Intensity, error)
```
GetNext24HourIntensity returns an array of Intensity objects, for all 30 minute
settlement periods between from and from+24h

While this could be implemented using GetIntensityBetween it uses the dedicated
/intensity/{from}/fw24h resource

#### func (*APIHandler) GetNext48HourIntensity

```go
func (ah *APIHandler) GetNext48HourIntensity(from time.Time) ([]*Intensity, error)
```
GetNext48HourIntensity returns an array of Intensity objects, for all 30 minute
settlement periods between from and from+48h

While this could be implemented using GetIntensityBetween it uses the dedicated
/intensity/{from}/fw48h resource

#### func (*APIHandler) GetPrior24HourIntensity

```go
func (ah *APIHandler) GetPrior24HourIntensity(from time.Time) ([]*Intensity, error)
```
GetPrior24HourIntensity returns an array of Intensity objects, for all 30 minute
settlement periods between from-24h and from

While this could be implemented using GetIntensityBetween it uses the dedicated
/intensity/{from}/pt24h resource

#### func (*APIHandler) GetStatistics

```go
func (ah *APIHandler) GetStatistics(from time.Time, to time.Time) (*Statistics, error)
```
GetStatistics returns a Statistics object giving carbon intensity statistics for
the period between from and to

The maximum date range is limited to 30 days

#### func (*APIHandler) GetStatisticsInBlocks

```go
func (ah *APIHandler) GetStatisticsInBlocks(from time.Time, to time.Time, blockSize time.Duration) ([]*Statistics, error)
```
GetStatisticsInBlocks returns an array of Statistics object giving carbon
intensity statistics for the period between from and to

Each Statistic object in the array covers a period of time given by blockSize.
The maximum date range is limited to 30 days. The block size given by blockSize
is rounded down to the nearest hour and must be between 1 and 24 inclusive.

#### func (*APIHandler) GetTodaysIntensity

```go
func (ah *APIHandler) GetTodaysIntensity() ([]*Intensity, error)
```
GetTodaysIntensity returns an array of Intensity objects, for all 30 minute
settlement periods in the current day

I strongly considered implementing this as GetIntensityForDay(time.Now()) but I
will use the dedicated /intensity/date resource provided by the API. I would be
very interested if the behaviour of these would ever differ (presumably round
trip delay could cause this).

#### type Intensity

```go
type Intensity struct {
	From     time.Time
	To       time.Time
	Forecast int
	Actual   int
	Index    string
}
```

Intensity respresents a result from the 'national carbon intensity' party of the
API. It represents forecast and estimated actual carbon intensity for a period
of time, given by From and To

Forecast and Actual are in units of gCO2/KWh. Index is a string in the set {
indexVeryLow, indexLow, indexModerate, indexHigh, indexVeryHigh }

For future periods, Forecast will be set but Actual wont be - it will be set to
-1. In this case Index will be based off the forecast

#### func (*Intensity) String

```go
func (ie *Intensity) String() string
```

#### type IntensityFactors

```go
type IntensityFactors struct {
	Biomass          int
	Coal             int
	DutchImports     int
	FrenchImports    int
	IrishImports     int
	GasCombinedCycle int
	GasOpenCycle     int
	Hydro            int
	Nuclear          int
	Oil              int
	Other            int
	PumpedStorage    int
	Solar            int
	Wind             int
}
```

IntensityFactors represents Carbon intensity factors used for different fuel
types in the carbon intensity estimations. Units are gCO2/KWh (grams of CO2 per
kilowatt hour).

#### type Statistics

```go
type Statistics struct {
	From    time.Time
	To      time.Time
	Max     int
	Average int
	Min     int
	Index   string
}
```

Statistics respresents a result from the 'national statistics' for a period of
time, given by From and To

Max Average and Min are the obvious statistical values for carbon intensity over
the given period, in units of gCO2/Kwh. Index is a string in the set {
indexVeryLow, indexLow, indexModerate, indexHigh, indexVeryHigh }. Future
periods use forecast data. Past data uses actual data.

#### func (*Statistics) String

```go
func (se *Statistics) String() string
```
