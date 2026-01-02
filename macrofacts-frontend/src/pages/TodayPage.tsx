import * as React from "react";
import { getToday } from "../api/logs";
import { getMeSettings } from "../api/me";
import type { TodayResponse, MeSettings } from "../api/types";
import { getInitialTheme, toggleTheme } from "../theme";
import { useNavigate } from "react-router-dom";



type Props = { onLogout: () => void };

type MealKey = "breakfast" | "lunch" | "dinner" | "snacks";
const MEAL_ORDER: MealKey[] = ["breakfast", "lunch", "dinner", "snacks"];

type MacroTotals = {
    calories: number;
    protein_g: number;
    carbs_g: number;
    fat_g: number;
};

type TodayMeal = TodayResponse["meals"][number];

function emptyTotals(): MacroTotals {
    return { calories: 0, protein_g: 0, carbs_g: 0, fat_g: 0 };
}

function safeMeals(resp: TodayResponse | null): TodayMeal[] {
    const fromApi = resp?.meals ?? [];
    const byKey = new Map<string, TodayMeal>();
    for (const m of fromApi) byKey.set(String(m.meal).toLowerCase(), m);

    return MEAL_ORDER.map((k) => {
        const found = byKey.get(k);
        return (
            found ?? {
                meal: k,
                entries: [],
                totals: emptyTotals(),
            }
        ) as TodayMeal;
    });
}

function clamp01(n: number) {
    if (!Number.isFinite(n)) return 0;
    return Math.max(0, Math.min(1, n));
}

