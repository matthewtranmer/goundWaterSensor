package endpoints

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
	"website/internal/dataproc"
	"website/internal/mathsfn"
)

type Handler struct {
	Middleware func(w http.ResponseWriter, r *http.Request)
}

type Endpoints struct {
	db *sql.DB
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Middleware(w, r)
}

func (e *Endpoints) getNewData(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get New Data Request")

	templateData, err := dataproc.CalculateOtherData(e.db)
	if err != nil {
		log.Println(err)
		return
	}

	json_bytes, err := json.Marshal(templateData)
	if err != nil {
		log.Println(err)
		return
	}

	json_str := string(json_bytes)
	fmt.Fprint(w, json_str)
}

func (e *Endpoints) getNewGraph(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get New Graph")
	current_time := time.Now().UTC().Format("15:04:05")

	start_date, err := mathsfn.GetTime(r.URL.Query()["startdate"][0] + " " + current_time)
	if err != nil {
		log.Println(err)
		return
	}

	end_date, err := mathsfn.GetTime(r.URL.Query()["enddate"][0] + " " + current_time)
	if err != nil {
		log.Println(err)
		return
	}

	if start_date.Compare(end_date) > 0 {
		return
	}

	templateData, err := dataproc.CalculateGraphData(e.db, start_date, end_date)
	if err != nil {
		log.Println(err)
		return
	}

	json_bytes, err := json.Marshal(templateData)
	if err != nil {
		log.Println(err)
		return
	}

	json_str := string(json_bytes)
	fmt.Fprint(w, json_str)
}

func (e *Endpoints) home(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Home Request")

	end_date := time.Now().UTC()
	start_date := end_date.Add(-24 * time.Hour)

	TemplateData, err := dataproc.CalculateAllTemplateData(e.db, start_date, end_date)
	if err != nil {
		log.Println(err)
		return
	}

	doc, err := template.ParseFiles("templates/home.html")
	if err != nil {
		log.Println(err)
		return
	}

	err = doc.Execute(w, TemplateData)
	if err != nil {
		log.Println(err)
		return
	}
}

func (e *Endpoints) StartServer(address string, dbaddress string, dbpassword string) {
	var err error
	e.db, err = sql.Open("mysql", "WorkerRW:"+dbpassword+"@tcp("+dbaddress+")/sensor")
	if err != nil {
		panic(err)
	}

	http.Handle("/sensor/", Handler{Middleware: e.home})
	http.Handle("/sensor/api/getNewData", Handler{Middleware: e.getNewData})
	http.Handle("/sensor/api/getNewGraph", Handler{Middleware: e.getNewGraph})

	http.ListenAndServe(address, nil)
}
