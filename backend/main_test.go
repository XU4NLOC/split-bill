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
	if len(resp.PerPerson) != 2 {
		t.Fatalf("expected 2 perPerson results, got %d", len(resp.PerPerson))
	}

	// An (payer): 3/5 * 10000 = 6000, owes 0
	an := resp.PerPerson[0]
	if an.Subtotal != 6000 {
		t.Errorf("expected An subtotal 6000, got %v", an.Subtotal)
	}
	if an.Owes != 0 {
		t.Errorf("expected An owes 0 (payer), got %v", an.Owes)
	}

	// Binh: 2/5 * 10000 = 4000, owes 4000
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

	// An: Pho 50000 + Che 40000 = 90000, owes 0 (payer)
	an := resp.PerPerson[0]
	if an.Subtotal != 90000 {
		t.Errorf("expected An subtotal 90000, got %v", an.Subtotal)
	}

	// Binh: Pho 50000 = 50000, owes 50000
	binh := resp.PerPerson[1]
	if binh.Owes != 50000 {
		t.Errorf("expected Binh owes 50000, got %v", binh.Owes)
	}

	// Chi: Pho 50000 = 50000, owes 50000
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
