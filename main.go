package main

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"streaming-project/database"
	"streaming-project/streaming"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln("No .env file found!")
	}
}

func main() {
	dbUser, exists := os.LookupEnv("POSTGRESQL_USER")
	if !exists {
		log.Fatalln("Username for Postgresql not found in .env file!")
	}

	dbPassword, exists := os.LookupEnv("POSTGRESQL_PASSWORD")
	if !exists {
		log.Fatalln("Password for Postgresql not found in .env file!")
	}

	dbPort, exists := os.LookupEnv("POSTGRESQL_PORT")
	if !exists {
		log.Fatalln("Port for Postgresql not found in .env file!")
	}
	dbPortInt, err := strconv.Atoi(dbPort)
	if err != nil {
		log.Fatalln(err)
	}

	dbName, exists := os.LookupEnv("POSTGRESQL_DBNAME")
	if !exists {
		log.Fatalln("Database name for Postgresql not found in .env file!")
	}

	db, err := database.NewDatabase(dbUser, dbPassword, dbName, dbPortInt)
	if err != nil {
		log.Fatalln(err)
	}

	err = db.Init()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Connected to PostgreSQl")

	err = streaming.StartConsumer(db)
	if err != nil {
		log.Fatalln(err)
	}

}
