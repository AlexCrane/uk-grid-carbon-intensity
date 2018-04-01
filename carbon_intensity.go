// Package carbonintensity provides a wrapper around the national grid carbon intensity API - see https://api.carbonintensity.org.uk/
package carbonintensity

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	natGridServerAddress = "https://api.carbonintensity.org.uk"
	natGridTimeFormat    = "2006-01-02T15:04Z07:00"

	indexVeryLow  = "very low"
	indexLow      = "low"
	indexModerate = "moderate"
	indexHigh     = "high"
	indexVeryHigh = "very high"
)

// APIHandler is the struct which provides functions for querying the carbon intensity API
type APIHandler struct {
	serverAddress string
}

type intensityResponse struct {
	entries []*Intensity
}

// Intensity respresents a result from the 'national carbon intensity' party of the API
// It represents forecast and estimated actual carbon intensity for a period of time, given by From and To
// Forecast and Actual are in units of gCO2/KWh
// Index is a string in the set { indexVeryLow, indexLow, indexModerate, indexHigh, indexVeryHigh }
// For period in the future, Forecast will be set but Actual wont be - it will be set to -1. In this case Index will be based off the forecast.
type Intensity struct {
	From     time.Time
	To       time.Time
	Forecast int
	Actual   int
	Index    string
}

type statisticsResponse struct {
	entries []*Statistics
}

// Statistics respresents a result from the 'national statistics' for a period of time, given by From and To
// Max Average and Min are the obvious statistical values for carbon intensity over the given period, in units of gCO2/Kwh
// Index is a string in the set { indexVeryLow, indexLow, indexModerate, indexHigh, indexVeryHigh }
// Future periods use forecast data. Past data uses actual data.
type Statistics struct {
	From    time.Time
	To      time.Time
	Max     int
	Average int
	Min     int
	Index   string
}

// IntensityFactors represents Carbon intensity factors used for different fuel types in the carbon intensity estimations
// Units are gCO2/KWh (grams of CO2 per kilowatt hour)
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

// NewCarbonIntensityAPIHandler returns an APIHandler ready to make queries of the national grid carbon intensity API server
func NewCarbonIntensityAPIHandler() *APIHandler {
	return newCarbonIntensityAPIHandlerInternal(natGridServerAddress)
}

// Allow for a test server to be provided
func newCarbonIntensityAPIHandlerInternal(serverAddress string) *APIHandler {
	return &APIHandler{
		serverAddress: serverAddress,
	}
}

func unmarshalInt(val interface{}, valIfNil int) int {
	if val == nil {
		return valIfNil
	}

	return int(val.(float64))
}

func (ir *intensityResponse) UnmarshalJSON(data []byte) error {
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	if decoded["data"] == nil {
		if decoded["error"] == nil {
			return fmt.Errorf("Failed to unmarshal JSON; %s", string(data))
		}

		errorMap := decoded["error"].(map[string]interface{})
		return fmt.Errorf("API error; Code: %s Message: %s", errorMap["code"].(string), errorMap["message"].(string))
	}

	decodedData := decoded["data"].([]interface{})
	ir.entries = make([]*Intensity, 0, len(decodedData))

	for _, value := range decodedData {
		decodedDataEntry := value.(map[string]interface{})

		toTime, err := time.Parse(natGridTimeFormat, decodedDataEntry["to"].(string))
		if err != nil {
			return nil
		}

		fromTime, err := time.Parse(natGridTimeFormat, decodedDataEntry["from"].(string))
		if err != nil {
			return nil
		}

		decodedIntensity := decodedDataEntry["intensity"].(map[string]interface{})

		newEntry := &Intensity{To: toTime,
			From:     fromTime,
			Forecast: unmarshalInt(decodedIntensity["forecast"], -1),
			Actual:   unmarshalInt(decodedIntensity["actual"], -1),
			Index:    decodedIntensity["index"].(string),
		}

		ir.entries = append(ir.entries, newEntry)
	}

	return nil
}

func (ie *Intensity) String() string {
	return fmt.Sprintf("%s -> %s {forecast: %d, actual: %d, index: %s}", ie.From.Format(natGridTimeFormat),
		ie.To.Format(natGridTimeFormat), ie.Forecast, ie.Actual, ie.Index)
}

