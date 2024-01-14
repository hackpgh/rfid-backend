/*
RFID Backend Server
===================
This Go server provides backend services for an RFID-based access control system
for HackPGH's magnetic locks and machines that require Safety Training sign-offs.
It interacts with the Wild Apricot API to fetch contact data and manages
a local SQLite database to store and process RFID tags and training information.

Project Structure:
- /config: Configuration file loading logic.
- /db: Database initialization and schema management.
- /db/schema: Database schema files.
- /handlers: HTTP handlers for different server endpoints.
- /models: Data structures representing database entities and API responses.
- /services: Business logic, including interaction with external APIs
             and database operations; also contains queries.go

Main Functionality:
- Initializes the SQLite database using the specified database path from `config.yml`.
- Sets up the Wild Apricot service for API interactions, enabling the retrieval of contact data.
- Creates a DBService instance for handling database operations.
- Initializes a CacheHandler with the DBService and configuration settings to handle HTTP requests.
- Registers HTTP endpoints `/api/machineCache` and `/api/doorCache` for fetching RFID data
  related to machines and door access.
- Starts a background routine that periodically fetches contact data from the Wild Apricot
  API and updates the local SQLite database. This ensures the database is regularly
  synchronized with the latest data from Wild Apricot.
- Launches an HTTPS server on port 443 to listen for incoming requests, using the SSL
  certificate and key specified in the `config.yml`.

Usage:
- Before running, ensure that the `config.yml` is properly set up with the necessary configuration, including database path, Wild Apricot account ID, SSL certificate, and key file locations.
- Run the server to start listening for HTTP requests on port 443 and to keep the local database synchronized with the Wild Apricot API data.
*/

package main

import (
	"log"
	"net/http"
	"path/filepath"
	"rfid-backend/config"
	"rfid-backend/db"
	"rfid-backend/handlers"
	"rfid-backend/services"
	"runtime"
	"time"
)

func getCurrentDirectory() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Dir(b)
}

func main() {
	cfg := config.LoadConfig()

	database, err := db.InitDB(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	wildApricotSvc := services.NewWildApricotService(database)
	dbService := services.NewDBService(database, cfg)

	cacheHandler := handlers.NewCacheHandler(dbService, cfg)

	http.HandleFunc("/api/machineCache", cacheHandler.HandleMachineCacheRequest())
	http.HandleFunc("/api/doorCache", cacheHandler.HandleDoorCacheRequest())

	// Start background task to fetch contacts and update the database
	go func() {
		ticker := time.NewTicker(6 * time.Minute)
		for range ticker.C {
			updateDatabaseFromWildApricot(wildApricotSvc, dbService, cfg.WildApricotAccountId)
		}
	}()

	log.Println("Starting HTTPS server on :443...")
	log.Printf("certfile: %+v\n", cfg.CertFile)
	log.Printf("keyfile: %+v\n", cfg.KeyFile)
	err = http.ListenAndServeTLS(":443", cfg.CertFile, cfg.KeyFile, nil)
	if err != nil {
		log.Fatalf("Failed to start HTTPS server: %v", err)
	}
}

func updateDatabaseFromWildApricot(waService *services.WildApricotService, dbService *services.DBService, accountId int) {
	log.Println("Fetching contacts from Wild Apricot and updating database...")
	contacts, err := waService.GetContacts(accountId)
	if err != nil {
		log.Printf("Failed to fetch contacts: %v", err)
		return
	}

	if err = dbService.ProcessContactsData(contacts); err != nil {
		log.Printf("Failed to update database: %v", err)
		return
	}

	log.Println("Database successfully updated with latest WA contact data.")
}
