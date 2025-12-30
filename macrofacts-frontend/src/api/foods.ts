import { apiFetch } from "./client";
import type { FoodSearchResponse, FoodDTO } from "./types";

export async function searchFoods(q: string, limit = 20, cursor?: string | null): Promise<FoodSearchResponse> {
    const params = new URLSearchParams();
    params.set("q", q);
    params.set("limit", String(limit));
    if (cursor) params.set("cursor", cursor);
    return apiFetch<FoodSearchResponse>(`/api/foods/search?${params.toString()}`);
}

export async function getFoodByBarcode(code: string): Promise<FoodDTO> {
    return apiFetch<FoodDTO>(`/api/foods/barcode/${encodeURIComponent(code)}`);
}
