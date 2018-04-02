package main

import (
	"fmt"
	"log"

	"github.com/AlexCrane/uk-grid-carbon-intensity"
)

const (
	// Each char will represent charRepSize gCO2/KWh
	charRepSize  = 10
	forecastChar = '-'
	actualChar   = '*'
)

func main() {
	handler := carbonintensity.NewCarbonIntensityAPIHandler()

	intensityArray, err := handler.GetTodaysIntensity()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%c represents forecast carbon intensity. Each %c represents %d gCO2/KWh\n", forecastChar, forecastChar, charRepSize)
	fmt.Printf("%c represents actual carbon intensity. Each %c represents %d gCO2/KWh\n\n", actualChar, actualChar, charRepSize)

	timeColumnLen := len(fmt.Sprintf("%s->%s ", "15:04", "15:04"))

	for i := 0; i < timeColumnLen; i++ {
		fmt.Printf("%c", ' ')
	}
	for i := 1; i < 80-timeColumnLen; i++ {
		if i == 5 || (i%10) == 0 {
			printed := fmt.Sprintf("%c%dg", rune(0x25BC), i*charRepSize)
			fmt.Print(printed)
			i += len([]rune(printed))
		} else {
			fmt.Printf("%c", ' ')
		}
	}
	fmt.Print("\n\n")

	for _, intensity := range intensityArray {
		fmt.Printf("%s->%s ", intensity.From.Format("15:04"), intensity.To.Format("15:04"))
		for i := 0; i < intensity.Forecast; i += charRepSize {
			fmt.Printf("%c", forecastChar)
		}
		fmt.Print("\n")
		if intensity.Actual != -1 {
			for i := 0; i < timeColumnLen; i++ {
				fmt.Printf("%c", ' ')
			}
			for i := 0; i < intensity.Actual; i += charRepSize {
				fmt.Printf("%c", actualChar)
			}
			fmt.Print("\n")
		} else {
			fmt.Print("\n")
		}
	}
}
