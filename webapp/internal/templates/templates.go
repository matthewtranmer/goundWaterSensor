package templates

type TemplateData struct {
	Height                  float64
	Percentage              int
	Percentage_changed_sign string
	Percentage_changed      int
	Distance_changed_sign   string
	Distance_changed        float64
	Start_date              string
	End_date                string
	Graph_labels            []float64
	Graph_data              []int
	Graph_times             []string
	Time_unit               string
	Start_date_min          string
	Start_date_max          string
	End_date_min            string
	End_date_max            string
}
