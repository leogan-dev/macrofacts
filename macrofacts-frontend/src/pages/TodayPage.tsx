import { useEffect, useMemo, useState } from "react";
import { getToday } from "../api/logs";
import type { TodayResponse, FoodDTO, FoodSource } from "../api/types";
import { getMeSettings, patchMeSettings } from "../api/me";
import { getInitialTheme, toggleTheme, type Theme } from "../theme";
import { AddFoodDialog } from "../ui/AddFoodDialog";

function clamp01(x: number) {
    if (x < 0) return 0;
    if (x > 1) return 1;
    return x;
}

function formatDateLong(isoDate: string) {
    const d = new Date(isoDate + "T00:00:00");
    return d.toLocaleDateString(undefined, { weekday: "long", month: "short", day: "numeric" });
}

export function TodayPage({ onLogout }: { onLogout: () => void }) {
    const [theme, setTheme] = useState<Theme>(() => getInitialTheme());
    const [loading, setLoading] = useState(true);
    const [data, setData] = useState<TodayResponse | null>(null);
    const [err, setErr] = useState<string | null>(null);

    const [addOpen, setAddOpen] = useState(false);
    const [addPrefill, setAddPrefill] = useState<FoodDTO | null>(null);

    async function load() {
        setLoading(true);
        setErr(null);
        try {
            // Ensure timezone captured early (privacy-first: only timezone string)
            const deviceTz = Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
            const s = await getMeSettings().catch(() => null);
            if (!s || s.timezone !== deviceTz) {
                await patchMeSettings({ timezone: deviceTz }).catch(() => {});
            }

            const t = await getToday();
            setData(t);
        } catch (e: any) {
            setErr(e?.message || "Failed to load today");
        } finally {
            setLoading(false);
        }
    }

    useEffect(() => {
        void load();
    }, []);

    const remaining = useMemo(() => {
        if (!data) return 0;
        return data.summary.calorieGoal - data.summary.caloriesConsumed;
    }, [data]);

    const openAdd = (prefill?: {
        source: FoodSource;
        barcode?: string | null;
        foodId?: string | null;
        name: string;
        brand?: string | null;
        per100g: any;
        serving?: any;
    }) => {
        if (!prefill) {
            setAddPrefill(null);
            setAddOpen(true);
            return;
        }

        // Convert recentFood -> FoodDTO for AddFoodDialog
        const f: FoodDTO = {
            id: prefill.foodId || prefill.barcode || "",
            source: prefill.source,
            name: prefill.name,
            brand: prefill.brand ?? null,
            barcode: prefill.barcode ?? null,
            kcalPer100g: prefill.per100g.calories ?? 0,
            proteinPer100g: prefill.per100g.protein_g ?? 0,
            carbsPer100g: prefill.per100g.carbs_g ?? 0,
            fatPer100g: prefill.per100g.fat_g ?? 0,
            servingG: prefill.serving?.grams ?? null,
        };

        setAddPrefill(f);
        setAddOpen(true);
    };

    return (
        <div className="dash">
            <header className="dashTop">
                <div className="dashBrand">
                    <div className="brandMark" />
                    <div>
                        <div className="dashTitle">Today</div>
                        <div className="dashSubtitle">{data ? formatDateLong(data.date) : "—"}</div>
                    </div>
                </div>

                <div className="dashTopRight">
                    <button
                        className="iconBtn"
                        title="Toggle theme"
                        onClick={() => {
                            toggleTheme();
                            setTheme(getInitialTheme());
                        }}
                    >
                        {theme === "dark" ? "☾" : "☀"}
                    </button>
                    <button className="btn ghost" onClick={onLogout}>
                        Logout
                    </button>
                </div>
            </header>

            {loading && (
                <div className="card dashCard">
                    <div className="muted">Loading…</div>
                </div>
            )}

            {!loading && err && (
                <div className="card dashCard">
                    <div className="errorText">{err}</div>
                    <div style={{ marginTop: 10 }}>
                        <button className="btn" onClick={load}>
                            Retry
                        </button>
                    </div>
                </div>
            )}

            {!loading && data && (
                <>
                    <section className="dashGrid">
                        <div className="card dashCard hero">
                            <div className="heroTop">
                                <div className="heroLabel">Calories</div>
                                <div className="heroMeta">
                                    <span className="muted">Consumed</span> <strong>{data.summary.caloriesConsumed}</strong>
                                    <span className="dot">•</span>
                                    <span className="muted">Goal</span> <strong>{data.summary.calorieGoal}</strong>
                                </div>
                            </div>

                            <div className="heroNumber">
                                {remaining >= 0 ? (
                                    <>
                                        <span className="heroBig">{remaining}</span>
                                        <span className="heroUnit">left</span>
                                    </>
                                ) : (
                                    <>
                                        <span className="heroBig">{Math.abs(remaining)}</span>
                                        <span className="heroUnit">over</span>
                                    </>
                                )}
                            </div>
                        </div>

                        <div className="card dashCard macros">
                            <MacroBar label="Protein" value={data.summary.macrosConsumed.protein_g} goal={data.summary.macrosGoal.protein_g} unit="g" />
                            <MacroBar label="Carbs" value={data.summary.macrosConsumed.carbs_g} goal={data.summary.macrosGoal.carbs_g} unit="g" />
                            <MacroBar label="Fat" value={data.summary.macrosConsumed.fat_g} goal={data.summary.macrosGoal.fat_g} unit="g" />
                        </div>
                    </section>

                    <section className="card dashCard actions">
                        <button className="btn" onClick={() => openAdd()}>
                            Add food
                        </button>
                        <button className="btn ghost" onClick={() => alert("Scan barcode is next. For now: Add food search, or /foods/barcode via API.")}>
                            Scan barcode
                        </button>
                        <button className="btn ghost" onClick={() => alert("Create custom food UI is next. API already exists (POST /api/foods).")}>
                            Create custom food
                        </button>
                    </section>

                    <section className="dashMeals">
                        {data.meals.map((m) => (
                            <div key={m.meal} className="card dashCard meal">
                                <div className="mealHead">
                                    <div className="mealTitle">{m.meal.toUpperCase()}</div>
                                    <div className="mealTotals">
                                        <strong>{m.totals.calories}</strong>
                                        <span className="muted">kcal</span>
                                        <span className="dot">•</span>
                                        <span className="muted">
                      P {Math.round(m.totals.protein_g)} / C {Math.round(m.totals.carbs_g)} / F {Math.round(m.totals.fat_g)}
                    </span>
                                    </div>
                                </div>

                                {m.entries.length === 0 ? (
                                    <div className="empty">Nothing logged yet.</div>
                                ) : (
                                    <div className="entries">
                                        {m.entries.map((e) => (
                                            <div key={e.id} className="entryRow">
                                                <div className="entryMain">
                                                    <div className="entryName">
                                                        {e.food.name} {e.food.brand ? <span className="muted">· {e.food.brand}</span> : null}
                                                    </div>
                                                    <div className="entryMeta muted">
                                                        {e.time} · {e.quantity_g}g
                                                    </div>
                                                </div>
                                                <div className="entryNums">
                                                    <div className="entryKcal">{e.computed.calories} kcal</div>
                                                    <div className="entryMacros muted">
                                                        P {Math.round(e.computed.protein_g)} · C {Math.round(e.computed.carbs_g)} · F {Math.round(e.computed.fat_g)}
                                                    </div>
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                )}

                                <div className="mealFoot">
                                    <button
                                        className="btn small"
                                        onClick={() => {
                                            setAddPrefill(null);
                                            setAddOpen(true);
                                        }}
                                    >
                                        + Add
                                    </button>
                                </div>
                            </div>
                        ))}
                    </section>

                    <section className="card dashCard recent">
                        <div className="mealHead">
                            <div className="mealTitle">RECENT</div>
                            <div className="muted">Quick add</div>
                        </div>

                        {data.recentFoods.length === 0 ? (
                            <div className="empty">Log something once and your shortcuts will appear here.</div>
                        ) : (
                            <div className="recentList">
                                {data.recentFoods.map((r) => (
                                    <button key={(r.foodId || r.barcode || r.name) + r.source} className="recentItem" onClick={() => openAdd(r)}>
                                        <div className="recentName">
                                            {r.name} {r.brand ? <span className="muted">· {r.brand}</span> : null}
                                        </div>
                                        <div className="muted">{r.per100g.calories} kcal / 100g</div>
                                    </button>
                                ))}
                            </div>
                        )}
                    </section>

                    <AddFoodDialog
                        open={addOpen}
                        onOpenChange={(v) => setAddOpen(v)}
                        onAdded={() => {
                            setAddOpen(false);
                            setAddPrefill(null);
                            void load();
                        }}
                        prefillFood={addPrefill}
                    />
                </>
            )}
        </div>
    );
}

function MacroBar({ label, value, goal, unit }: { label: string; value: number; goal: number; unit: string }) {
    const pct = goal > 0 ? clamp01(value / goal) : 0;
    return (
        <div className="macroRow">
            <div className="macroTop">
                <div className="macroLabel">{label}</div>
                <div className="macroNums">
                    <strong>{Math.round(value)}</strong>
                    <span className="muted">
            {" "}
                        / {goal}
                        {unit}
          </span>
                </div>
            </div>
            <div className="bar">
                <div className="barFill" style={{ width: `${pct * 100}%` }} />
            </div>
        </div>
    );
}
