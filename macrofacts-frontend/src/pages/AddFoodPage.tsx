import * as React from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { searchFoods, getFoodByBarcode, createCustomFood } from "../api/foods";
import { createLogEntry } from "../api/logs";

type MealKey = "breakfast" | "lunch" | "dinner" | "snacks";

function useQuery() {
    const { search } = useLocation();
    return React.useMemo(() => new URLSearchParams(search), [search]);
}

function clampMeal(v: string | null): MealKey {
    const m = String(v ?? "").toLowerCase();
    if (m === "breakfast" || m === "lunch" || m === "dinner" || m === "snacks") return m;
    return "breakfast";
}

export default function AddFoodPage() {
    const nav = useNavigate();
    const q = useQuery();
    const meal = clampMeal(q.get("meal"));

    const [tab, setTab] = React.useState<"search" | "barcode" | "custom">("search");

    const [query, setQuery] = React.useState("");
    const [barcode, setBarcode] = React.useState("");
    const [grams, setGrams] = React.useState(100);

    const [results, setResults] = React.useState<any[]>([]);
    const [loading, setLoading] = React.useState(false);
    const [err, setErr] = React.useState<string | null>(null);

    const [customName, setCustomName] = React.useState("");
    const [customKcal, setCustomKcal] = React.useState(0);
    const [customP, setCustomP] = React.useState(0);
    const [customC, setCustomC] = React.useState(0);
    const [customF, setCustomF] = React.useState(0);

    // debounced search
    React.useEffect(() => {
        if (tab !== "search") return;
        const qq = query.trim();
        if (qq.length < 2) {
            setResults([]);
            setErr(null);
            return;
        }

        setLoading(true);
        setErr(null);

        const t = window.setTimeout(async () => {
            try {
                const res = await searchFoods(qq, 25, undefined);
                setResults(res.items ?? res ?? []);
            } catch (e: any) {
                setErr(e?.message ?? "Search failed");
            } finally {
                setLoading(false);
            }
        }, 250);

        return () => window.clearTimeout(t);
    }, [query, tab]);

    async function logSelected(food: any) {
        setLoading(true);
        setErr(null);
        try {
            await createLogEntry({
                meal,
                foodId: food.id ?? food._id ?? food.food_id ?? food.code,
                quantity_g: Math.max(1, Math.round(Number(grams) || 1)),
                source: food.source ?? "off",
            });

            // go back to Today; TodayPage will refetch and pulse
            nav("/today");
        } catch (e: any) {
            setErr(e?.message ?? "Failed to log");
        } finally {
            setLoading(false);
        }
    }

    async function runBarcode() {
        const code = barcode.trim();
        if (!code) return;
        setLoading(true);
        setErr(null);
        setResults([]);
        try {
            const food = await getFoodByBarcode(code);
            if (!food) {
                setErr("No product found for this barcode.");
                return;
            }
            setResults([food]);
        } catch (e: any) {
            setErr(e?.message ?? "Barcode lookup failed");
        } finally {
            setLoading(false);
        }
    }

    async function submitCustom() {
        if (!customName.trim()) {
            setErr("Name is required.");
            return;
        }
        setLoading(true);
        setErr(null);
        try {
            const created = await createCustomFood({
                name: customName.trim(),
                kcalPer100g: Number(customKcal) || 0,
                proteinPer100g: Number(customP) || 0,
                carbsPer100g: Number(customC) || 0,
                fatPer100g: Number(customF) || 0,
            });


            await logSelected({ ...created, source: "custom" });
        } catch (e: any) {
            setErr(e?.message ?? "Failed to create custom food");
        } finally {
            setLoading(false);
        }
    }

    const title = meal.charAt(0).toUpperCase() + meal.slice(1);

    const showCreateCustomCTA =
        tab !== "custom" &&
        !loading &&
        ((tab === "search" && query.trim().length >= 2 && results.length === 0) || err?.includes("No product found"));

    return (
        <div className="dash addFoodPage">
            <header className="dashTop">
                <div className="dashTopLeft">
                    <button className="btn" onClick={() => nav("/today")}>Back</button>
                    <div>
                        <div className="dashTitle">Add food</div>
                        <div className="dashSub">{title}</div>
                    </div>
                </div>
            </header>

            <section className="card cardPad">
                <div className="segRow">
                    <button className={`seg ${tab === "search" ? "isActive" : ""}`} onClick={() => setTab("search")}>
                        Search
                    </button>
                    <button className={`seg ${tab === "barcode" ? "isActive" : ""}`} onClick={() => setTab("barcode")}>
                        Barcode
                    </button>
                    <button className={`seg ${tab === "custom" ? "isActive" : ""}`} onClick={() => setTab("custom")}>
                        Custom
                    </button>
                </div>

                <div className="dlgSection">
                    <div className="dlgLabel">Amount eaten</div>
                    <div className="dlgRow">
                        <input
                            className="dlgInput"
                            type="number"
                            min={1}
                            step={1}
                            value={grams}
                            onChange={(e) =>
                                setGrams(Math.max(1, Math.round(Number(e.target.value) || 1)))
                            }
                        />
                        <span className="dlgUnit">g</span>
                    </div>
                    <div className="dlgHint">
                        This amount will be logged to your {title.toLowerCase()}.
                    </div>
                </div>


                {tab === "search" && (
                    <>
                        <div className="dlgRowStack">
                            <input className="dlgInput" value={query} onChange={(e) => setQuery(e.target.value)} placeholder="Search foods…" />
                            <div className="dlgHint">Searches your Open Food Facts Mongo dump.</div>
                        </div>

                        {loading && <div className="dlgHint">Searching…</div>}

                        <div className="foodList">
                            {results.map((f: any) => (
                                <div key={f.id ?? f._id ?? f.code} className="foodRow" onClick={() => logSelected(f)}>
                                    <div className="foodLeft">
                                        <div className="foodName">{f.name ?? f.product_name ?? "Food"}</div>
                                        <div className="foodMeta">
                                            {Math.round(f.kcalPer100g ?? f.calories_per_100g ?? f.per100g?.calories ?? 0)} kcal / 100g
                                        </div>
                                    </div>
                                    <div className="foodRight">Select</div>
                                </div>
                            ))}
                        </div>
                    </>
                )}

                {tab === "barcode" && (
                    <>
                        <div className="dlgRowStack">
                            <input className="dlgInput" value={barcode} onChange={(e) => setBarcode(e.target.value)} placeholder="Enter barcode…" />
                            <div className="dlgActions">
                                <button className="btnSmall btnSmallPrimary" onClick={runBarcode} disabled={loading}>
                                    Lookup
                                </button>
                            </div>
                            <div className="dlgHint">Camera scanner comes next. This keeps the flow ready.</div>
                        </div>

                        <div className="foodList">
                            {results.map((f: any) => (
                                <div key={f.id ?? f._id ?? f.code} className="foodRow" onClick={() => logSelected(f)}>
                                    <div className="foodLeft">
                                        <div className="foodName">{f.name ?? f.product_name ?? "Food"}</div>
                                        <div className="foodMeta">
                                            {Math.round(f.kcalPer100g ?? f.calories_per_100g ?? f.per100g?.calories ?? 0)} kcal / 100g
                                        </div>
                                    </div>
                                    <div className="foodRight">Select</div>
                                </div>
                            ))}
                        </div>
                    </>
                )}

                {tab === "custom" && (
                    <>
                        {/* Food identity */}
                        <div className="dlgSection">
                            <div className="dlgLabel">Food</div>
                            <input
                                className="dlgInput"
                                value={customName}
                                onChange={(e) => setCustomName(e.target.value)}
                                placeholder="Food name"
                            />
                        </div>

                        {/* Nutrition definition */}
                        <div className="dlgSection">
                            <div className="dlgLabel">Nutrition (per 100g)</div>

                            <input
                                className="dlgInput"
                                type="number"
                                value={customKcal}
                                onChange={(e) => setCustomKcal(Number(e.target.value) || 0)}
                                placeholder="Calories (kcal)"
                            />

                            <div className="dlgRow">
                                <input
                                    className="dlgInput"
                                    type="number"
                                    value={customP}
                                    onChange={(e) => setCustomP(Number(e.target.value) || 0)}
                                    placeholder="Protein (g)"
                                />
                                <input
                                    className="dlgInput"
                                    type="number"
                                    value={customC}
                                    onChange={(e) => setCustomC(Number(e.target.value) || 0)}
                                    placeholder="Carbs (g)"
                                />
                                <input
                                    className="dlgInput"
                                    type="number"
                                    value={customF}
                                    onChange={(e) => setCustomF(Number(e.target.value) || 0)}
                                    placeholder="Fat (g)"
                                />
                            </div>

                            <div className="dlgHint">
                                Enter values per 100g. We scale them to your amount above.
                            </div>
                        </div>

                        <div className="dlgActions">
                            <button
                                className="btnSmall btnSmallPrimary"
                                onClick={submitCustom}
                                disabled={loading}
                            >
                                Create and log
                            </button>
                        </div>
                    </>
                )}

                {showCreateCustomCTA && (
                    <div className="dlgActions">
                        <button className="btnSmall" onClick={() => setTab("custom")}>
                            Create custom food
                        </button>
                    </div>
                )}

                {err && <div className="dlgError">{err}</div>}
            </section>
        </div>
    );
}
