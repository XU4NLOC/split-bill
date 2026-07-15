"use client";

import { useCallback, useMemo, useRef, useState } from "react";
import styles from "./page.module.css";
import { calculateSplit } from "@/lib/api";
import { captureElement, copyImageToClipboard, downloadImage } from "@/lib/export";
import { formatVND } from "@/lib/format";
import type { Item, Person, SplitResponse } from "@/lib/types";

let idCounter = 0;
function genId(prefix: string): string {
  idCounter += 1;
  return `${prefix}_${Date.now().toString(36)}_${idCounter}`;
}

function newPerson(name: string, isPayer = false): Person {
  return { id: genId("person"), name, isPayer };
}

function newItem(): Item {
  return { id: genId("item"), name: "", quantity: 1, totalPrice: 0, assignments: [] };
}

export default function Home() {
  const [people, setPeople] = useState<Person[]>(() => [
    newPerson("", true),
    newPerson(""),
  ]);
  const [items, setItems] = useState<Item[]>([]);
  const [results, setResults] = useState<SplitResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [exporting, setExporting] = useState(false);
  const [exportStatus, setExportStatus] = useState<"idle" | "copied" | "error">("idle");
  const resultsRef = useRef<HTMLDivElement>(null);

  const runningTotal = useMemo(
    () => items.reduce((sum, item) => sum + (Number(item.totalPrice) || 0), 0),
    [items]
  );

  function updatePerson(id: string, patch: Partial<Person>) {
    setPeople((prev) => prev.map((p) => (p.id === id ? { ...p, ...patch } : p)));
    setResults(null);
  }

  function setPayer(id: string) {
    setPeople((prev) => prev.map((p) => ({ ...p, isPayer: p.id === id })));
    setResults(null);
  }

  function addPerson() {
    const p = newPerson("");
    setPeople((prev) => [...prev, p]);
    setItems((prev) =>
      prev.map((it) => ({
        ...it,
        assignments: [...it.assignments, { personId: p.id, quantity: 0 }],
      }))
    );
  }

  function removePerson(id: string) {
    setPeople((prev) => {
      const next = prev.filter((p) => p.id !== id);
      if (!next.some((p) => p.isPayer) && next.length > 0) {
        next[0] = { ...next[0], isPayer: true };
      }
      return next;
    });
    setItems((prev) =>
      prev.map((it) => ({
        ...it,
        assignments: it.assignments.filter((a) => a.personId !== id),
      }))
    );
    setResults(null);
  }

  function addItem() {
    const namedPeople = people.filter((p) => p.name.trim() !== "");
    const item = newItem();
    item.assignments = namedPeople.map((p) => ({ personId: p.id, quantity: 0 }));
    setItems((prev) => [...prev, item]);
  }

  function updateItem(id: string, patch: Partial<Item>) {
    setItems((prev) => prev.map((it) => (it.id === id ? { ...it, ...patch } : it)));
    setResults(null);
  }

  function removeItem(id: string) {
    setItems((prev) => prev.filter((it) => it.id !== id));
    setResults(null);
  }

  function updateAssignment(itemId: string, personId: string, qty: number) {
    setItems((prev) =>
      prev.map((it) => {
        if (it.id !== itemId) return it;
        const exists = it.assignments.find((a) => a.personId === personId);
        if (exists) {
          return {
            ...it,
            assignments: it.assignments.map((a) =>
              a.personId === personId ? { ...a, quantity: qty } : a
            ),
          };
        }
        return {
          ...it,
          assignments: [...it.assignments, { personId, quantity: qty }],
        };
      })
    );
    setResults(null);
  }

  function resetAll() {
    setPeople([newPerson("", true), newPerson("")]);
    setItems([]);
    setResults(null);
    setError(null);
  }

  function validate(): string | null {
    const namedPeople = people.filter((p) => p.name.trim() !== "");
    if (namedPeople.length < 1) return "Add at least one person.";
    if (!people.some((p) => p.isPayer)) return "Mark one person as the payer.";
    if (items.length === 0) return "Add at least one item.";
    for (const it of items) {
      if (it.name.trim() === "") return "Every item needs a name.";
      if (it.quantity < 1) return `"${it.name}" needs a quantity of at least 1.`;
      if (it.totalPrice <= 0) return `"${it.name}" needs a total price above 0.`;
      const assigned = it.assignments.reduce((s, a) => s + a.quantity, 0);
      if (assigned !== it.quantity)
        return `"${it.name}": assigned ${assigned}/${it.quantity} — quantities must match.`;
    }
    return null;
  }

  async function handleCalculate() {
    const validationError = validate();
    if (validationError) {
      setError(validationError);
      setResults(null);
      return;
    }
    setError(null);
    setLoading(true);
    try {
      const cleanPeople = people.filter((p) => p.name.trim() !== "");
      const resp = await calculateSplit(cleanPeople, items);
      setResults(resp);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Something went wrong.");
      setResults(null);
    } finally {
      setLoading(false);
    }
  }

  const handleDownload = useCallback(async () => {
    if (!resultsRef.current || exporting) return;
    setExporting(true);
    try {
      const blob = await captureElement(resultsRef.current);
      downloadImage(blob, "splitbill-result.png");
    } catch {
      setExportStatus("error");
      setTimeout(() => setExportStatus("idle"), 2000);
    } finally {
      setExporting(false);
    }
  }, [exporting]);

  const handleCopy = useCallback(async () => {
    if (!resultsRef.current || exporting) return;
    setExporting(true);
    try {
      const blob = await captureElement(resultsRef.current);
      const ok = await copyImageToClipboard(blob);
      setExportStatus(ok ? "copied" : "error");
      setTimeout(() => setExportStatus("idle"), 2000);
    } catch {
      setExportStatus("error");
      setTimeout(() => setExportStatus("idle"), 2000);
    } finally {
      setExporting(false);
    }
  }, [exporting]);

  const payer = people.find((p) => p.isPayer);
  const namedPeople = people.filter((p) => p.name.trim() !== "");

  return (
    <main className={styles.page}>
      <div className={styles.receiptWrap}>
        <div className={`${styles.torn} ${styles.tornTop}`} />
        <div className={styles.receipt}>
          <header className={styles.header}>
            <div className={styles.eyebrow}>Chia Bill · Split Calculator</div>
            <h1 className={styles.title}>Who Owes What</h1>
            <p className={styles.subtitle}>
              Add everyone at the table, enter the bill details, and assign how
              many of each item each person gets. We&apos;ll work out exactly what
              each person owes the payer.
            </p>
          </header>

          <section className={styles.section}>
            <h2 className={styles.sectionTitle}>
              <span>People</span>
              <span>tap ● to set payer</span>
            </h2>
            <div className={styles.rowList}>
              {people.map((p) => (
                <div className={styles.personRow} key={p.id}>
                  <button
                    type="button"
                    className={styles.payerToggle}
                    data-active={p.isPayer}
                    title={p.isPayer ? "This person paid the bill" : "Mark as payer"}
                    aria-pressed={p.isPayer}
                    onClick={() => setPayer(p.id)}
                  >
                    ●
                  </button>
                  <input
                    className={styles.textInput}
                    placeholder="Name"
                    value={p.name}
                    onChange={(e) => updatePerson(p.id, { name: e.target.value })}
                  />
                  <button
                    type="button"
                    className={styles.removeBtn}
                    onClick={() => removePerson(p.id)}
                    aria-label={`Remove ${p.name || "person"}`}
                  >
                    ×
                  </button>
                </div>
              ))}
            </div>
            <button type="button" className={styles.addBtn} onClick={addPerson}>
              + Add person
            </button>
          </section>

          <section className={styles.section}>
            <h2 className={styles.sectionTitle}>
              <span>Items</span>
              <span>{items.length} item{items.length === 1 ? "" : "s"}</span>
            </h2>
            <div className={styles.rowList}>
              {items.map((it) => {
                const assigned = it.assignments.reduce((s, a) => s + a.quantity, 0);
                const matched = assigned === it.quantity;
                return (
                  <div className={styles.itemCard} key={it.id}>
                    <div className={styles.itemRow1}>
                      <input
                        className={`${styles.textInput} ${styles.itemName}`}
                        placeholder="Item name"
                        value={it.name}
                        onChange={(e) => updateItem(it.id, { name: e.target.value })}
                      />
                      <button
                        type="button"
                        className={styles.removeBtn}
                        onClick={() => removeItem(it.id)}
                        aria-label={`Remove ${it.name || "item"}`}
                      >
                        ×
                      </button>
                    </div>
                    <div className={styles.itemRow2}>
                      <span className={styles.itemLabel}>Qty</span>
                      <input
                        className={styles.itemInput}
                        type="number"
                        min={1}
                        step={1}
                        value={it.quantity}
                        onChange={(e) => {
                          const v = parseInt(e.target.value, 10);
                          updateItem(it.id, { quantity: isNaN(v) ? 1 : v });
                        }}
                      />
                      <span className={styles.itemLabel}>Total</span>
                      <input
                        className={styles.itemInput}
                        type="number"
                        min={0}
                        step={1000}
                        value={it.totalPrice}
                        onChange={(e) => {
                          const v = parseFloat(e.target.value);
                          updateItem(it.id, { totalPrice: isNaN(v) ? 0 : v });
                        }}
                      />
                    </div>
                    {namedPeople.length > 0 && (
                      <div className={styles.assignmentList}>
                        {namedPeople.map((p) => {
                          const assignment = it.assignments.find(
                            (a) => a.personId === p.id
                          );
                          return (
                            <div className={styles.assignmentRow} key={p.id}>
                              <span className={styles.assignmentName}>{p.name}</span>
                              <input
                                className={styles.assignmentQty}
                                type="number"
                                min={0}
                                max={it.quantity}
                                step={1}
                                value={assignment?.quantity ?? 0}
                                onChange={(e) => {
                                  const v = parseInt(e.target.value, 10);
                                  updateAssignment(it.id, p.id, isNaN(v) ? 0 : v);
                                }}
                              />
                            </div>
                          );
                        })}
                      </div>
                    )}
                    <div
                      className={`${styles.assignedInfo} ${
                        matched ? styles.assignedInfoValid : styles.assignedInfoInvalid
                      }`}
                    >
                      assigned: {assigned} / {it.quantity}
                    </div>
                  </div>
                );
              })}
            </div>
            <button type="button" className={styles.addBtn} onClick={addItem}>
              + Add item
            </button>

            <div className={styles.totalRow}>
              <span className={styles.totalLabel}>Running total</span>
              <span className={styles.totalAmount}>{formatVND(runningTotal)}</span>
            </div>
          </section>

          <section>
            <button
              type="button"
              className={styles.calculateBtn}
              onClick={handleCalculate}
              disabled={loading}
            >
              {loading ? "Calculating…" : "Calculate split"}
            </button>
            {error && <div className={styles.errorBox}>{error}</div>}
          </section>

          {results && payer && (
            <section className={styles.resultsSection} ref={resultsRef}>
              <div className={styles.stamp}>
                {formatVND(results.roundedTotal)}
                <br />
                paid by {payer.name}
              </div>
              <h2 className={styles.sectionTitle}>
                <span>Per person</span>
                <span>subtotal</span>
              </h2>
              <div className={styles.rowList}>
                {results.perPerson.map((pr) => (
                  <div className={styles.leaderRow} key={pr.personId}>
                    <span className={styles.leaderName}>
                      {pr.name}
                      {pr.isPayer ? " (paid)" : ""}
                    </span>
                    <span className={styles.leaderFill} />
                    <span className={styles.leaderAmount}>
                      {formatVND(pr.roundedSubtotal)}
                    </span>
                  </div>
                ))}
              </div>

              <h2
                className={styles.sectionTitle}
                style={{ marginTop: 22 }}
              >
                <span>Settle up</span>
              </h2>
              {results.settlements.length === 0 ? (
                <div className={styles.allSettled}>
                  Everything&apos;s already even — nothing to settle.
                </div>
              ) : (
                <div className={styles.settlementList}>
                  {results.settlements.map((s, i) => (
                    <div className={styles.settlementCard} key={i}>
                      <div className={styles.settlementLine}>
                        <strong>{s.fromName}</strong> owes{" "}
                        <strong>{s.toName}</strong>
                      </div>
                      <div className={styles.settlementAmount}>
                        {formatVND(s.amount)}
                      </div>
                    </div>
                  ))}
                </div>
              )}

              <div className={styles.exportRow}>
                <button
                  type="button"
                  className={styles.exportBtn}
                  onClick={handleDownload}
                  disabled={exporting}
                >
                  {exporting ? "Exporting…" : "Download"}
                </button>
                <button
                  type="button"
                  className={styles.exportBtn}
                  onClick={handleCopy}
                  disabled={exporting}
                >
                  {exporting ? "Exporting…" : "Copy"}
                </button>
              </div>
              {exportStatus === "copied" && (
                <div className={styles.exportToast}>Copied!</div>
              )}
              {exportStatus === "error" && (
                <div className={styles.exportToastError}>Export failed</div>
              )}
            </section>
          )}

          <button type="button" className={styles.resetLink} onClick={resetAll}>
            Start a new bill
          </button>
        </div>
        <div className={`${styles.torn} ${styles.tornBottom}`} />
      </div>
    </main>
  );
}
