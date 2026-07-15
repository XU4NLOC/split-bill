package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
)

// ---------- Data models ----------

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
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Quantity    int          `json:"quantity"`
	TotalPrice  float64      `json:"totalPrice"`
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

// ---------- Handlers ----------

func splitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "only POST is supported")
		return
	}

	var req SplitRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	resp, err := calculateSplit(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// calculateSplit is pure business logic, kept separate from HTTP for testability.
func calculateSplit(req SplitRequest) (*SplitResponse, error) {
	if len(req.People) == 0 {
		return nil, errString("at least one person is required")
	}

	peopleByID := map[string]Person{}
	var payerID string
	payerCount := 0
	for _, p := range req.People {
		name := strings.TrimSpace(p.Name)
		if name == "" {
			return nil, errString("every person must have a name")
		}
		peopleByID[p.ID] = p
		if p.IsPayer {
			payerCount++
			payerID = p.ID
		}
	}
	if payerCount == 0 {
		return nil, errString("exactly one person must be marked as the payer")
	}
	if payerCount > 1 {
		return nil, errString("only one person can be marked as the payer")
	}

	costs := map[string]float64{}
	counts := map[string]int{}
	var total float64

	for _, item := range req.Items {
		if strings.TrimSpace(item.Name) == "" {
			return nil, errString("every item must have a name")
		}
		if item.Quantity <= 0 {
			return nil, errString(fmt.Sprintf("item '%s': quantity must be at least 1", item.Name))
		}
		if item.TotalPrice < 0 {
			return nil, errString(fmt.Sprintf("item '%s': total price cannot be negative", item.Name))
		}

		assignedTotal := 0
		for _, a := range item.Assignments {
			if a.Quantity < 0 {
				return nil, errString(fmt.Sprintf("item '%s': assignment quantity cannot be negative", item.Name))
			}
			if _, ok := peopleByID[a.PersonID]; !ok {
				return nil, errString(fmt.Sprintf("item '%s': assignment references unknown person", item.Name))
			}
			assignedTotal += a.Quantity
			costs[a.PersonID] += float64(a.Quantity) * (item.TotalPrice / float64(item.Quantity))
			counts[a.PersonID]++
		}

		if assignedTotal != item.Quantity {
			return nil, errString(fmt.Sprintf("item '%s': assigned quantity (%d) does not match total quantity (%d)", item.Name, assignedTotal, item.Quantity))
		}

		total += item.TotalPrice
	}

	perPerson := make([]PersonResult, 0, len(req.People))
	settlements := make([]Settlement, 0, len(req.People)-1)
	payer := peopleByID[payerID]

	// 1. Payer always rounds DOWN (floors to 1000)
	payerExact := costs[payerID]
	payerAmount := floorTo1000(payerExact)

	// 2. Leftover is what non-payers must cover
	leftover := total - payerAmount

	// 3. Collect non-payers and split leftover equally
	type npInfo struct {
		person Person
		exact  float64
	}
	nonPayers := make([]npInfo, 0, len(req.People)-1)
	for _, p := range req.People {
		if !p.IsPayer {
			nonPayers = append(nonPayers, npInfo{person: p, exact: costs[p.ID]})
		}
	}

	roundedAmounts := map[string]float64{}
	if len(nonPayers) > 0 {
		base := math.Floor(leftover/float64(len(nonPayers))/1000) * 1000
		remainder := leftover - base*float64(len(nonPayers))
		extraCount := int(remainder / 1000)
		for i := len(nonPayers) - 1; i >= 0 && extraCount > 0; i-- {
			roundedAmounts[nonPayers[i].person.ID] = base + 1000
			extraCount--
		}
		for i := 0; i < len(nonPayers); i++ {
			if _, ok := roundedAmounts[nonPayers[i].person.ID]; !ok {
				roundedAmounts[nonPayers[i].person.ID] = base
			}
		}
	}

	// 4. Build results
	for _, p := range req.People {
		subtotal := costs[p.ID]
		owes := 0.0
		roundedOwes := 0.0
		if p.IsPayer {
			roundedOwes = payerAmount
		} else {
			owes = subtotal
			roundedOwes = roundedAmounts[p.ID]
		}
		perPerson = append(perPerson, PersonResult{
			PersonID:        p.ID,
			Name:            p.Name,
			IsPayer:         p.IsPayer,
			Subtotal:        subtotal,
			RoundedSubtotal: roundedOwes,
			ItemCount:       counts[p.ID],
			Owes:            owes,
			RoundedOwes:     roundedOwes,
		})
		if !p.IsPayer && roundedOwes > 0 {
			settlements = append(settlements, Settlement{
				FromID:   p.ID,
				FromName: p.Name,
				ToID:     payer.ID,
				ToName:   payer.Name,
				Amount:   roundedOwes,
			})
		}
	}

	// 5. Compute rounded total as sum of all rounded amounts
	roundedTotal := payerAmount
	for _, np := range nonPayers {
		roundedTotal += roundedAmounts[np.person.ID]
	}

	return &SplitResponse{
		Total:        total,
		RoundedTotal: roundedTotal,
		PerPerson:    perPerson,
		Settlements:  settlements,
	}, nil
}

func floorTo1000(v float64) float64 {
	return math.Floor(v/1000) * 1000
}

func ceilTo1000(v float64) float64 {
	return math.Ceil(v/1000) * 1000
}

// ---------- Helpers ----------

type simpleError string

func (e simpleError) Error() string { return string(e) }

func errString(s string) error { return simpleError(s) }

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/split", withCORS(splitHandler))
	mux.HandleFunc("/api/health", withCORS(healthHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("splitbill backend listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
