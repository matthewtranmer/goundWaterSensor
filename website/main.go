package main

import (
	"os"
	"strings"
	"website/internal/endpoints"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	data, err := os.ReadFile("dbpassword")
	if err != nil {
		panic(err)
	}

	dbpassword := string(data)
	dbpassword = strings.TrimSuffix(dbpassword, "\n")

	ep := new(endpoints.Endpoints)
	ep.StartServer("127.0.0.1:3000", "127.0.0.1:3306", dbpassword)
}
