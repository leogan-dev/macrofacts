export type Theme = "light" | "dark";

const STORAGE_KEY = "theme";

function getStoredTheme(): Theme | null {
    const t = localStorage.getItem(STORAGE_KEY);
    return t === "light" || t === "dark" ? t : null;
}

export function getInitialTheme(): Theme {
    const stored = getStoredTheme();
    if (stored) return stored;

    if (window.matchMedia?.("(prefers-color-scheme: dark)").matches) {
        return "dark";
    }
    return "light";
}

function applyTheme(theme: Theme) {
    document.documentElement.setAttribute("data-theme", theme);
    localStorage.setItem(STORAGE_KEY, theme);
}

export function toggleTheme(): Theme {
    const current = getStoredTheme() ?? getInitialTheme();
    const next: Theme = current === "dark" ? "light" : "dark";
    applyTheme(next);
    return next;
}

// Apply immediately on load (safe to call multiple times)
applyTheme(getInitialTheme());
