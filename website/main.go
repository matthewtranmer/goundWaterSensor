package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Handler struct {
	Middleware func(w http.ResponseWriter, r *http.Request)
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Middleware(w, r)
}

type Pages struct {
	db *sql.DB
}

func getMode(numbers []float64) float64 {
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

type TemplateData struct {
	Height                  float64
	Percentage              int
	Percentage_changed_sign string
	Percentage_changed      int
	Distance_changed_sign   string
	Distance_changed        float64
	Start_date              string
	End_date                string
	Graph_labels            []int
	Graph_data              []int
}

func getDateTime(datetime time.Time) string {
	return fmt.Sprintf("%d-%d-%d %d:%d:%d\n",
		datetime.Year(),
		datetime.Month(),
		datetime.Day(),
		datetime.Hour(),
		datetime.Minute(),
		datetime.Second(),
	)
}

func getTime(datetime string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", datetime)
}

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
		index++
	}

	return heights, nil
}

func calculateChanges(db *sql.DB) (percentage int, distance float64, err error) {
	day_ago := time.Now().Add(-24 * time.Hour)
	hour_ago := time.Now().Add(-1 * time.Hour)

	statement, err := db.Prepare("SELECT height FROM readings ORDER BY time DESC LIMIT 1")
	if err != nil {
		return -1, -1, err
	}

	row := statement.QueryRow()
	if err != nil {
		return -1, -1, err
	}

	current_reading := 0.0
	err = row.Scan(&current_reading)
	if err != sql.ErrNoRows && err != nil {
		return -1, -1, err
	}

	statement, err = db.Prepare("SELECT height FROM readings WHERE time <= ? ORDER BY time DESC LIMIT 1")
	if err != nil {
		return -1, -1, err
	}

	row = statement.QueryRow(getDateTime(hour_ago))
	if err != nil {
		return -1, -1, err
	}

	hour_ago_reading := 0.0
	err = row.Scan(&hour_ago_reading)
	if err != sql.ErrNoRows && err != nil {
		return -1, -1, err
	}

	statement, err = db.Prepare("SELECT height FROM readings WHERE time <= ? ORDER BY time DESC LIMIT 1")
	if err != nil {
		return -1, -1, err
	}

	row = statement.QueryRow(getDateTime(day_ago))
	if err != nil {
		return -1, -1, err
	}

	day_ago_reading := 0.0
	err = row.Scan(&day_ago_reading)
	if err != sql.ErrNoRows && err != nil {
		return -1, -1, err
	}

	if day_ago_reading != 0.0 {
		percentage = int(math.Round((current_reading - day_ago_reading) / day_ago_reading * 100))
	} else {
		percentage = int(math.Round((current_reading + 1 - day_ago_reading + 1) / (day_ago_reading + 1) * 100))
	}

	distance = current_reading - hour_ago_reading
	return percentage, distance, nil
}

func calculateGraphData(db *sql.DB, start_date time.Time, end_date time.Time) (*TemplateData, error) {
	time_interval := time.Minute * 15

	data, err := getReadings(db, start_date, end_date, time_interval)
	if err != nil {
		return nil, err
	}

	templateData := new(TemplateData)

	templateData.Graph_data = data

	for i := range data {
		templateData.Graph_labels = append(templateData.Graph_labels, i*int(time_interval.Minutes()))
	}

	return templateData, nil
}

