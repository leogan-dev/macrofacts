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

    // Custom food fields
    const [customName, setCustomName] = React.useState("");
    const [kcal, setKcal] = React.useState(0);
    const [protein, setProtein] = React.useState(0);
    const [carbs, setCarbs] = React.useState(0);
    const [fat, setFat] = React.useState(0);

    const [sugars, setSugars] = React.useState(0);
    const [fiber, setFiber] = React.useState(0);
    const [salt, setSalt] = React.useState(0);
    const [sodium, setSodium] = React.useState(0);
    const [satFat, setSatFat] = React.useState(0);
    const [monoFat, setMonoFat] = React.useState(0);
    const [polyFat, setPolyFat] = React.useState(0);

    React.useEffect(() => {
        if (tab !== "search") return;
        const qq = query.trim();
        if (qq.length < 2) {
            setResults([]);
            setErr(null);
            return;
        }

        setLoading(true);
        const t = window.setTimeout(async () => {
            try {
                const res = await searchFoods(qq, 25, undefined);
                setResults(res.items ?? res ?? []);
            } catch {
                setErr("Search failed");
            } finally {
                setLoading(false);
            }
        }, 250);

        return () => window.clearTimeout(t);
    }, [query, tab]);

    async function logSelected(food: any) {
        setLoading(true);
        try {
            await createLogEntry({
                meal,
                foodId: food.id ?? food.code,
                quantity_g: Math.max(1, Math.round(grams)),
                source: food.source ?? "off",
            });
            nav("/today");
        } finally {
            setLoading(false);
        }
    }

    async function submitCustom() {
        if (!customName.trim()) {
            setErr("Name is required");
            return;
        }

        setLoading(true);
        setErr(null);

        try {
            const nutriments: Record<string, number> = {
                "energy-kcal_100g": kcal,
                "proteins_100g": protein,
                "carbohydrates_100g": carbs,
                "fat_100g": fat,
                "sugars_100g": sugars,
                "fiber_100g": fiber,
                "salt_100g": salt,
                "sodium_100g": sodium,
                "saturated-fat_100g": satFat,
                "monounsaturated-fat_100g": monoFat,
                "polyunsaturated-fat_100g": polyFat,
            };

            const created = await createCustomFood({
                name: customName.trim(),
                kcalPer100g: kcal,
                proteinPer100g: protein,
                carbsPer100g: carbs,
                fatPer100g: fat,
                nutriments,
            });

            await logSelected({ ...created, source: "custom" });
        } catch (e: any) {
            setErr(e?.message ?? "Failed to create custom food");
        } finally {
            setLoading(false);
        }
    }

    const title = meal.charAt(0).toUpperCase() + meal.slice(1);

    return (
        <div className="dash addFoodPage">
            <section className="card cardPad">
                <div className="segRow">
                    <button className={`seg ${tab === "search" ? "isActive" : ""}`} onClick={() => setTab("search")}>Search</button>
                    <button className={`seg ${tab === "barcode" ? "isActive" : ""}`} onClick={() => setTab("barcode")}>Barcode</button>
                    <button className={`seg ${tab === "custom" ? "isActive" : ""}`} onClick={() => setTab("custom")}>Custom</button>
                </div>

                {tab === "custom" && (
                    <>
                        <div className="dlgSection">
                            <div className="dlgLabel">Food</div>
                            <input className="dlgInput" value={customName} onChange={e => setCustomName(e.target.value)} placeholder="Food name" />
                        </div>

                        <div className="dlgSection">
                            <div className="dlgLabel">Nutrition (per 100g)</div>

                            <input className="dlgInput" type="number" value={kcal} onChange={e => setKcal(+e.target.value || 0)} placeholder="Calories (kcal)" />

                            <div className="dlgRow">
                                <input className="dlgInput" type="number" value={protein} onChange={e => setProtein(+e.target.value || 0)} placeholder="Protein (g)" />
                                <input className="dlgInput" type="number" value={carbs} onChange={e => setCarbs(+e.target.value || 0)} placeholder="Carbs (g)" />
                                <input className="dlgInput" type="number" value={fat} onChange={e => setFat(+e.target.value || 0)} placeholder="Fat (g)" />
                            </div>

                            <div className="dlgRow">
                                <input className="dlgInput" type="number" value={sugars} onChange={e => setSugars(+e.target.value || 0)} placeholder="Sugars (g)" />
                                <input className="dlgInput" type="number" value={fiber} onChange={e => setFiber(+e.target.value || 0)} placeholder="Fiber (g)" />
                            </div>

                            <div className="dlgRow">
                                <input className="dlgInput" type="number" value={salt} onChange={e => setSalt(+e.target.value || 0)} placeholder="Salt (g)" />
                                <input className="dlgInput" type="number" value={sodium} onChange={e => setSodium(+e.target.value || 0)} placeholder="Sodium (g)" />
                            </div>

                            <div className="dlgRow">
                                <input className="dlgInput" type="number" value={satFat} onChange={e => setSatFat(+e.target.value || 0)} placeholder="Saturated fat (g)" />
                                <input className="dlgInput" type="number" value={monoFat} onChange={e => setMonoFat(+e.target.value || 0)} placeholder="Monounsaturated fat (g)" />
                                <input className="dlgInput" type="number" value={polyFat} onChange={e => setPolyFat(+e.target.value || 0)} placeholder="Polyunsaturated fat (g)" />
                            </div>

                            <div className="dlgHint">
                                All values are per 100g, same format as Open Food Facts.
                            </div>
                        </div>

                        <div className="dlgActions">
                            <button className="btnSmall btnSmallPrimary" onClick={submitCustom} disabled={loading}>
                                Create and log
                            </button>
                        </div>
                    </>
                )}

                {err && <div className="dlgError">{err}</div>}
            </section>
        </div>
    );
}
