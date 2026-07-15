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
	if an.RoundedOwes != 32000 {
		t.Errorf("expected An (payer) rounded 32000, got %v", an.RoundedOwes)
	}

	binh := resp.PerPerson[1]
	if binh.RoundedOwes != 34000 {
		t.Errorf("expected Binh (non-payer) rounded 34000, got %v", binh.RoundedOwes)
	}

	chi := resp.PerPerson[2]
	if chi.RoundedOwes != 34000 {
		t.Errorf("expected Chi (non-payer) rounded 34000, got %v", chi.RoundedOwes)
	}

	if resp.RoundedTotal != 100000 {
		t.Errorf("expected roundedTotal 100000, got %v", resp.RoundedTotal)
	}
}

func TestCalculateSplit_EvenSplit(t *testing.T) {
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

func TestCalculateSplit_ZeroItemPersonExcluded(t *testing.T) {
	req := SplitRequest{
		People: []Person{
			{ID: "p1", Name: "An", IsPayer: true},
			{ID: "p2", Name: "Binh"},
			{ID: "p3", Name: "Chi"},
			{ID: "p4", Name: "Dung"},
			{ID: "p5", Name: "Em"},
			{ID: "p6", Name: "Fi"},
		},
		Items: []Item{
			{
				ID:         "i1",
				Name:       "Item",
				Quantity:   12,
				TotalPrice: 100000,
				Assignments: []Assignment{
					{PersonID: "p1", Quantity: 2},
					{PersonID: "p2", Quantity: 3},
					{PersonID: "p3", Quantity: 2},
					{PersonID: "p4", Quantity: 1},
					{PersonID: "p5", Quantity: 4},
					{PersonID: "p6", Quantity: 0},
				},
			},
		},
	}

	resp, err := calculateSplit(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	an := resp.PerPerson[0]
	if an.RoundedOwes != 15000 {
		t.Errorf("expected An (payer) rounded 15000, got %v", an.RoundedOwes)
	}

	fi := resp.PerPerson[5]
	if fi.RoundedOwes != 0 {
		t.Errorf("expected Fi (0 items) rounded 0, got %v", fi.RoundedOwes)
	}
	if fi.Subtotal != 0 {
		t.Errorf("expected Fi subtotal 0, got %v", fi.Subtotal)
	}

	expected := map[string]float64{
		"p2": 25000,
		"p3": 17000,
		"p4": 9000,
		"p5": 34000,
	}
	for _, pr := range resp.PerPerson {
		if exp, ok := expected[pr.PersonID]; ok {
			if pr.RoundedOwes != exp {
				t.Errorf("expected %s rounded %v, got %v", pr.Name, exp, pr.RoundedOwes)
			}
		}
	}

	if resp.RoundedTotal != 100000 {
		t.Errorf("expected roundedTotal 100000, got %v", resp.RoundedTotal)
	}

	sum := 0.0
	for _, pr := range resp.PerPerson {
		sum += pr.RoundedOwes
	}
	if sum != 100000 {
		t.Errorf("sum of rounded amounts = %v, expected 100000", sum)
	}
}