export default function TodayPage(props: Props) {
    const [theme, setTheme] = React.useState(() => getInitialTheme());
    const [settings, setSettings] = React.useState<MeSettings | null>(null);
    const [today, setToday] = React.useState<TodayResponse | null>(null);
    const [loading, setLoading] = React.useState(true);
    const [err, setErr] = React.useState<string | null>(null);

    const [pulseRemaining, setPulseRemaining] = React.useState(false);
    const prevRemainingRef = React.useRef<number | null>(null);
    const pulseTimerRef = React.useRef<number | null>(null);

    async function refresh() {
        setLoading(true);
        setErr(null);
        try {
            const [s, t] = await Promise.all([getMeSettings(), getToday()]);
            setSettings(s);
            setToday(t);
        } catch (e: any) {
            setErr(e?.message ?? "Failed to load");
        } finally {
            setLoading(false);
        }
    }

    React.useEffect(() => {
        refresh();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    const dateLabel = React.useMemo(() => {
        const d = new Date();
        return d.toLocaleDateString(undefined, {
            year: "numeric",
            month: "2-digit",
            day: "2-digit",
        });
    }, []);

    const calGoal = settings?.calorieGoal ?? 2000;
    const proteinGoal = settings?.proteinGoalG ?? 150;
    const carbsGoal = settings?.carbsGoalG ?? 200;
    const fatGoal = settings?.fatGoalG ?? 70;

    function computeTotals(t: TodayResponse | null) {
        const totals: MacroTotals = { calories: 0, protein_g: 0, carbs_g: 0, fat_g: 0 };
        if (!t) return totals;

        for (const meal of t.meals ?? []) {
            for (const entry of meal.entries ?? []) {
                const m = entry.computed; // <-- your API shape uses computed
                if (!m) continue;
                totals.calories += m.calories ?? 0;
                totals.protein_g += m.protein_g ?? 0;
                totals.carbs_g += m.carbs_g ?? 0;
                totals.fat_g += m.fat_g ?? 0;
            }
        }
        return totals;
    }

    const totals = React.useMemo(() => computeTotals(today), [today]);
    const calConsumed = Math.max(0, Math.round(totals.calories ?? 0));
    const calRemaining = Math.max(0, Math.round(calGoal - calConsumed));
    const pctConsumed = clamp01(calGoal > 0 ? calConsumed / calGoal : 0);

    // Pulse remaining number when it changes (i.e. logging food)
    React.useEffect(() => {
        if (loading) return;

        const prev = prevRemainingRef.current;
        prevRemainingRef.current = calRemaining;

        // skip initial set
        if (prev === null) return;

        if (prev !== calRemaining) {
            setPulseRemaining(false);
            if (pulseTimerRef.current) window.clearTimeout(pulseTimerRef.current);

            // next frame so animation restarts reliably
            requestAnimationFrame(() => {
                setPulseRemaining(true);
                pulseTimerRef.current = window.setTimeout(() => setPulseRemaining(false), 520);
            });
        }
    }, [calRemaining, loading]);

    React.useEffect(() => {
        return () => {
            if (pulseTimerRef.current) window.clearTimeout(pulseTimerRef.current);
        };
    }, []);

    const protein = totals.protein_g ?? 0;
    const carbs = totals.carbs_g ?? 0;
    const fat = totals.fat_g ?? 0;

    const meals = safeMeals(today);
    const nav = useNavigate();

    return (
        <div className="dash">
            <header className="dashTop">
                <div className="dashTopLeft">
                    <div className="logoMark" />
                    <div>
                        <div className="dashTitle">Today</div>
                        <div className="dashSub">{dateLabel}</div>
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
                        aria-label="Toggle theme"
                    >
                        {theme === "dark" ? "☾" : "☀"}
                    </button>

                    <button className="btn" onClick={props.onLogout}>
                        Logout
                    </button>
                </div>
            </header>

            {err && (
                <div className="card">
                    <div className="muted">{err}</div>
                    <button className="btn" onClick={refresh} style={{ marginTop: 10 }}>
                        Retry
                    </button>
                </div>
            )}

            {!err && loading && (
                <div className="card">
                    <div className="muted">Loading…</div>
                </div>
            )}

            {!err && !loading && (
                <>
                    <section className="card hero">
                        <div className="heroCenter">
                            <CalorieGauge remaining={calRemaining} pct={pctConsumed} />

                            <div className="heroUnder">
                                <div className="heroLabel">Calories remaining</div>

                                <div className="heroChips">
                                    <div className="chip">
                                        <span className="chipLabel">Consumed</span>
                                        <span className="chipValue">{calConsumed}</span>
                                    </div>
                                    <div className="chip">
                                        <span className="chipLabel">Goal</span>
                                        <span className="chipValue">{calGoal}</span>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div className="macroGrid">
                            <MacroBar label="Protein" value={protein} goal={proteinGoal} unit="g" kind="protein" />
                            <MacroBar label="Carbs" value={carbs} goal={carbsGoal} unit="g" kind="carbs" />
                            <MacroBar label="Fat" value={fat} goal={fatGoal} unit="g" kind="fat" />
                        </div>

                        <div className="quickActions">
                            <button className="btnPrimary" disabled>
                                Add food
                            </button>
                            <button className="btn" disabled>
                                Scan barcode
                            </button>
                            <button className="btn" disabled>
                                Create custom food
                            </button>
                        </div>
                    </section>

                    <section className="card">
                        <div className="sectionTitle">Meals</div>

                        <div className="mealGrid">
                            {meals.map((m) => (
                                <MealSection
                                    key={m.meal}
                                    meal={String(m.meal)}
                                    entries={m.entries as any[]}
                                    onAdd={() => nav(`/add-food?meal=${String(m.meal).toLowerCase()}`)}
                                />
                            ))}
                        </div>
                    </section>
                </>
            )}
        </div>
    );
}

function CalorieGauge(props: {
    remaining: number;
    pct: number; // consumed / goal (0..1)
}) {
    const size = 260;
    const stroke = 18;
    const r = (size - stroke) / 2;
    const c = 2 * Math.PI * r;

    const pct = clamp01(props.pct);

    // Animate ring on load + on pct changes
    const [animPct, setAnimPct] = React.useState(0);
    React.useEffect(() => {
        const t = window.setTimeout(() => setAnimPct(pct), 60);
        return () => window.clearTimeout(t);
    }, [pct]);

    // Pulse the number when remaining changes (logging food)
    const [pulse, setPulse] = React.useState(false);
    React.useEffect(() => {
        setPulse(true);
        const t = window.setTimeout(() => setPulse(false), 260);
        return () => window.clearTimeout(t);
    }, [props.remaining]);

    const dash = c * animPct;

    return (
        <div className="gaugeWrap">
            <div className="gauge" aria-label="Calories progress">
                <svg className="gaugeSvg" width={size} height={size} viewBox={`0 0 ${size} ${size}`} role="img">
                    <defs>
                        <linearGradient id="mfGaugeGrad" x1="0" y1="0" x2="1" y2="1">
                            <stop offset="0%" stopColor="var(--dash-accent)" />
                            <stop offset="100%" stopColor="var(--dash-accent-2)" />
                        </linearGradient>
                    </defs>

                    {/* Track */}
                    <circle
                        className="gaugeTrack"
                        cx={size / 2}
                        cy={size / 2}
                        r={r}
                        fill="none"
                        strokeWidth={stroke}
                    />

                    {/* Progress: start at 12 o’clock */}
                    <circle
                        className="gaugeProgress"
                        cx={size / 2}
                        cy={size / 2}
                        r={r}
                        fill="none"
                        stroke="url(#mfGaugeGrad)"
                        strokeWidth={stroke}
                        strokeLinecap="round"
                        strokeDasharray={`${dash} ${c}`}
                        strokeDashoffset={0}
                        transform={`rotate(90 ${size / 2} ${size / 2})`}
                    />
                </svg>

                {/* Big number only */}
                <div className={`gaugeCenter ${pulse ? "isPulse" : ""}`}>
                    <div className="gaugeNumber">{props.remaining}</div>
                </div>
            </div>

            {/* Label + chips BELOW the ring */}
            <div className="gaugeBelow">
                <div className="heroChips heroChipsCenter">
                    {/* keep your existing chips here */}
                </div>
            </div>
        </div>
    );
}


function MacroBar(props: {
    label: string;
    value: number;
    goal: number;
    unit: string;
    kind: "protein" | "carbs" | "fat";
}) {
    const value = Number.isFinite(props.value) ? props.value : 0;
    const goal = props.goal > 0 ? props.goal : 1;
    const pct = Math.max(0, Math.min(1, value / goal));

    return (
        <div className={`macro macro-${props.kind}`}>
            <div className="macroTop">
                <div className="macroLabel">{props.label}</div>
                <div className="macroValue">
                    {Math.round(value)}/{Math.round(goal)}
                    {props.unit}
                </div>
            </div>
            <div className="bar">
                <div className="barFill" style={{ width: `${pct * 100}%` }} />
            </div>
        </div>
    );
}

function MealSection(props: { meal: string; entries: any[]; onAdd: () => void }) {
    const title = props.meal.charAt(0).toUpperCase() + props.meal.slice(1);
    const entries = Array.isArray(props.entries) ? props.entries : [];

    return (
        <div className="meal">
            <div className="mealHead">
                <div className="mealTitle">{title}</div>

                <button className="mealAddBtn" onClick={props.onAdd}>
                    <span className="plus">+</span>
                    Add food
                </button>
            </div>

            {entries.length === 0 ? (
                <div className="muted">Nothing logged yet.</div>
            ) : (
                <div className="mealList">
                    {entries.map((e: any) => {
                        const m = e.computed;
                        return (
                            <div key={e.id} className="mealRow">
                                <div className="mealRowLeft">
                                    <div className="mealFood">{e.food?.name ?? "Food"}</div>
                                    <div className="mutedSmall">{e.quantity_g ?? "—"}g</div>
                                </div>
                                <div className="mealRowRight">
                                    <div className="mealKcal">{Math.round(m?.calories ?? 0)} kcal</div>
                                    <div className="mutedSmall">
                                        P {Math.round(m?.protein_g ?? 0)} • C {Math.round(m?.carbs_g ?? 0)} • F {Math.round(m?.fat_g ?? 0)}
                                    </div>
                                </div>
                            </div>
                        );
                    })}
                </div>
            )}
        </div>
    );
}

