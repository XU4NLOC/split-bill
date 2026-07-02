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
  itemCount: number;
  owes: number;
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
  perPerson: PersonResult[];
  settlements: Settlement[];
}

export interface ApiError {
  error: string;
}
