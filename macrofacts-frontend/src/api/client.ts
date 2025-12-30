import type { ApiErrorEnvelope } from "./types";

export function getToken(): string | null {
    return localStorage.getItem("token");
}

export function setToken(token: string) {
    localStorage.setItem("token", token);
}

export function clearToken() {
    localStorage.removeItem("token");
}

export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
    const token = getToken();

    const headers = new Headers(init.headers || {});
    if (init.body && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");
    if (token && !headers.has("Authorization")) headers.set("Authorization", `Bearer ${token}`);

    const res = await fetch(path, { ...init, headers });

    if (!res.ok) {
        const data: ApiErrorEnvelope | null = await res.json().catch(() => null);
        const msg = data?.error?.message || `Request failed (${res.status})`;
        throw new Error(msg);
    }

    return (await res.json()) as T;
}
