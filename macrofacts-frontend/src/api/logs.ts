import { apiFetch } from "./client";
import type { CreateLogEntryRequest, CreateLogEntryResponse, TodayResponse } from "./types";

export async function getToday(): Promise<TodayResponse> {
    return apiFetch<TodayResponse>("/api/logs/today");
}

export async function createLogEntry(req: CreateLogEntryRequest): Promise<CreateLogEntryResponse> {
    return apiFetch<CreateLogEntryResponse>("/api/logs/entries", {
        method: "POST",
        body: JSON.stringify(req),
    });
}
