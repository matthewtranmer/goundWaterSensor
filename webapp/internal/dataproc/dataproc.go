package dataproc

import (
	"database/sql"
	"errors"
	"math"
	"strings"
	"time"

	"website/internal/mathsfn"
	"website/internal/templates"
)

func getLastReadings(db *sql.DB, readings int) ([]float64, error) {
	statement, err := db.Prepare("SELECT height FROM readings ORDER BY time DESC LIMIT ?")
	if err != nil {
		return nil, err
	}

	rows, err := statement.Query(readings)
	if err != nil {
		return nil, err
	}

	var heights []float64
	index := 0
	for rows.Next() {
		heights = append(heights, 0.0)
		rows.Scan(&heights[index])
		heights[index] = math.Round((heights[index])*100) / 100
		index++
	}

	return heights, nil
}

func calculateChanges(db *sql.DB) (percentage int, distance float64, err error) {
	day_ago := time.Now().Add(-24 * time.Hour)
	three_hour_ago := time.Now().Add(-3 * time.Hour)

	statement, err := db.Prepare("SELECT height FROM readings ORDER BY time DESC LIMIT 10")
	if err != nil {
		return -1, -1, err
	}

	rows, err := statement.Query()
	if err != nil {
		return -1, -1, err
	}

	var current_readings []float64
	reading := 0.0
	for rows.Next() {
		err = rows.Scan(&reading)
		current_readings = append(current_readings, reading)
		if err != sql.ErrNoRows && err != nil {
			return -1, -1, err
		}
	}

	current_reading := mathsfn.GetMode(current_readings)

	statement, err = db.Prepare("SELECT height FROM readings WHERE time <= ? ORDER BY time DESC LIMIT 10")
	if err != nil {
		return -1, -1, err
	}

	rows, err = statement.Query(mathsfn.GetDateTime(three_hour_ago))
	if err != nil {
		return -1, -1, err
	}

	var three_hour_ago_readings []float64
	for rows.Next() {
		err = rows.Scan(&reading)
		three_hour_ago_readings = append(three_hour_ago_readings, reading)
		if err != sql.ErrNoRows && err != nil {
			return -1, -1, err
		}
	}

	three_hour_ago_reading := mathsfn.GetMode(three_hour_ago_readings)

	statement, err = db.Prepare("SELECT height FROM readings WHERE time <= ? ORDER BY time DESC LIMIT 10")
	if err != nil {
		return -1, -1, err
	}

	rows, err = statement.Query(mathsfn.GetDateTime(day_ago))
	if err != nil {
		return -1, -1, err
	}

	var day_ago_readings []float64
	for rows.Next() {
		err = rows.Scan(&reading)
		day_ago_readings = append(day_ago_readings, reading)
		if err != sql.ErrNoRows && err != nil {
			return -1, -1, err
		}
	}

	day_ago_reading := mathsfn.GetMode(day_ago_readings)

	percentage = calculatePercentFilled(current_reading) - calculatePercentFilled(day_ago_reading)

	//Round to 2dp
	distance = math.Round((three_hour_ago_reading-current_reading)*100) / 100
	return percentage, distance, nil
}

func CalculateGraphData(db *sql.DB, start_date time.Time, end_date time.Time) (*templates.TemplateData, error) {
	data, times, time_interval, err := getReadings(db, start_date, end_date)
	if err != nil {
		return nil, err
	}

	templateData := new(templates.TemplateData)
	templateData.Graph_data = data
	templateData.Graph_times = times

	if end_date.Sub(start_date).Hours() > 48 {
		templateData.Time_unit = "Days"

		for i := range data {
			templateData.Graph_labels = append(templateData.Graph_labels, float64(i)*time_interval.Minutes()/60/24)
		}
	} else {
		templateData.Time_unit = "Hours"

		for i := range data {
			templateData.Graph_labels = append(templateData.Graph_labels, float64(i)*time_interval.Minutes()/60)
		}
	}

	return templateData, nil
}

func CalculateOtherData(db *sql.DB) (*templates.TemplateData, error) {
	heights, err := getLastReadings(db, 10)
	if err != nil {
		return nil, err
	}

	height := mathsfn.GetMode(heights)
	percentage := calculatePercentFilled(height)

	templateData := new(templates.TemplateData)

	templateData.Height = math.Round((getMaxHeight()-height)*100) / 100
	templateData.Percentage = percentage

	percent_change, distance_change, err := calculateChanges(db)
	if err != nil {
		return nil, err
	}

	templateData.Percentage_changed = percent_change

	if percent_change == 0 {
		templateData.Percentage_changed_sign = ""
	} else if percent_change > 0 {
		templateData.Percentage_changed_sign = "+"
	} else {
		templateData.Percentage_changed *= -1
		templateData.Percentage_changed_sign = "-"
	}

	templateData.Distance_changed = distance_change

	if distance_change == 0 {
		templateData.Distance_changed_sign = ""
	} else if distance_change > 0 {
		templateData.Distance_changed_sign = "+"
	} else {
		templateData.Distance_changed *= -1
		templateData.Distance_changed_sign = "-"
	}

	return templateData, nil
}

