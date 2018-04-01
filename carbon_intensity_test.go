package carbonintensity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TODO: For now tests just function to make sure we don't crash against the real server
// Other tests should be written to point at a test server that will return known data that we can check we are correctly parsing

func TestCurrentIntensity(t *testing.T) {
	handler := NewCarbonIntensityApiHandler()

	intensity, err := handler.GetCurrentIntensity()
	assert.NoError(t, err)
	t.Logf("%v\n", intensity)
}

func TestOtherTimePeriodsIntensity(t *testing.T) {
	handler := NewCarbonIntensityApiHandler()

	intensity, err := handler.GetIntensityForTimePeriod(time.Now().Add(8*time.Hour))
	assert.NoError(t, err)
	t.Logf("%v\n", intensity)

	intensity, err = handler.GetIntensityForTimePeriod(time.Now().Add(-8*time.Hour))
	assert.NoError(t, err)
	t.Logf("%v\n", intensity)
}

func TestOtherDayAndSettlementPeriodIntensity(t *testing.T) {
	handler := NewCarbonIntensityApiHandler()

	// The day is split into 48 half-hour settlement periods
	// Joyfully (and perhaps obviously) the periods of the day follow local time (and so British Summer Time)
	// They are also 1-index (numbered 1 to 48 inclusive)
	// Fortunately time.In() is a pretty sweet function!
	london, err := time.LoadLocation("Europe/London")
	assert.NoError(t, err)
	
	intensity, err := handler.GetIntensityForDayAndSettlementPeriod(time.Now().Add(-24*time.Hour), 1)
	assert.NoError(t, err)
	assert.Equal(t, 0, intensity.From.In(london).Hour())
	assert.Equal(t, 0, intensity.From.In(london).Minute())
	assert.Equal(t, 0, intensity.To.In(london).Hour())
	assert.Equal(t, 30, intensity.To.In(london).Minute())
	t.Logf("%v\n", intensity)

	intensity, err = handler.GetIntensityForDayAndSettlementPeriod(time.Now().Add(-24*time.Hour), 24)
	assert.Equal(t, 11, intensity.From.In(london).Hour())
	assert.Equal(t, 30, intensity.From.In(london).Minute())
	assert.Equal(t, 12, intensity.To.In(london).Hour())
	assert.Equal(t, 00, intensity.To.In(london).Minute())
	assert.NoError(t, err)
	t.Logf("%v\n", intensity)

	intensity, err = handler.GetIntensityForDayAndSettlementPeriod(time.Now().Add(-24*time.Hour), 48)
	// This will only work if you are in GMT timezone and so is a terrible test :D
	assert.Equal(t, 23, intensity.From.In(london).Hour())
	assert.Equal(t, 30, intensity.From.In(london).Minute())
	assert.Equal(t, 0, intensity.To.In(london).Hour())
	assert.Equal(t, 0, intensity.To.In(london).Minute())
	assert.NoError(t, err)
	t.Logf("%v\n", intensity)

	intensity, err = handler.GetIntensityForDayAndSettlementPeriod(time.Now().Add(-24*time.Hour), 0)
	assert.Error(t, err)

	intensity, err = handler.GetIntensityForDayAndSettlementPeriod(time.Now().Add(-24*time.Hour), 49)
	assert.Error(t, err)
}

func TestTodaysIntensity(t *testing.T) {
	handler := NewCarbonIntensityApiHandler()

	intensityArr, err := handler.GetTodaysIntensity()
	assert.NoError(t, err)
	for _, intensity := range intensityArr {
		t.Logf("%v\n", intensity)
	}
}

func TestOtherDaysIntensity(t *testing.T) {
	handler := NewCarbonIntensityApiHandler()

	intensityArr, err := handler.GetIntensityForDay(time.Now().Add(24 * time.Hour))
	assert.NoError(t, err)
	for _, intensity := range intensityArr {
		t.Logf("%v\n", intensity)
	}

	intensityArr, err = handler.GetIntensityForDay(time.Now().Add(-24 * time.Hour))
	assert.NoError(t, err)
	for _, intensity := range intensityArr {
		t.Logf("%v\n", intensity)
	}
}

func TestIntensityBetween(t *testing.T) {
	handler := NewCarbonIntensityApiHandler()

	intensityArr, err := handler.GetIntensityBetween(time.Now().Add(-24 * time.Hour), time.Now().Add(24 * time.Hour))
	assert.NoError(t, err)
	for _, intensity := range intensityArr {
		t.Logf("%v\n", intensity)
	}
}

func TestNext24HourIntensity(t *testing.T) {
	handler := NewCarbonIntensityApiHandler()

	intensityArr, err := handler.GetNext24HourIntensity(time.Now())
	assert.NoError(t, err)
	for _, intensity := range intensityArr {
		t.Logf("%v\n", intensity)
	}
}

func TestNext48HourIntensity(t *testing.T) {
	handler := NewCarbonIntensityApiHandler()

	intensityArr, err := handler.GetNext48HourIntensity(time.Now())
	assert.NoError(t, err)
	for _, intensity := range intensityArr {
		t.Logf("%v\n", intensity)
	}
}

func TestPrior24HourIntensity(t *testing.T) {
	handler := NewCarbonIntensityApiHandler()

	intensityArr, err := handler.GetPrior24HourIntensity(time.Now())
	assert.NoError(t, err)
	for _, intensity := range intensityArr {
		t.Logf("%v\n", intensity)
	}
}

func TestIntensityFactors(t *testing.T) {
	handler := NewCarbonIntensityApiHandler()

	factors, err := handler.GetIntensityFactors()
	assert.NoError(t, err)
	t.Logf("%#v\n", factors)
}

func TestStatistics(t *testing.T) {
	handler := NewCarbonIntensityApiHandler()

	// Let's try a week of stats
	stats, err := handler.GetStatistics(time.Now().Add(time.Hour * (24 * -7)), time.Now())
	assert.NoError(t, err)
	t.Logf("%v\n", stats)
}

func TestStatisticsInBlocks(t *testing.T) {
	handler := NewCarbonIntensityApiHandler()

	// Let's try a week of stats, in 4 hour blocks
	statsArr, err := handler.GetStatisticsInBlocks(time.Now().Add(time.Hour * (24 * -7)), time.Now(), time.Hour * 4)
	assert.NoError(t, err)
	assert.Equal(t, 42, len(statsArr))
	for _, stats := range statsArr {
		t.Logf("%v\n", stats)
	}
}
