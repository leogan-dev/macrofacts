import { apiFetch } from "./client";
import type { MeSettings, UpdateSettingsRequest } from "./types";

export async function getMeSettings(): Promise<MeSettings> {
    return apiFetch<MeSettings>("/api/me/settings");
}

export async function patchMeSettings(req: UpdateSettingsRequest): Promise<MeSettings> {
    return apiFetch<MeSettings>("/api/me/settings", {
        method: "PATCH",
        body: JSON.stringify(req),
    });
}
