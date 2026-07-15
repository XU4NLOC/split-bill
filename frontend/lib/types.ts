export interface Person {
  id: string;
  name: string;
  isPayer: boolean;
}

export interface Assignment {
  personId: string;
  quantity: number;
}

export interface Item {
  id: string;
  name: string;
  quantity: number;
  totalPrice: number;
  assignments: Assignment[];
}

export interface PersonResult {
  personId: string;
  name: string;
  isPayer: boolean;
  subtotal: number;
  roundedSubtotal: number;
  itemCount: number;
  owes: number;
  roundedOwes: number;
}

export interface Settlement {
  fromId: string;
  fromName: string;
  toId: string;
  toName: string;
  amount: number;
}

export interface SplitResponse {
  total: number;
  roundedTotal: number;
  perPerson: PersonResult[];
  settlements: Settlement[];
}

export interface ApiError {
  error: string;
}
