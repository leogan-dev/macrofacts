export type ApiErrorEnvelope = {
    error?: {
        code?: string;
        message?: string;
        request_id?: string;
        details?: unknown;
    };
};

export type LoginResponse = { token: string };

export type MeResponse = { id: string; username: string };

export type MeSettings = {
    timezone: string;
    calorieGoal: number;
    proteinGoalG: number;
    carbsGoalG: number;
    fatGoalG: number;
};

export type UpdateSettingsRequest = Partial<MeSettings>;

export type FoodSource = "off" | "custom";

export type FoodDTO = {
    id: string;
    source: FoodSource;

    name: string;
    brand?: string | null;
    barcode?: string | null;

    // Optional OFF strings
    servingSize?: string | null;
    quantity?: string | null;

    // Normalized serving size (custom foods store this; OFF may sometimes have it)
    servingG?: number | null;

    // Per 100g core macros (nullable when unknown from OFF)
    kcalPer100g?: number | null;
    proteinPer100g?: number | null;
    carbsPer100g?: number | null;
    fatPer100g?: number | null;

    // Common secondary nutrients per 100g
    fiberPer100g?: number | null;
    sugarPer100g?: number | null;
    saltPer100g?: number | null;

    verified?: boolean;
};

export type FoodSearchResponse = {
    items: FoodDTO[];
    nextCursor?: string | null;
};

export type MacroTotals = {
    calories: number;
    protein_g: number;
    carbs_g: number;
    fat_g: number;
};

export type TodayResponse = {
    date: string;
    summary: {
        calorieGoal: number;
        caloriesConsumed: number;
        macrosGoal: { protein_g: number; carbs_g: number; fat_g: number };
        macrosConsumed: { protein_g: number; carbs_g: number; fat_g: number };
    };
    meals: Array<{
        meal: "breakfast" | "lunch" | "dinner" | "snacks";
        totals: MacroTotals;
        entries: Array<{
            id: string;
            time: string;
            food: {
                name: string;
                brand?: string | null;
                source: FoodSource;
                foodId?: string | null;
                barcode?: string | null;
            };
            quantity_g: number;
            computed: MacroTotals;
        }>;
    }>;
    recentFoods: Array<{
        source: FoodSource;
        foodId?: string | null;
        barcode?: string | null;
        name: string;
        brand?: string | null;
        per100g: MacroTotals;
        serving?: { label: string; grams: number } | null;
    }>;
};

export type CreateLogEntryRequest = {
    meal: "breakfast" | "lunch" | "dinner" | "snacks";
    source: FoodSource;
    foodId?: string;
    barcode?: string;
    quantity_g: number;
};

export type CreateLogEntryResponse = { id: string };
