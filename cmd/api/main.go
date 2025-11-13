package main

import (
	"log"
	"net/http"
	"os"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/wire"
)

func main() {
	router := wire.InitializeInvoiceAPI()
	

	port := os.Getenv("APP_PORT")

	log.Printf("API listening on port %s\n", port)
    log.Fatal(http.ListenAndServe(":"+port, router))
}
