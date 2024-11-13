package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// PostReceiptJSON is what is expected to be the payload for the POST request.
// In the example structures in the repo, they are all strings, so we parse them
// as strings here.
//
// Some notes regarding the payload structure:
// 1. The date follows the "YYYY-MM-DD" structure
// 2. The time follows the "HH:MM" structure, using the 24 hr clock
//
// Note: the schema doesn't note any required fields, so I have all of them as
// 'omitempty'. This would be a straightforward change with more information
type PostReceiptJSON struct {
	Retailer     string `json:"retailer,omitempty"`
	PurchaseDate string `json:"purchaseDate,omitempty"`
	PurchaseTime string `json:"purchaseTime,omitempty"`
	Total        string `json:"total,omitempty"`
	Items        []Item `json:"items,omitempty"`
}

// Item is a singular item that appears in a slice in an instance of a
// PostReceiptJSON
type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

// StorageReceipt is a representation of what a receipt is in storage
type StorageReceipt struct {
	PostReceiptJSON
	Id uuid.UUID `json:"id"`
}

// ReceiptHandler is the structure being used to contain all the handlers for
// handling receipts, as well as maintaining an in-memory storage of the
// receipts, as part of the prompt
type ReceiptHandler struct {
	mu       sync.Mutex
	Receipts []StorageReceipt
}

func NewReceiptHandler() *ReceiptHandler {
	return &ReceiptHandler{}
}

// handleProcessReceipts handles a POST request and stores the receipt in
// storage
func (rh *ReceiptHandler) handleProcessReceipts(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var postReceipt PostReceiptJSON
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&postReceipt); err != nil {
		fmt.Printf("decoding post receipt: %+v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		if _, err = w.Write([]byte("invalid receipt")); err != nil {
			fmt.Println("writing to response: ", err)
			return
		}
		return
	}

	newUUID, err := uuid.NewV7()
	if err != nil {
		fmt.Printf("creating new UUID: %+v\n", err)
		return
	}

	storedData := StorageReceipt{
		Id:              newUUID,
		PostReceiptJSON: postReceipt,
	}

	rh.mu.Lock()
	defer rh.mu.Unlock()
	rh.Receipts = append(rh.Receipts, storedData)

	// Note: These anonymous structs are used just for this prompt for
	// simplicity
	jsonResponse := struct {
		Id string `json:"id"`
	}{
		Id: storedData.Id.String(),
	}

	b, err := json.Marshal(jsonResponse)
	if err != nil {
		fmt.Printf("marshaling json: %+v\n", err)
		return
	}

	if _, err = w.Write(b); err != nil {
		fmt.Printf("writing to response: %+v\n", err)
		return
	}
}

// handleGetReceiptPoints is a handler that returns the points value for a
// given receipt UUID
func (rh *ReceiptHandler) handleGetReceiptPoints(w http.ResponseWriter, r *http.Request) {
	idFromURL := chi.URLParam(r, "id")

	rh.mu.Lock()
	defer rh.mu.Unlock()
	for _, receipt := range rh.Receipts {
		if idFromURL == receipt.Id.String() {
			points, err := findPoints(receipt)
			if err != nil {
				fmt.Println("error with calculating points: ", err)
			}
			// Note: These anonymous structs are used just for this prompt for
			// simplicity
			responsePoints := struct {
				Points int `json:"points"`
			}{
				Points: points,
			}

			b, err := json.Marshal(responsePoints)
			if err != nil {
				fmt.Printf("marshaling json: %+v\n", err)
				return
			}

			if _, err = w.Write(b); err != nil {
				fmt.Printf("writing to response: %+v\n", err)
				return
			}
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	if _, err := w.Write([]byte("Id not found")); err != nil {
		fmt.Printf("writing to response: %+v\n", err)
		return
	}
}

// findPoints is a helper function to count up how many points a receipt is
// worth, following these given rules:
// 1. One point for every alphanumeric character in the retailer name.
//
// 2. 50 points if the total is a round dollar amount with no cents.
//
// 3. 25 points if the total is a multiple of 0.25.
//
// 4. 5 points for every two items on the receipt.
//
// 5. If the trimmed length of the item description is a multiple of 3, multiply
// the price by 0.2 and round up to the nearest integer. The result is the
// number of points earned.
//
// 6. 6 points if the day in the purchase date is odd.
//
// 7. 10 points if the time of purchase is after 2:00pm and before 4:00pm
//
// Note: depending on feedback, each of these rules can be turned into a small,
// separate function
func findPoints(re StorageReceipt) (int, error) {
	total := 0

	// #1
	count := 0
	for _, c := range re.Retailer {
		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			count++
		}
	}
	total += count

	totalPrice, err := strconv.ParseFloat(re.Total, 64)
	if err != nil {
		return 0, err
	}

	// #2
	if totalPrice == math.Floor(totalPrice) {
		total += 50
	}

	// #3
	if math.Mod(totalPrice, 0.25) == 0 {
		total += 25
	}

	// #4
	numItems := len(re.Items)
	m := math.Floor(float64(numItems) / 2)
	total += int(m * 5)

	// #5
	for _, item := range re.Items {
		trimmedDesc := strings.TrimSpace(item.ShortDescription)
		if len(trimmedDesc)%3 == 0 {
			itemPrice, err := strconv.ParseFloat(item.Price, 64)
			if err != nil {
				return 0, err
			}

			earnedPts := math.Ceil(itemPrice * 0.2)
			total += int(earnedPts)
		}
	}

	// #6
	splitStrs := strings.Split(re.PurchaseDate, "-")
	if len(splitStrs) != 3 {
		return 0, fmt.Errorf("invalid date. Must follow the YYYY-MM-DD scheme")
	}
	date := splitStrs[2]
	dateNum, err := strconv.Atoi(date)
	if err != nil {
		return 0, err
	}

	if dateNum%2 == 1 {
		total += 6
	}

	// #7
	timeSplit := strings.Split(re.PurchaseTime, ":")
	hrs, err := strconv.Atoi(timeSplit[0])
	if err != nil {
		return 0, err
	}
	mins, err := strconv.Atoi(timeSplit[1])
	if err != nil {
		return 0, err
	}

	if hrs == 14 && mins == 0 {
		fmt.Println("do nothing, because it's exactly 2pm")
	} else if hrs >= 14 && hrs < 16 {
		total += 10
	}

	return total, nil
}
