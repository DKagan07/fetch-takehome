## Submission for the Receipt Processor Challenge

### Running this application

To run this application, please run `go mod tidy` and `go run main.go
receiptHandler.go`. This will run the server on port 8085.

As described in the brief, the 2 endpoints that are available are:

1. POST to `/receipts/process` with a payload for a receipt following the JSON
   structure of PostReceiptJSON in `receiptHandler.go`

This will return a JSON structure of the generated ID of that receipt.
This ID will be used for the other endpoint to retreive the number of points of
that receipt

2. GET to `/receipts/{id}/points`

This will return the number of points of that receipt following the rules listed
here:

- One point for every alphanumeric character in the retailer name.
- 50 points if the total is a round dollar amount with no cents.
- 25 points if the total is a multiple of 0.25.
- 5 points for every two items on the receipt.
- If the trimmed length of the item description is a multiple of 3, multiply the
  price by 0.2 and round up to the nearest integer. The result is the number
  of points earned.
- 6 points if the day in the purchase date is odd.
- 10 points if the time of purchase is after 2:00pm and before 4:00pm.

##### Testing

I have added 1 small unit test just to show that testing can be done. To run the
test, please run `go test -v *.go`.
