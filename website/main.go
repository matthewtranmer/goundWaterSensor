package main

import (
	"website/internal/endpoints"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	ep := new(endpoints.Endpoints)
	ep.StartServer()
}
