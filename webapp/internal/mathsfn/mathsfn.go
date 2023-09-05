package mathsfn

import (
	"fmt"
	"time"
)

func GetMode(numbers []float64) float64 {
	counts := make(map[float64]int)

	for _, num := range numbers {
		counts[num]++
	}

	mode := 0.0
	count := 0
	for number, occurrences := range counts {
		if occurrences > count {
			mode = number
			count = occurrences
		}
	}

	return mode
}

func GetDateTime(datetime time.Time) string {
	return fmt.Sprintf("%d-%d-%d %d:%d:%d\n",
		datetime.Year(),
		datetime.Month(),
		datetime.Day(),
		datetime.Hour(),
		datetime.Minute(),
		datetime.Second(),
	)
}

func GetTime(datetime string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", datetime)
}
