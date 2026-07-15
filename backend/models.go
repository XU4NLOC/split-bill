package main

// ---------- Request / Response models ----------

type Person struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	IsPayer bool   `json:"isPayer"`
}

type Assignment struct {
	PersonID string `json:"personId"`
	Quantity int    `json:"quantity"`
}

type Item struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Quantity   int          `json:"quantity"`
	TotalPrice float64      `json:"totalPrice"`
	Assignments []Assignment `json:"assignments"`
}

type SplitRequest struct {
	People []Person `json:"people"`
	Items  []Item   `json:"items"`
}

type PersonResult struct {
	PersonID        string  `json:"personId"`
	Name            string  `json:"name"`
	IsPayer         bool    `json:"isPayer"`
	Subtotal        float64 `json:"subtotal"`
	RoundedSubtotal float64 `json:"roundedSubtotal"`
	ItemCount       int     `json:"itemCount"`
	Owes            float64 `json:"owes"`
	RoundedOwes     float64 `json:"roundedOwes"`
}

type Settlement struct {
	FromID   string  `json:"fromId"`
	FromName string  `json:"fromName"`
	ToID     string  `json:"toId"`
	ToName   string  `json:"toName"`
	Amount   float64 `json:"amount"`
}

type SplitResponse struct {
	Total        float64        `json:"total"`
	RoundedTotal float64        `json:"roundedTotal"`
	PerPerson    []PersonResult `json:"perPerson"`
	Settlements  []Settlement   `json:"settlements"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// ---------- Error helper ----------

type simpleError string

func (e simpleError) Error() string { return string(e) }

func errString(s string) error { return simpleError(s) }
