package main

import (
	"fmt"
	"math"
	"strings"
)

// ---------- Business Logic ----------

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

	// 1. Collect non-payers with items
	type npInfo struct {
		person Person
		exact  float64
	}
	nonPayers := make([]npInfo, 0, len(req.People)-1)
	for _, p := range req.People {
		if !p.IsPayer && costs[p.ID] > 0 {
			nonPayers = append(nonPayers, npInfo{person: p, exact: costs[p.ID]})
		}
	}

	// 2. Each non-payer pays ceil(exact) — at least their fair share, rounded up to 1000
	//    Payer absorbs the difference: total - sum(nonPayers)
	sumNonPayers := 0.0
	roundedAmounts := map[string]float64{}
	for _, np := range nonPayers {
		rounded := ceilTo1000(np.exact)
		roundedAmounts[np.person.ID] = rounded
		sumNonPayers += rounded
	}
	payerAmount := total - sumNonPayers

	// 3. Build results
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

	// 4. Compute rounded total as sum of all rounded amounts
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

// ---------- Math Helpers ----------

func floorTo1000(v float64) float64 {
	return math.Floor(v/1000) * 1000
}

func ceilTo1000(v float64) float64 {
	return math.Ceil(v/1000) * 1000
}