func (ah *APIHandler) getAPIResponse(resource string) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf("%s%s", ah.serverAddress, resource))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

// GetIntensityForDay returns an array of Intensity objects, for all 30 minute settlement periods in day represented by date
func (ah *APIHandler) GetIntensityForDay(date time.Time) ([]*Intensity, error) {
	year, month, day := date.Date()

	responseBytes, err := ah.getAPIResponse(fmt.Sprintf("/intensity/date/%04d-%02d-%02d", year, month, day))
	if err != nil {
		return nil, err
	}

	response := intensityResponse{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, err
	}

	return response.entries, nil
}

// GetIntensityForDayAndSettlementPeriod returns an Intensity object, for the given 30 minute settlement period (settlementPeriod) in the day represented by date
// National grid split the day into 48 half-hour settlement periods
// The periods of the day follow UK local time
// The settlement periods are 1-index (numbered 1 to 48 inclusive)
func (ah *APIHandler) GetIntensityForDayAndSettlementPeriod(date time.Time, settlementPeriod int) (*Intensity, error) {
	if settlementPeriod < 1 || settlementPeriod > 48 {
		return nil, fmt.Errorf("Invalid settlmentPeriod %d; must be 1 <= settlementPeriod <= 48", settlementPeriod)
	}

	year, month, day := date.Date()

	responseBytes, err := ah.getAPIResponse(fmt.Sprintf("/intensity/date/%04d-%02d-%02d/%d", year, month, day, settlementPeriod))
	if err != nil {
		return nil, err
	}

	response := intensityResponse{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, err
	}

	if len(response.entries) != 1 {
		return nil, fmt.Errorf("Unexpected API response; unexpected number of entries; %s", string(responseBytes))
	}

	return response.entries[0], nil
}

// GetTodaysIntensity returns an array of Intensity objects, for all 30 minute settlement periods in the current day
// I strongly considered implementing this as GetIntensityForDay(time.Now()) but I will use the dedicated /intensity/date resource
// provided by the API. I would be very interested if the behaviour of these would ever differ (presumably round trip delay could cause this)
func (ah *APIHandler) GetTodaysIntensity() ([]*Intensity, error) {
	responseBytes, err := ah.getAPIResponse("/intensity/date")
	if err != nil {
		return nil, err
	}

	response := intensityResponse{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, err
	}

	return response.entries, nil
}

// GetIntensityForTimePeriod returns an Intensity object, for the 30 minute settlement period containing time
func (ah *APIHandler) GetIntensityForTimePeriod(time time.Time) (*Intensity, error) {
	responseBytes, err := ah.getAPIResponse(fmt.Sprintf("/intensity/%s", time.Format(natGridTimeFormat)))
	if err != nil {
		return nil, err
	}

	response := intensityResponse{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, err
	}

	if len(response.entries) != 1 {
		return nil, fmt.Errorf("Unexpected API response; unexpected number of entries; %s", string(responseBytes))
	}

	return response.entries[0], nil
}

// GetCurrentIntensity returns an Intensity object, for the current 30 minute settlement period
// I strongly considered implementing this as GetIntensityForTimePeriod(time.Now()) but I will use the dedicated /intensity resource
// provided by the API. I would be very interested if the behaviour of these would ever differ (presumably round trip delay could cause this)
func (ah *APIHandler) GetCurrentIntensity() (*Intensity, error) {
	responseBytes, err := ah.getAPIResponse("/intensity")
	if err != nil {
		return nil, err
	}

	response := intensityResponse{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, err
	}

	if len(response.entries) != 1 {
		return nil, fmt.Errorf("Unexpected API response; unexpected number of entries; %s", string(responseBytes))
	}

	return response.entries[0], nil
}

