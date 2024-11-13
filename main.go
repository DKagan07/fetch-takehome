package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const PORT = 8085

func main() {
	rh := NewReceiptHandler()

	r := chi.NewRouter()

	r.Post("/receipts/process", rh.handleProcessReceipts)
	r.Get("/receipts/{id}/points", rh.handleGetReceiptPoints)

	fmt.Printf("Starting server on PORT: %d\n", PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", PORT), r))
}
