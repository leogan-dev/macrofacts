-- schema.sql

create extension if not exists pgcrypto;

create table if not exists users (
                                     id uuid primary key default gen_random_uuid(),
    username text not null unique,
    password_hash text not null,
    created_at timestamptz not null default now()
    );

-- Global custom foods (shared by everyone)
create table if not exists foods_custom (
                                            id uuid primary key default gen_random_uuid(),
    created_by_user_id uuid not null references users(id) on delete cascade,

    name text not null,
    brand text null,

    -- optional, but if present must be unique (global)
    barcode text null,

    -- Nutrition per 100g (keep it simple + consistent)
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

-- unique barcode if provided
create unique index if not exists foods_custom_barcode_uq
    on foods_custom (barcode)
    where barcode is not null;

-- Fast name search (ILIKE on lowercase name)
create index if not exists foods_custom_name_lower_idx
    on foods_custom (lower(name));

create index if not exists foods_custom_created_at_idx
    on foods_custom (created_at desc);