// GetIntensityBetween returns an array of Intensity objects, for all 30 minute settlement periods between from and to
// The maximum date range is limited to 30 days
func (ah *APIHandler) GetIntensityBetween(from time.Time, to time.Time) ([]*Intensity, error) {
	if !from.Before(to) {
		return nil, fmt.Errorf("from (%s) must be strictly earlier than to (%s)", from.String(), to.String())
	}

	if to.Sub(from) > (time.Hour * 24 * 30) {
		return nil, fmt.Errorf("The maximum date range is limited to 30 days. From (%s) To (%s)", from.String(), to.String())
	}

	responseBytes, err := ah.getAPIResponse(fmt.Sprintf("/intensity/%s/%s", from.Format(natGridTimeFormat), to.Format(natGridTimeFormat)))
	if err != nil {
		return nil, err
	}

	response := intensityResponse{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, err
	}

	return response.entries, nil
}

// GetNext24HourIntensity returns an array of Intensity objects, for all 30 minute settlement periods between from and from+24h
// While this could be implemented using GetIntensityBetween it uses the dedicated /intensity/{from}/fw24h resource
func (ah *APIHandler) GetNext24HourIntensity(from time.Time) ([]*Intensity, error) {
	responseBytes, err := ah.getAPIResponse(fmt.Sprintf("/intensity/%s/fw24h", from.Format(natGridTimeFormat)))
	if err != nil {
		return nil, err
	}

	response := intensityResponse{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, err
	}

	return response.entries, nil
}

// GetNext48HourIntensity returns an array of Intensity objects, for all 30 minute settlement periods between from and from+48h
// While this could be implemented using GetIntensityBetween it uses the dedicated /intensity/{from}/fw48h resource
func (ah *APIHandler) GetNext48HourIntensity(from time.Time) ([]*Intensity, error) {
	responseBytes, err := ah.getAPIResponse(fmt.Sprintf("/intensity/%s/fw48h", from.Format(natGridTimeFormat)))
	if err != nil {
		return nil, err
	}

	response := intensityResponse{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, err
	}

	return response.entries, nil
}

// GetPrior24HourIntensity returns an array of Intensity objects, for all 30 minute settlement periods between from-24h and from
// While this could be implemented using GetIntensityBetween it uses the dedicated /intensity/{from}/pt24h resource
func (ah *APIHandler) GetPrior24HourIntensity(from time.Time) ([]*Intensity, error) {
	responseBytes, err := ah.getAPIResponse(fmt.Sprintf("/intensity/%s/pt24h", from.Format(natGridTimeFormat)))
	if err != nil {
		return nil, err
	}

	response := intensityResponse{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, err
	}

	return response.entries, nil
}

// GetIntensityFactors gets an IntensityFactors struct
func (ah *APIHandler) GetIntensityFactors() (*IntensityFactors, error) {
	responseBytes, err := ah.getAPIResponse("/intensity/factors")
	if err != nil {
		return nil, err
	}

	var response map[string]interface{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, err
	}

	if response["data"] == nil {
		if response["error"] == nil {
			return nil, fmt.Errorf("Failed to unmarshal JSON; %s", string(responseBytes))
		}

		errorMap := response["error"].(map[string]interface{})
		return nil, fmt.Errorf("API error; Code: %s Message: %s", errorMap["code"].(string), errorMap["message"].(string))
	}

	responseData := response["data"].([]interface{})

	if len(responseData) != 1 {
		return nil, fmt.Errorf("Unexpected API response; unexpected number of entries; %s", string(responseBytes))
	}

	factorDict := responseData[0].(map[string]interface{})

	return &IntensityFactors{
		Biomass:          int(factorDict["Biomass"].(float64)),
		Coal:             int(factorDict["Coal"].(float64)),
		DutchImports:     int(factorDict["Dutch Imports"].(float64)),
		FrenchImports:    int(factorDict["French Imports"].(float64)),
		GasCombinedCycle: int(factorDict["Gas (Combined Cycle)"].(float64)),
		GasOpenCycle:     int(factorDict["Gas (Open Cycle)"].(float64)),
		Hydro:            int(factorDict["Hydro"].(float64)),
		IrishImports:     int(factorDict["Irish Imports"].(float64)),
		Nuclear:          int(factorDict["Nuclear"].(float64)),
		Oil:              int(factorDict["Oil"].(float64)),
		Other:            int(factorDict["Other"].(float64)),
		PumpedStorage:    int(factorDict["Pumped Storage"].(float64)),
		Solar:            int(factorDict["Solar"].(float64)),
		Wind:             int(factorDict["Wind"].(float64)),
	}, nil
}

