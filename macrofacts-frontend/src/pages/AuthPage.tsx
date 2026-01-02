import { useState } from "react";
import * as Form from "@radix-ui/react-form";
import * as Label from "@radix-ui/react-label";
import { login, register } from "../api/auth";
import { getInitialTheme, toggleTheme, type Theme } from "../theme";

type Mode = "login" | "register";

export function AuthPage({ onAuthed }: { onAuthed: () => void }) {
    const [mode, setMode] = useState<Mode>("login");
    const [theme, setTheme] = useState<Theme>(() => getInitialTheme());

    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [busy, setBusy] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const title = mode === "login" ? "Back at it." : "Let’s build momentum.";
    const subtitle =
        mode === "login"
            ? "Log in and keep today’s numbers clean."
            : "Create an account. No email required.";

    async function submit() {
        setError(null);
        setBusy(true);
        try {
            const u = username.trim();

            if (mode === "register") {
                await register(u, password);
            }

            await login(u, password);
            onAuthed();
        } catch (err) {
            setError(err instanceof Error ? err.message : "Something went wrong");
        } finally {
            setBusy(false);
        }
    }

    return (
        <div className="shell">
            {/* Floating theme toggle */}
            <button
                type="button"
                className="themeToggle"
                onClick={() => setTheme((t) => toggleTheme())}
                aria-label="Toggle theme"
                title="Toggle theme"
            >
                {theme === "dark" ? (
                    // Sun icon
                    <svg viewBox="0 0 24 24">
                        <circle cx="12" cy="12" r="5" />
                        <line x1="12" y1="1" x2="12" y2="3" />
                        <line x1="12" y1="21" x2="12" y2="23" />
                        <line x1="4.22" y1="4.22" x2="5.64" y2="5.64" />
                        <line x1="18.36" y1="18.36" x2="19.78" y2="19.78" />
                        <line x1="1" y1="12" x2="3" y2="12" />
                        <line x1="21" y1="12" x2="23" y2="12" />
                        <line x1="4.22" y1="19.78" x2="5.64" y2="18.36" />
                        <line x1="18.36" y1="5.64" x2="19.78" y2="4.22" />
                    </svg>
                ) : (
                    // Moon icon
                    <svg viewBox="0 0 24 24">
                        <path d="M21 12.79A9 9 0 1 1 11.21 3A7 7 0 0 0 21 12.79z" />
                    </svg>
                )}
            </button>

            <div className="authWrap">
                <div className="brand">
                    <div className="brandMark" />
                    <div className="brandName">MacroFacts</div>
                </div>

                <div className="card cardPad">
                    <h1 className="h1">{title}</h1>
                    <p className="sub">{subtitle}</p>

                    <div className="sep" />

                    <Form.Root
                        onSubmit={(e) => {
                            e.preventDefault();
                            void submit();
                        }}
                    >
                        <Form.Field name="username">
                            <Label.Root className="label" htmlFor="username">
                                Username
                            </Label.Root>

                            <Form.Control asChild>
                                <input
                                    id="username"
                                    className="input"
                                    type="text"
                                    autoComplete="username"
                                    placeholder="Your username"
                                    value={username}
                                    onChange={(e) => setUsername(e.target.value)}
                                    minLength={3}
                                    maxLength={32}
                                    required
                                />
                            </Form.Control>

                            <Form.Message match="valueMissing" className="errInline">
                                Pick a username.
                            </Form.Message>
                            <Form.Message match="tooShort" className="errInline">
                                Username must be at least 3 characters.
                            </Form.Message>
                        </Form.Field>

                        <div style={{ height: 14 }} />

                        <Form.Field name="password">
                            <Label.Root className="label" htmlFor="password">
                                Password
                            </Label.Root>

                            <Form.Control asChild>
                                <input
                                    id="password"
                                    className="input"
                                    type="password"
                                    autoComplete={mode === "login" ? "current-password" : "new-password"}
                                    placeholder={mode === "login" ? "Your password" : "Create a strong password"}
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    minLength={10}
                                    required
                                />
                            </Form.Control>

                            <Form.Message match="valueMissing" className="errInline">
                                Add a password.
                            </Form.Message>
                            <Form.Message match="tooShort" className="errInline">
                                Use at least 10 characters.
                            </Form.Message>
                        </Form.Field>

                        {error && <div className="errBox">{error}</div>}

                        <div style={{ height: 16 }} />

                        <button className="btn btnPrimary" type="submit" disabled={busy}>
                            {busy ? "Working…" : mode === "login" ? "Log in" : "Create account"}
                        </button>

                        <div style={{ height: 12 }} />

                        <div className="row">
                            <div style={{ color: "var(--muted)", fontSize: 13 }}>
                                {mode === "login" ? "New here?" : "Already have an account?"}
                            </div>

                            <button
                                type="button"
                                className="linkBtn"
                                onClick={() => {
                                    setError(null);
                                    setMode(mode === "login" ? "register" : "login");
                                }}
                            >
                                {mode === "login" ? "Create account" : "Log in"}
                            </button>
                        </div>
                    </Form.Root>
                </div>

                <div className="note">Privacy by default. No email required.</div>
            </div>
        </div>
    );
}
