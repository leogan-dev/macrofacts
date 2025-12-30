import * as Dialog from "@radix-ui/react-dialog";
import { useEffect, useMemo, useState } from "react";
import type { CreateLogEntryRequest, FoodDTO } from "../api/types";
import { searchFoods } from "../api/foods";
import { createLogEntry } from "../api/logs";

function debounce<T extends (...args: any[]) => void>(fn: T, ms: number) {
    let t: any;
    return (...args: Parameters<T>) => {
        clearTimeout(t);
        t = setTimeout(() => fn(...args), ms);
    };
}

function num(x: number | null | undefined) {
    return typeof x === "number" && Number.isFinite(x) ? x : 0;
}

export function AddFoodDialog({
                                  open,
                                  onOpenChange,
                                  onAdded,
                                  prefillFood,
                              }: {
    open: boolean;
    onOpenChange: (v: boolean) => void;
    onAdded: () => void;
    prefillFood: FoodDTO | null;
}) {
    const [step, setStep] = useState<"search" | "entry">("search");
    const [q, setQ] = useState("");
    const [results, setResults] = useState<FoodDTO[]>([]);
    const [loading, setLoading] = useState(false);
    const [selected, setSelected] = useState<FoodDTO | null>(null);
    const [meal, setMeal] = useState<CreateLogEntryRequest["meal"]>("breakfast");
    const [grams, setGrams] = useState<number>(100);
    const [err, setErr] = useState<string | null>(null);

    useEffect(() => {
        if (!open) {
            setStep("search");
            setQ("");
            setResults([]);
            setLoading(false);
            setSelected(null);
            setMeal("breakfast");
            setGrams(100);
            setErr(null);
        }
    }, [open]);

    useEffect(() => {
        if (open && prefillFood) {
            setSelected(prefillFood);
            setStep("entry");
            setGrams(Math.round(prefillFood.servingG || 100));
        }
    }, [open, prefillFood]);

    const doSearch = useMemo(
        () =>
            debounce(async (text: string) => {
                if (!text || text.trim().length < 2) {
                    setResults([]);
                    return;
                }
                setLoading(true);
                setErr(null);
                try {
                    const r = await searchFoods(text.trim(), 25);
                    setResults(r.items || []);
                } catch (e: any) {
                    setErr(e?.message || "Search failed");
                } finally {
                    setLoading(false);
                }
            }, 250),
        []
    );

    useEffect(() => {
        if (!open) return;
        doSearch(q);
    }, [q, doSearch, open]);

    const preview = useMemo(() => {
        if (!selected) return null;
        const factor = grams / 100;

        const kcal = Math.round(num(selected.kcalPer100g) * factor);
        const protein = num(selected.proteinPer100g) * factor;
        const carbs = num(selected.carbsPer100g) * factor;
        const fat = num(selected.fatPer100g) * factor;

        const fiber = num(selected.fiberPer100g) * factor;
        const sugar = num(selected.sugarPer100g) * factor;
        const salt = num(selected.saltPer100g) * factor;

        return { kcal, protein, carbs, fat, fiber, sugar, salt };
    }, [selected, grams]);

    async function confirm() {
        if (!selected) return;
        setErr(null);

        const req: CreateLogEntryRequest = {
            meal,
            source: selected.source,
            quantity_g: grams,
        };

        if (selected.source === "off") {
            if (!selected.barcode) {
                setErr("This OFF item is missing a barcode. Try another result.");
                return;
            }
            req.barcode = selected.barcode;
        } else {
            req.foodId = selected.id;
        }

        try {
            await createLogEntry(req);
            onAdded();
        } catch (e: any) {
            setErr(e?.message || "Failed to add entry");
        }
    }

    return (
        <Dialog.Root open={open} onOpenChange={onOpenChange}>
            <Dialog.Portal>
                <Dialog.Overlay className="dlgOverlay" />
                <Dialog.Content className="dlgContent card">
                    <div className="dlgHead">
                        <Dialog.Title className="dlgTitle">{step === "search" ? "Add food" : "Add entry"}</Dialog.Title>
                        <Dialog.Close className="iconBtn" aria-label="Close">
                            ✕
                        </Dialog.Close>
                    </div>

                    {err ? (
                        <div className="errorText" style={{ marginTop: 10 }}>
                            {err}
                        </div>
                    ) : null}

                    {step === "search" ? (
                        <>
                            <div style={{ marginTop: 12 }}>
                                <input
                                    className="input"
                                    placeholder="Search foods… (e.g. skyr, chicken, rice)"
                                    value={q}
                                    onChange={(e) => setQ(e.target.value)}
                                    autoFocus
                                />
                                <div className="hint">Search hits OpenFoodFacts first. Custom foods UI comes next.</div>
                            </div>

                            <div className="results">
                                {loading ? <div className="muted">Searching…</div> : null}
                                {!loading && results.length === 0 && q.trim().length >= 2 ? (
                                    <div className="empty">No results. Try fewer words.</div>
                                ) : null}

                                {results.map((f) => (
                                    <button
                                        key={`${f.source}:${f.id}`}
                                        className="resultRow"
                                        onClick={() => {
                                            setSelected(f);
                                            setStep("entry");
                                            setGrams(Math.round(f.servingG || 100));
                                        }}
                                    >
                                        <div className="resultMain">
                                            <div className="resultName">
                                                {f.name} {f.brand ? <span className="muted">· {f.brand}</span> : null}
                                            </div>
                                            <div className="muted">{Math.round(num(f.kcalPer100g))} kcal / 100g</div>
                                        </div>
                                        <div className="resultMacros muted">
                                            P {Math.round(num(f.proteinPer100g))} · C {Math.round(num(f.carbsPer100g))} · F{" "}
                                            {Math.round(num(f.fatPer100g))}
                                        </div>
                                    </button>
                                ))}
                            </div>

                            <div className="dlgFoot">
                                <button className="btn ghost" onClick={() => onOpenChange(false)}>
                                    Cancel
                                </button>
                            </div>
                        </>
                    ) : (
                        <>
                            {selected ? (
                                <div className="entryBox">
                                    <div className="entryTitle">
                                        {selected.name} {selected.brand ? <span className="muted">· {selected.brand}</span> : null}
                                    </div>

                                    <div className="entryGrid">
                                        <label className="label">
                                            Meal
                                            <select className="select" value={meal} onChange={(e) => setMeal(e.target.value as any)}>
                                                <option value="breakfast">Breakfast</option>
                                                <option value="lunch">Lunch</option>
                                                <option value="dinner">Dinner</option>
                                                <option value="snacks">Snacks</option>
                                            </select>
                                        </label>

                                        <label className="label">
                                            Quantity (grams)
                                            <input
                                                className="input"
                                                type="number"
                                                min={1}
                                                max={5000}
                                                value={grams}
                                                onChange={(e) => setGrams(Math.max(1, Math.min(5000, Number(e.target.value || 0))))}
                                            />
                                            {selected.servingG ? (
                                                <div className="hint">
                                                    Serving: {Math.round(selected.servingG)}g{" "}
                                                    <button className="linkBtn" type="button" onClick={() => setGrams(Math.round(selected.servingG || 100))}>
                                                        use serving
                                                    </button>
                                                </div>
                                            ) : (
                                                <div className="hint">Tip: 100g is the standard nutrition reference.</div>
                                            )}
                                        </label>
                                    </div>

                                    <div className="preview">
                                        <div className="previewBig">{preview ? preview.kcal : 0} kcal</div>
                                        <div className="muted">
                                            P {preview ? Math.round(preview.protein) : 0} · C {preview ? Math.round(preview.carbs) : 0} · F{" "}
                                            {preview ? Math.round(preview.fat) : 0}
                                        </div>
                                        <div className="muted" style={{ marginTop: 6 }}>
                                            Fiber {preview ? Math.round(preview.fiber) : 0}g · Sugar {preview ? Math.round(preview.sugar) : 0}g · Salt{" "}
                                            {preview ? Math.round(preview.salt) : 0}g
                                        </div>
                                    </div>
                                </div>
                            ) : null}

                            <div className="dlgFoot">
                                <button
                                    className="btn ghost"
                                    onClick={() => {
                                        setStep("search");
                                        setSelected(null);
                                    }}
                                >
                                    Back
                                </button>
                                <button className="btn" onClick={confirm} disabled={!selected}>
                                    Add to today
                                </button>
                            </div>
                        </>
                    )}
                </Dialog.Content>
            </Dialog.Portal>
        </Dialog.Root>
    );
}