func (sr *statisticsResponse) UnmarshalJSON(data []byte) error {
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	if decoded["data"] == nil {
		if decoded["error"] == nil {
			return fmt.Errorf("Failed to unmarshal JSON; %s", string(data))
		}

		errorMap := decoded["error"].(map[string]interface{})
		return fmt.Errorf("API error; Code: %s Message: %s", errorMap["code"].(string), errorMap["message"].(string))
	}

	decodedData := decoded["data"].([]interface{})
	sr.entries = make([]*Statistics, 0, len(decodedData))

	for _, value := range decodedData {
		decodedDataEntry := value.(map[string]interface{})

		toTime, err := time.Parse(natGridTimeFormat, decodedDataEntry["to"].(string))
		if err != nil {
			return nil
		}

		fromTime, err := time.Parse(natGridTimeFormat, decodedDataEntry["from"].(string))
		if err != nil {
			return nil
		}

		decodedIntensity := decodedDataEntry["intensity"].(map[string]interface{})

		newEntry := &Statistics{To: toTime,
			From:    fromTime,
			Max:     unmarshalInt(decodedIntensity["max"], -1),
			Average: unmarshalInt(decodedIntensity["average"], -1),
			Min:     unmarshalInt(decodedIntensity["min"], -1),
			Index:   decodedIntensity["index"].(string),
		}

		sr.entries = append(sr.entries, newEntry)
	}

	return nil
}

func (se *Statistics) String() string {
	return fmt.Sprintf("%s -> %s {max: %d, average: %d, min %d, index: %s}", se.From.Format(natGridTimeFormat), se.To.Format(natGridTimeFormat),
		se.Max, se.Average, se.Min, se.Index)
}

// GetStatistics returns a Statistics object giving carbon intensity statistics for the period between from and to
// The maximum date range is limited to 30 days
func (ah *APIHandler) GetStatistics(from time.Time, to time.Time) (*Statistics, error) {
	if !from.Before(to) {
		return nil, fmt.Errorf("from (%s) must be strictly earlier than to (%s)", from.String(), to.String())
	}

	if to.Sub(from) > (time.Hour * 24 * 30) {
		return nil, fmt.Errorf("The maximum date range is limited to 30 days. From (%s) To (%s)", from.String(), to.String())
	}

	responseBytes, err := ah.getAPIResponse(fmt.Sprintf("/intensity/stats/%s/%s", from.Format(natGridTimeFormat), to.Format(natGridTimeFormat)))
	if err != nil {
		return nil, err
	}

	response := statisticsResponse{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, err
	}

	if len(response.entries) != 1 {
		return nil, fmt.Errorf("Unexpected API response; unexpected number of entries; %s", string(responseBytes))
	}

	return response.entries[0], nil
}

// GetStatisticsInBlocks returns an array of Statistics object giving carbon intensity statistics for the period between from and to
// Each Statistic object in the array covers a period of time given by blockSize
// The maximum date range is limited to 30 days
// The block size given by blockSize is rounded down to the nearest hour and must be between 1 and 24 inclusive
func (ah *APIHandler) GetStatisticsInBlocks(from time.Time, to time.Time, blockSize time.Duration) ([]*Statistics, error) {
	if !from.Before(to) {
		return nil, fmt.Errorf("from (%s) must be strictly earlier than to (%s)", from.String(), to.String())
	}

	if to.Sub(from) > (time.Hour * 24 * 30) {
		return nil, fmt.Errorf("The maximum date range is limited to 30 days. From (%s) To (%s)", from.String(), to.String())
	}

	blockSizeHours := int(blockSize.Hours())

	if blockSizeHours < 1 || blockSizeHours > 24 {
		return nil, fmt.Errorf("Invalid blocksize %s; must be between 1 and 24 hours inclusive", blockSize.String())
	}

	responseBytes, err := ah.getAPIResponse(fmt.Sprintf("/intensity/stats/%s/%s/%d", from.Format(natGridTimeFormat),
		to.Format(natGridTimeFormat), blockSizeHours))
	if err != nil {
		return nil, err
	}

	response := statisticsResponse{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, err
	}

	return response.entries, nil
}
