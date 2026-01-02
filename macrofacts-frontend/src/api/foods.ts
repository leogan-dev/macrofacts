// src/api/foods.ts
import { apiFetch } from "./client";
import type { FoodSearchResponse, FoodDTO, CreateCustomFoodRequest } from "./types";

type ItemResponse = { item: FoodDTO | null };

export async function searchFoods(q: string, limit = 20, cursor?: string | null): Promise<FoodSearchResponse> {
    const params = new URLSearchParams();
    params.set("q", q);
    params.set("limit", String(limit));
    if (cursor) params.set("cursor", cursor);
    return apiFetch<FoodSearchResponse>(`/api/foods/search?${params.toString()}`);
}

export async function getFoodByBarcode(code: string): Promise<FoodDTO> {
    const res = await apiFetch<ItemResponse>(`/api/foods/barcode/${encodeURIComponent(code)}`);
    if (!res.item) throw new Error("Not found");
    return res.item;
}

export async function createCustomFood(req: CreateCustomFoodRequest): Promise<FoodDTO> {
    const res = await apiFetch<ItemResponse>(`/api/foods/custom`, {
        method: "POST",
        body: JSON.stringify(req),
    });
    if (!res.item) throw new Error("Create failed");
    return res.item;
}
