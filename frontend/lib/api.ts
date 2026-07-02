import type { Person, Item, SplitResponse, ApiError } from "./types";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export async function calculateSplit(
  people: Person[],
  items: Item[]
): Promise<SplitResponse> {
  const res = await fetch(`${API_URL}/api/split`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ people, items }),
  });

  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as ApiError | null;
    throw new Error(body?.error || `Request failed with status ${res.status}`);
  }

  return res.json();
}
