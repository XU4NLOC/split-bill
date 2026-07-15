package main

import "testing"

func TestCalculateSplit_Basic(t *testing.T) {
	req := SplitRequest{
		People: []Person{
			{ID: "p1", Name: "An", IsPayer: true},
			{ID: "p2", Name: "Binh"},
		},
		Items: []Item{
			{
				ID:         "i1",
				Name:       "Candy",
				Quantity:   5,
				TotalPrice: 10000,
				Assignments: []Assignment{
					{PersonID: "p1", Quantity: 3},
					{PersonID: "p2", Quantity: 2},
				},
			},
		},
	}

	resp, err := calculateSplit(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 10000 {
		t.Errorf("expected total 10000, got %v", resp.Total)
	}
	if resp.RoundedTotal != 10000 {
		t.Errorf("expected roundedTotal 10000, got %v", resp.RoundedTotal)
	}
	if len(resp.PerPerson) != 2 {
		t.Fatalf("expected 2 perPerson results, got %d", len(resp.PerPerson))
	}

	an := resp.PerPerson[0]
	if an.Subtotal != 6000 {
		t.Errorf("expected An subtotal 6000, got %v", an.Subtotal)
	}
	if an.Owes != 0 {
		t.Errorf("expected An owes 0 (payer), got %v", an.Owes)
	}

	binh := resp.PerPerson[1]
	if binh.Subtotal != 4000 {
		t.Errorf("expected Binh subtotal 4000, got %v", binh.Subtotal)
	}
	if binh.Owes != 4000 {
		t.Errorf("expected Binh owes 4000, got %v", binh.Owes)
	}

	if len(resp.Settlements) != 1 {
		t.Fatalf("expected 1 settlement, got %d", len(resp.Settlements))
	}
	s := resp.Settlements[0]
	if s.FromID != "p2" || s.ToID != "p1" {
		t.Errorf("expected settlement Binh->An, got %s->%s", s.FromID, s.ToID)
	}
	if s.Amount != 4000 {
		t.Errorf("expected settlement amount 4000, got %v", s.Amount)
	}
}

func TestCalculateSplit_RoundingPayerDownNonPayerUp(t *testing.T) {
	// 3 people, 100000 VND total, each buys 1 of 3 items
	// Exact share: 33333.33 each
	// Payer (An): floor(33333.33) = 33000
	// Leftover: 100000 - 33000 = 67000
	// 2 non-payers: 67000 / 2 = 33500 each → base = 33000, remainder = 1000
	// Last non-payer gets +1000 → 34000
	// Total: 33000 + 33000 + 34000 = 100000 ✓
	req := SplitRequest{
		People: []Person{
			{ID: "p1", Name: "An", IsPayer: true},
			{ID: "p2", Name: "Binh"},
			{ID: "p3", Name: "Chi"},
		},
		Items: []Item{
			{
				ID:         "i1",
				Name:       "Pho",
				Quantity:   3,
				TotalPrice: 100000,
				Assignments: []Assignment{
					{PersonID: "p1", Quantity: 1},
					{PersonID: "p2", Quantity: 1},
					{PersonID: "p3", Quantity: 1},
				},
			},
		},
	}

	resp, err := calculateSplit(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	an := resp.PerPerson[0]
	if an.RoundedOwes != 33000 {
		t.Errorf("expected An (payer) rounded 33000, got %v", an.RoundedOwes)
	}

	binh := resp.PerPerson[1]
	if binh.RoundedOwes != 33000 {
		t.Errorf("expected Binh (non-payer) rounded 33000, got %v", binh.RoundedOwes)
	}

	chi := resp.PerPerson[2]
	if chi.RoundedOwes != 34000 {
		t.Errorf("expected Chi (non-payer) rounded 34000, got %v", chi.RoundedOwes)
	}

	// Total must match exactly
	if resp.RoundedTotal != 100000 {
		t.Errorf("expected roundedTotal 100000, got %v", resp.RoundedTotal)
	}

	// Settlement amounts
	if len(resp.Settlements) != 2 {
		t.Fatalf("expected 2 settlements, got %d", len(resp.Settlements))
	}
	amounts := map[string]float64{}
	for _, s := range resp.Settlements {
		amounts[s.FromName] = s.Amount
	}
	if amounts["Binh"] != 33000 {
		t.Errorf("expected Binh settlement 33000, got %v", amounts["Binh"])
	}
	if amounts["Chi"] != 34000 {
		t.Errorf("expected Chi settlement 34000, got %v", amounts["Chi"])
	}
}

func TestCalculateSplit_EvenSplit(t *testing.T) {
	// 2 people, 10000 VND, 1 item each
	// Exact: 5000 each
	// Payer: floor(5000) = 5000
	// Leftover: 10000 - 5000 = 5000
	// 1 non-payer: base = 5000, no remainder
	// Total: 5000 + 5000 = 10000 ✓
	req := SplitRequest{
		People: []Person{
			{ID: "p1", Name: "An", IsPayer: true},
			{ID: "p2", Name: "Binh"},
		},
		Items: []Item{
			{
				ID:         "i1",
				Name:       "Che",
				Quantity:   2,
				TotalPrice: 10000,
				Assignments: []Assignment{
					{PersonID: "p1", Quantity: 1},
					{PersonID: "p2", Quantity: 1},
				},
			},
		},
	}

	resp, err := calculateSplit(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.PerPerson[0].RoundedOwes != 5000 {
		t.Errorf("expected payer rounded 5000, got %v", resp.PerPerson[0].RoundedOwes)
	}
	if resp.PerPerson[1].RoundedOwes != 5000 {
		t.Errorf("expected non-payer rounded 5000, got %v", resp.PerPerson[1].RoundedOwes)
	}
	if resp.RoundedTotal != 10000 {
		t.Errorf("expected roundedTotal 10000, got %v", resp.RoundedTotal)
	}
}

func TestCalculateSplit_FourPeople(t *testing.T) {
	// 4 people, 100000 VND total
	// Exact: 25000 each
	// Payer: floor(25000) = 25000
	// Leftover: 100000 - 25000 = 75000
	// 3 non-payers: 75000 / 3 = 25000 each → no remainder
	// Total: 25000 + 25000*3 = 100000 ✓
	req := SplitRequest{
		People: []Person{
			{ID: "p1", Name: "An", IsPayer: true},
			{ID: "p2", Name: "Binh"},
			{ID: "p3", Name: "Chi"},
			{ID: "p4", Name: "Dung"},
		},
		Items: []Item{
			{
				ID:         "i1",
				Name:       "Pho",
				Quantity:   4,
				TotalPrice: 100000,
				Assignments: []Assignment{
					{PersonID: "p1", Quantity: 1},
					{PersonID: "p2", Quantity: 1},
					{PersonID: "p3", Quantity: 1},
					{PersonID: "p4", Quantity: 1},
				},
			},
		},
	}

	resp, err := calculateSplit(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, pr := range resp.PerPerson {
		if pr.RoundedOwes != 25000 {
			t.Errorf("expected %s rounded 25000, got %v", pr.Name, pr.RoundedOwes)
		}
	}
	if resp.RoundedTotal != 100000 {
		t.Errorf("expected roundedTotal 100000, got %v", resp.RoundedTotal)
	}
}

func TestCalculateSplit_MultipleItems(t *testing.T) {
	req := SplitRequest{
		People: []Person{
			{ID: "p1", Name: "An", IsPayer: true},
			{ID: "p2", Name: "Binh"},
			{ID: "p3", Name: "Chi"},
		},
		Items: []Item{
			{
				ID:         "i1",
				Name:       "Pho",
				Quantity:   3,
				TotalPrice: 150000,
				Assignments: []Assignment{
					{PersonID: "p1", Quantity: 1},
					{PersonID: "p2", Quantity: 1},
					{PersonID: "p3", Quantity: 1},
				},
			},
			{
				ID:         "i2",
				Name:       "Che",
				Quantity:   2,
				TotalPrice: 40000,
				Assignments: []Assignment{
					{PersonID: "p1", Quantity: 2},
				},
			},
		},
	}

	resp, err := calculateSplit(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 190000 {
		t.Errorf("expected total 190000, got %v", resp.Total)
	}

	an := resp.PerPerson[0]
	if an.Subtotal != 90000 {
		t.Errorf("expected An subtotal 90000, got %v", an.Subtotal)
	}

	binh := resp.PerPerson[1]
	if binh.Owes != 50000 {
		t.Errorf("expected Binh owes 50000, got %v", binh.Owes)
	}

	chi := resp.PerPerson[2]
	if chi.Owes != 50000 {
		t.Errorf("expected Chi owes 50000, got %v", chi.Owes)
	}

	if len(resp.Settlements) != 2 {
		t.Fatalf("expected 2 settlements, got %d", len(resp.Settlements))
	}
}

func TestCalculateSplit_NoPayer(t *testing.T) {
	req := SplitRequest{
		People: []Person{{ID: "p1", Name: "An"}},
	}
	if _, err := calculateSplit(req); err == nil {
		t.Fatal("expected error when no payer is set")
	}
}

func TestCalculateSplit_AssignmentMismatch(t *testing.T) {
	req := SplitRequest{
		People: []Person{
			{ID: "p1", Name: "An", IsPayer: true},
			{ID: "p2", Name: "Binh"},
		},
		Items: []Item{
			{
				ID:         "i1",
				Name:       "Candy",
				Quantity:   5,
				TotalPrice: 10000,
				Assignments: []Assignment{
					{PersonID: "p1", Quantity: 2},
					{PersonID: "p2", Quantity: 1},
				},
			},
		},
	}
	_, err := calculateSplit(req)
	if err == nil {
		t.Fatal("expected error when assignment quantity does not match item quantity")
	}
}

func TestCalculateSplit_UnknownPersonInAssignment(t *testing.T) {
	req := SplitRequest{
		People: []Person{
			{ID: "p1", Name: "An", IsPayer: true},
		},
		Items: []Item{
			{
				ID:         "i1",
				Name:       "Candy",
				Quantity:   5,
				TotalPrice: 10000,
				Assignments: []Assignment{
					{PersonID: "p1", Quantity: 3},
					{PersonID: "ghost", Quantity: 2},
				},
			},
		},
	}
	_, err := calculateSplit(req)
	if err == nil {
		t.Fatal("expected error when assignment references unknown person")
	}
}