func getDateOfOldestRecord(db *sql.DB) (*time.Time, error) {
	statement, err := db.Prepare("SELECT time FROM readings ORDER BY time ASC LIMIT 1")
	if err != nil {
		return nil, err
	}

	date := ""
	row := statement.QueryRow()

	err = row.Scan(&date)
	if err != nil {
		return nil, err
	}

	date_str := strings.Split(date, " ")[0]
	created_date, err := time.Parse("2006-01-02", date_str)
	if err != nil {
		return nil, err
	}

	return &created_date, nil
}

func CalculateAllTemplateData(db *sql.DB, start_date time.Time, end_date time.Time) (*templates.TemplateData, error) {
	templateData, err := CalculateOtherData(db)
	if err != nil {
		return nil, err
	}

	graphData, err := CalculateGraphData(db, start_date, end_date)
	if err != nil {
		return nil, err
	}

	templateData.Graph_data = graphData.Graph_data
	templateData.Graph_labels = graphData.Graph_labels
	templateData.Graph_times = graphData.Graph_times

	templateData.Start_date = start_date.Format("2006-01-02")
	templateData.End_date = end_date.Format("2006-01-02")

	templateData.Start_date_max = templateData.Start_date
	templateData.End_date_max = templateData.End_date

	oldest_date, err := getDateOfOldestRecord(db)
	if err != nil {
		return nil, err
	}

	templateData.Start_date_min = oldest_date.Format("2006-01-02")
	templateData.End_date_min = oldest_date.Add(time.Hour * -24).Format("2006-01-02")

	templateData.Time_unit = "Hours"

	return templateData, nil
}

func getMaxHeight() float64 {
	return 0.52
}

func calculatePercentFilled(height float64) int {
	if height < 0 {
		height = 0
	}

	max_height := getMaxHeight()
	if height > max_height {
		height = max_height
	}

	percentage := int(math.Round((-(height - max_height) / max_height) * 100))
	return percentage
}

func getReadings(db *sql.DB, start_date time.Time, end_date time.Time) (readings []int, associated_times []string, interval time.Duration, err error) {
	days := int(math.Round((end_date.Sub(start_date).Hours() / 24)))

	statement, err := db.Prepare("SELECT COUNT(ID) FROM readings WHERE time >= ? AND time <= ?")
	if err != nil {
		return nil, nil, 0, err
	}

	row := statement.QueryRow(mathsfn.GetDateTime(start_date), mathsfn.GetDateTime(end_date))
	if row == nil {
		return nil, nil, 0, errors.New("row count query failed")
	}

	row_count := 0
	row.Scan(&row_count)

	time_interval := time.Minute * 20 * time.Duration(days)
	readings_per_interval := 20

	total_readings_wanted := (days * 24 * 60) / int(time_interval.Minutes()) * readings_per_interval

	modulus := row_count / total_readings_wanted
	if modulus < 1 {
		modulus = 1
	}

	statement, err = db.Prepare("SELECT t.time, t.height FROM (SELECT height, time, ROW_NUMBER() OVER (ORDER BY time) AS rownumber FROM readings) AS t WHERE t.rownumber % ? = 0 AND t.time >= ? AND t.time <= ? ORDER BY t.time")
	if err != nil {
		return nil, nil, 0, err
	}

	var heights []float64
	var times []time.Time

	height := 0.0
	db_time := ""

	rows, err := statement.Query(modulus, mathsfn.GetDateTime(start_date), mathsfn.GetDateTime(end_date))
	if err != nil {
		return nil, nil, 0, err
	}

	for rows.Next() {
		rows.Scan(&db_time, &height)

		parsed_time, err := mathsfn.GetTime(db_time)
		if err != nil {
			return nil, nil, 0, err
		}

		heights = append(heights, height)
		times = append(times, parsed_time)
	}

	index := 0

	for start_date.Unix() <= end_date.Unix() {
		average := 0.0
		average_count := 0

		for index < len(heights) && start_date.Add(time_interval).Unix() > times[index].Unix() {
			average += heights[index]
			average_count += 1
			index += 1
		}

		if average_count == 0 {
			readings = append(readings, -1)
			associated_times = append(associated_times, "")
		} else {
			average /= float64(average_count)
			readings = append(readings, calculatePercentFilled(average))

			current_dt := mathsfn.GetDateTime(start_date.Add(time_interval / 2))
			associated_times = append(associated_times, current_dt+"UTC")
		}

		start_date = start_date.Add(time_interval)
	}

	//Reverse Readings
	for i, j := 0, len(readings)-1; i < j; i, j = i+1, j-1 {
		readings[i], readings[j] = readings[j], readings[i]
		associated_times[i], associated_times[j] = associated_times[j], associated_times[i]
	}

	return readings, associated_times, time_interval, nil
}
