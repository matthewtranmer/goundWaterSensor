package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"website/internal/endpoints"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	fmt.Println("Starting Up")

	db_file, exists := os.LookupEnv("DATABASE_SECRET_FILE")
	if !exists {
		log.Fatalln("Database secret not given")
		panic("Database secret not given")
	}

	data, err := os.ReadFile(db_file)
	if err != nil {
		log.Fatalf("Could not read DB file: %s\n", err)
		panic(err)
	}

	dbpassword := string(data)
	dbpassword = strings.TrimSuffix(dbpassword, "\n")

	ep := new(endpoints.Endpoints)
	ep.StartServer("0.0.0.0:3000", "mysql:3306", dbpassword)
}