func calculateOtherData(db *sql.DB) (*TemplateData, error) {
	heights, err := getLastReadings(db, 10)
	if err != nil {
		return nil, err
	}

	height := getMode(heights)
	percentage := calculatePercentFilled(height)

	templateData := new(TemplateData)

	templateData.Height = height
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

func calculateAllTemplateData(db *sql.DB, start_date time.Time, end_date time.Time) (*TemplateData, error) {
	templateData, err := calculateOtherData(db)
	if err != nil {
		return nil, err
	}

	graphData, err := calculateGraphData(db, start_date, end_date)
	if err != nil {
		return nil, err
	}

	templateData.Graph_data = graphData.Graph_data
	templateData.Graph_labels = graphData.Graph_labels

	templateData.Start_date = start_date.Format("2006-01-02")
	templateData.End_date = end_date.Format("2006-01-02")

	return templateData, nil
}

func (p *Pages) getNewData(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get New Data Request")

	templateData, err := calculateOtherData(p.db)
	if err != nil {
		fmt.Println(err)
		return
	}

	json_bytes, err := json.Marshal(templateData)
	if err != nil {
		fmt.Println(err)
		return
	}

	json_str := string(json_bytes)
	fmt.Fprint(w, json_str)
}

func (p *Pages) home(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Home Request")

	//start_date, _ := getTime("2023-08-13 13:27:26")
	//end_date, _ := getTime("2023-08-14 13:27:26")

	//end_date := time.Date(2023, 8, 14, 13, 29, 17, 763533949, time.Local)

	end_date := time.Now().UTC()
	start_date := end_date.Add(-24 * time.Hour)

	TemplateData, err := calculateAllTemplateData(p.db, start_date, end_date)
	if err != nil {
		fmt.Println(err)
		return
	}

	doc, err := template.ParseFiles("templates/home.html")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = doc.Execute(w, TemplateData)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func calculatePercentFilled(height float64) int {
	const max_height = 0.52
	if height > max_height {
		height = max_height
	}

	percentage := int(math.Round((-(height - max_height) / max_height) * 100))
	return percentage
}

func getReadings(db *sql.DB, start_date time.Time, end_date time.Time, interval time.Duration) (readings []int, err error) {
	statement, err := db.Prepare("SELECT height, time FROM readings WHERE time >= ? AND time <= ? ORDER BY time")
	if err != nil {
		return nil, err
	}

	var heights []float64
	var times []time.Time

	height := 0.0
	time := ""

	rows, err := statement.Query(getDateTime(start_date), getDateTime(end_date))
	if err != nil {
		return nil, err
	}

	DEBUG := 0

	for rows.Next() {
		DEBUG += 1
		rows.Scan(&height, &time)

		parsed_time, err := getTime(time)
		if err != nil {
			return nil, err
		}

		heights = append(heights, height)
		times = append(times, parsed_time)
	}

	index := 0

	for start_date.Unix() <= end_date.Unix() {
		average := 0.0
		average_count := 0

		for index < len(heights) && start_date.Add(interval).Unix() > times[index].Unix() {
			average += heights[index]
			average_count += 1
			index += 1
		}

		if average_count == 0 {
			readings = append(readings, -1)
		} else {
			average /= float64(average_count)
			readings = append(readings, calculatePercentFilled(average))
		}

		start_date = start_date.Add(interval)
	}

	return readings, nil
}

func (p *Pages) getNewGraph(w http.ResponseWriter, r *http.Request) {
	start_date, err := getTime(r.URL.Query()["startdate"][0] + " 00:00:01")
	if err != nil {
		fmt.Println(err)
		return
	}

	end_date, err := getTime(r.URL.Query()["enddate"][0] + " 23:59:59")
	if err != nil {
		fmt.Println(err)
		return
	}

	if start_date.Compare(end_date) > 0 {
		return
	}

	templateData, err := calculateGraphData(p.db, start_date, end_date)
	if err != nil {
		fmt.Println(err)
		return
	}

	json_bytes, err := json.Marshal(templateData)
	if err != nil {
		fmt.Println(err)
		return
	}

	json_str := string(json_bytes)
	fmt.Fprint(w, json_str)
}

func main() {
	//testng only
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static", fs))

	pages := new(Pages)

	data, err := os.ReadFile("dbpassword")
	if err != nil {
		panic(err)
	}

	dbpassword := string(data)
	dbpassword = strings.TrimSuffix(dbpassword, "\n")

	pages.db, err = sql.Open("mysql", "WorkerRW:"+dbpassword+"@tcp(127.0.0.1:3306)/sensor")
	if err != nil {
		panic(err)
	}

	http.Handle("/", Handler{Middleware: pages.home})
	http.Handle("/api/getNewData", Handler{Middleware: pages.getNewData})
	http.Handle("/api/getNewGraph", Handler{Middleware: pages.getNewGraph})

	http.ListenAndServe("127.0.0.1:3000", nil)

}
