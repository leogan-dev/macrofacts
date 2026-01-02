-- schema.sql

create extension if not exists pgcrypto;

-- =========================
-- Users
-- =========================
create table if not exists users (
                                     id uuid primary key default gen_random_uuid(),
    username text not null unique,
    password_hash text not null,
    created_at timestamptz not null default now(),

    -- Settings (v1)
    timezone text not null default 'UTC',
    calorie_goal integer not null default 2000,
    protein_goal_g integer not null default 150,
    carbs_goal_g integer not null default 200,
    fat_goal_g integer not null default 70
    );

-- =========================
-- Custom foods
-- =========================
create table if not exists foods_custom (
                                            id uuid primary key default gen_random_uuid(),
    created_by_user_id uuid not null references users(id) on delete cascade,

    name text not null,
    brand text null,
    barcode text null,

    kcal_per_100g numeric not null,
    protein_g_per_100g numeric not null,
    fat_g_per_100g numeric not null,
    carbs_g_per_100g numeric not null,

    fiber_g_per_100g numeric null,
    sugar_g_per_100g numeric null,
    salt_g_per_100g numeric null,

    serving_g numeric null,

    verified boolean not null default false,
    created_at timestamptz not null default now(),

    constraint foods_custom_name_len check (char_length(name) between 1 and 200),
    constraint foods_custom_barcode_len check (barcode is null or char_length(barcode) between 3 and 64),
    constraint foods_custom_macros_nonneg check (
                                                    kcal_per_100g >= 0 and
                                                    protein_g_per_100g >= 0 and
                                                    fat_g_per_100g >= 0 and
                                                    carbs_g_per_100g >= 0 and
(fiber_g_per_100g is null or fiber_g_per_100g >= 0) and
(sugar_g_per_100g is null or sugar_g_per_100g >= 0) and
(salt_g_per_100g is null or salt_g_per_100g >= 0) and
(serving_g is null or serving_g > 0)
    )
    );

create unique index if not exists foods_custom_barcode_uq
    on foods_custom (barcode)
    where barcode is not null;

create index if not exists foods_custom_name_lower_idx
    on foods_custom (lower(name));

create index if not exists foods_custom_created_at_idx
    on foods_custom (created_at desc);

-- =========================
-- Food log entries (snapshot)
-- =========================
create table if not exists food_log_entries (
                                                id uuid primary key default gen_random_uuid(),
    user_id uuid not null references users(id) on delete cascade,

    date date not null,
    meal text not null,

    source text not null, -- 'off' | 'custom'
    food_id uuid null,
    barcode text null,

    food_name text not null,
    brand text null,

    quantity_g integer not null,

    calories integer not null,
    protein_g numeric not null,
    carbs_g numeric not null,
    fat_g numeric not null,

    created_at timestamptz not null default now(),

    constraint food_log_entries_meal_chk check (meal in ('breakfast','lunch','dinner','snacks')),
    constraint food_log_entries_source_chk check (source in ('off','custom')),
    constraint food_log_entries_qty_chk check (quantity_g > 0),
    constraint food_log_entries_cal_chk check (calories >= 0),
    constraint food_log_entries_macros_chk check (protein_g >= 0 and carbs_g >= 0 and fat_g >= 0),
    constraint food_log_entries_ident_chk check (
(source = 'off' and barcode is not null and food_id is null) or
(source = 'custom' and food_id is not null)
    )
    );

create index if not exists food_log_entries_user_date_idx
    on food_log_entries (user_id, date);

create index if not exists food_log_entries_user_date_meal_idx
    on food_log_entries (user_id, date, meal);

create index if not exists food_log_entries_user_created_at_idx
    on food_log_entries (user_id, created_at desc);
