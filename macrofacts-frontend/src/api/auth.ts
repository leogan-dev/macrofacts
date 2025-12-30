import { apiFetch, setToken } from "./client";
import type { LoginResponse } from "./types";

export async function register(username: string, password: string): Promise<void> {
    await apiFetch<void>("/api/auth/register", {
        method: "POST",
        body: JSON.stringify({ username, password }),
    });
}

export async function login(username: string, password: string): Promise<void> {
    const res = await apiFetch<LoginResponse>("/api/auth/login", {
        method: "POST",
        body: JSON.stringify({ username, password }),
    });
    setToken(res.token);
}
