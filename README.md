# Split the Bill Calculator (Just an entertaining project)

Itemized bill splitting for a group dinner: mark who paid, list each item with
its price, assign each item to the person who ordered it, and get back exactly
what everyone owes the payer. No accounts, no database — everything lives in
the browser session and the calculation happens on a Go backend.

## Stack

- **Backend:** Go (standard library only, no framework/DB) — `/api/split`
- **Frontend:** Next.js 14 (App Router) + TypeScript
- **Currency:** VND

## Run locally with Docker

```bash
docker compose up --build
```

- Frontend: http://localhost:3000
- Backend health check: http://localhost:8080/api/health

## Run without Docker

**Backend**

```bash
cd backend
go run .
# listens on :8080
```

**Frontend** (in a second terminal)

```bash
cd frontend
npm install
NEXT_PUBLIC_API_URL=http://localhost:8080 npm run dev
# serves on :3000
```

## How the math works

1. Every item is assigned to exactly one person (the person who ordered it).
2. Each person's subtotal = sum of the prices of items assigned to them.
3. The payer already covered the whole bill, so every non-payer owes their
   own subtotal back to the payer.
4. No tax/tip/service-charge logic is applied — enter item prices as the
   final amount you want split.

## API

`POST /api/split`

```json
{
  "people": [
    { "id": "p1", "name": "An", "isPayer": true },
    { "id": "p2", "name": "Binh", "isPayer": false }
  ],
  "items": [
    { "id": "i1", "name": "Pho", "price": 50000, "personId": "p1" },
    { "id": "i2", "name": "Bun cha", "price": 60000, "personId": "p2" }
  ]
}
```

Returns total, a per-person subtotal breakdown, and a list of settlements
(who owes whom, and how much).

## Tests

```bash
cd backend
go test ./...
```

## Notes / possible next steps

- Shared items (split among a subset of people) and tax/tip handling were
  intentionally left out of this version — the current model assumes one
  owner per item and a tax-free total.
- Bills are one-time/in-memory only; nothing is persisted between sessions.
