package main

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPoint(t *testing.T) {
	uuid1, err := uuid.NewV7()
	assert.NoError(t, err)

	uuid2, err := uuid.NewV7()
	assert.NoError(t, err)

	tests := []StorageReceipt{
		{
			Id: uuid1,
			PostReceiptJSON: PostReceiptJSON{
				Retailer:     "Target",
				PurchaseDate: "2022-01-01",
				PurchaseTime: "13:01",
				Items: []Item{
					{
						ShortDescription: "Mountain Dew 12PK",
						Price:            "6.49",
					}, {
						ShortDescription: "Emils Cheese Pizza",
						Price:            "12.25",
					}, {
						ShortDescription: "Knorr Creamy Chicken",
						Price:            "1.26",
					}, {
						ShortDescription: "Doritos Nacho Cheese",
						Price:            "3.35",
					}, {
						ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ",
						Price:            "12.00",
					},
				},
				Total: "35.35",
			},
		},
		{
			Id: uuid2,
			PostReceiptJSON: PostReceiptJSON{
				Retailer:     "M&M Corner Market",
				PurchaseDate: "2022-03-20",
				PurchaseTime: "14:33",
				Items: []Item{
					{
						ShortDescription: "Gatorade",
						Price:            "2.25",
					}, {
						ShortDescription: "Gatorade",
						Price:            "2.25",
					}, {
						ShortDescription: "Gatorade",
						Price:            "2.25",
					}, {
						ShortDescription: "Gatorade",
						Price:            "2.25",
					},
				},
				Total: "9.00",
			},
		},
	}

	exp := []struct {
		Points int
	}{
		{
			Points: 28,
		},
		{
			Points: 109,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("TestingReceipt%d", i+1), func(t *testing.T) {
			pts, err := findPoints(test)
			assert.NoError(t, err)
			assert.Equal(t, pts, exp[i].Points)
		})
	}
}
