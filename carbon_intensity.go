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

	indexLow      = "low"
	indexModerate = "moderate"
	indexHigh     = "high"
)

type APIHandler struct {
	serverAddress string
}

type intensityResponse struct {
	entries []*IntensityEntry
}

type IntensityEntry struct {
	From      time.Time
	To        time.Time
	Intensity Intensity
}

type Intensity struct {
	Forecast int
	Actual   int
	Index    string
}

type statisticsResponse struct {
	entries []*StatisticsEntry
}

type StatisticsEntry struct {
	From  time.Time
	To    time.Time
	Stats Statistics
}

type Statistics struct {
	Max     int
	Average int
	Min     int
	Index   string
}

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
	ir.entries = make([]*IntensityEntry, 0, len(decodedData))

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

		newEntry := &IntensityEntry{To: toTime,
			From: fromTime,
			Intensity: Intensity{
				Forecast: unmarshalInt(decodedIntensity["forecast"], -1),
				Actual:   unmarshalInt(decodedIntensity["actual"], -1),
				Index:    decodedIntensity["index"].(string),
			},
		}

		ir.entries = append(ir.entries, newEntry)
	}

	return nil
}

func (ie *IntensityEntry) String() string {
	return fmt.Sprintf("%s -> %s {forecast: %d, actual: %d, index: %s}", ie.From.Format(natGridTimeFormat), ie.To.Format(natGridTimeFormat),
		ie.Intensity.Forecast, ie.Intensity.Actual, ie.Intensity.Index)
}

func (ah *APIHandler) getAPIResponse(resource string) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf("%s%s", ah.serverAddress, resource))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (ah *APIHandler) GetIntensityForDay(date time.Time) ([]*IntensityEntry, error) {
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

func (ah *APIHandler) GetIntensityForDayAndSettlementPeriod(date time.Time, settlementPeriod int) (*IntensityEntry, error) {
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

func (ah *APIHandler) GetTodaysIntensity() ([]*IntensityEntry, error) {
	return ah.GetIntensityForDay(time.Now())
}

func (ah *APIHandler) GetIntensityForTimePeriod(time time.Time) (*IntensityEntry, error) {
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

func (ah *APIHandler) GetCurrentIntensity() (*IntensityEntry, error) {
	return ah.GetIntensityForTimePeriod(time.Now())
}

func (ah *APIHandler) GetIntensityBetween(from time.Time, to time.Time) ([]*IntensityEntry, error) {
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

// The following three could easily be implemented using GetIntensityBetween , but as they are
// specific resources provided by the REST API, let's have them using those specific resources

func (ah *APIHandler) GetNext24HourIntensity(from time.Time) ([]*IntensityEntry, error) {
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

func (ah *APIHandler) GetNext48HourIntensity(from time.Time) ([]*IntensityEntry, error) {
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

func (ah *APIHandler) GetPrior24HourIntensity(from time.Time) ([]*IntensityEntry, error) {
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
	sr.entries = make([]*StatisticsEntry, 0, len(decodedData))

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

		newEntry := &StatisticsEntry{To: toTime,
			From: fromTime,
			Stats: Statistics{
				Max:     unmarshalInt(decodedIntensity["max"], -1),
				Average: unmarshalInt(decodedIntensity["average"], -1),
				Min:     unmarshalInt(decodedIntensity["min"], -1),
				Index:   decodedIntensity["index"].(string),
			},
		}

		sr.entries = append(sr.entries, newEntry)
	}

	return nil
}

func (se *StatisticsEntry) String() string {
	return fmt.Sprintf("%s -> %s {max: %d, average: %d, min %d, index: %s}", se.From.Format(natGridTimeFormat), se.To.Format(natGridTimeFormat),
		se.Stats.Max, se.Stats.Average, se.Stats.Min, se.Stats.Index)
}

func (ah *APIHandler) GetStatistics(from time.Time, to time.Time) (*StatisticsEntry, error) {
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

func (ah *APIHandler) GetStatisticsInBlocks(from time.Time, to time.Time, blockSize time.Duration) ([]*StatisticsEntry, error) {
	responseBytes, err := ah.getAPIResponse(fmt.Sprintf("/intensity/stats/%s/%s/%d", from.Format(natGridTimeFormat), to.Format(natGridTimeFormat), int(blockSize.Hours())))
	if err != nil {
		return nil, err
	}

	response := statisticsResponse{}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, err
	}

	return response.entries, nil
}
