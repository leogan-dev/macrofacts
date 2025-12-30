import { useState } from "react";
import * as Form from "@radix-ui/react-form";
import * as Label from "@radix-ui/react-label";
import { login, register } from "../api/auth";
import { getInitialTheme, toggleTheme, type Theme } from "../theme";

type Mode = "login" | "register";

export function AuthPage({ onAuthed }: { onAuthed: () => void }) {
    const [mode, setMode] = useState<Mode>("login");
    const [loading, setLoading] = useState(false);
    const [err, setErr] = useState<string | null>(null);
    const [theme, setTheme] = useState<Theme>(() => getInitialTheme());

    async function onSubmit(formData: FormData) {
        const username = String(formData.get("username") || "").trim();
        const password = String(formData.get("password") || "");

        setErr(null);
        setLoading(true);

        try {
            if (mode === "register") {
                await register(username, password);
            }
            await login(username, password);
            onAuthed();
        } catch (e: any) {
            setErr(e?.message || "Something went wrong");
        } finally {
            setLoading(false);
        }
    }

    return (
        <div className="shell">
            <div className="authWrap">
                <div className="brand">
                    <div className="brandMark" />
                    <div className="brandName">MacroFacts</div>

                    <button
                        className="themeFab"
                        onClick={() => {
                            toggleTheme();
                            setTheme(getInitialTheme());
                        }}
                        title="Toggle theme"
                    >
                        {theme === "dark" ? "☾" : "☀"}
                    </button>
                </div>

                <div className="card authCard">
                    <div className="authHead">
                        <div className="authTitle">{mode === "login" ? "Welcome back" : "Create account"}</div>
                        <div className="authSub muted">
                            Privacy-first. Username-only. No email.
                        </div>
                    </div>

                    {err ? <div className="errorText">{err}</div> : null}

                    <Form.Root className="form" onSubmit={(e) => {
                        e.preventDefault();
                        void onSubmit(new FormData(e.currentTarget));
                    }}>
                        <Form.Field className="field" name="username">
                            <div className="fieldTop">
                                <Form.Label asChild>
                                    <Label.Root className="label">Username</Label.Root>
                                </Form.Label>
                            </div>
                            <Form.Control asChild>
                                <input className="input" autoComplete="username" placeholder="KEVIN" required />
                            </Form.Control>
                            <div className="hint">We canonicalize usernames to UPPERCASE.</div>
                        </Form.Field>

                        <Form.Field className="field" name="password">
                            <div className="fieldTop">
                                <Form.Label asChild>
                                    <Label.Root className="label">Password</Label.Root>
                                </Form.Label>
                            </div>
                            <Form.Control asChild>
                                <input className="input" type="password" autoComplete={mode === "login" ? "current-password" : "new-password"} required />
                            </Form.Control>
                            <div className="hint">Minimum 8 characters.</div>
                        </Form.Field>

                        <button className="btn" disabled={loading} type="submit">
                            {loading ? "…" : mode === "login" ? "Login" : "Create account"}
                        </button>

                        <div className="switchRow">
                            <button
                                className="linkBtn"
                                type="button"
                                onClick={() => setMode(mode === "login" ? "register" : "login")}
                            >
                                {mode === "login" ? "New here? Create an account" : "Already have an account? Login"}
                            </button>
                        </div>
                    </Form.Root>
                </div>

                <div className="muted footNote">
                    No tracking. No email. No ads. Just nutrition.
                </div>
            </div>
        </div>
    );
}
