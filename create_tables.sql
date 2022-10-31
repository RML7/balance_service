create table public.balance(
    user_id uuid not null PRIMARY KEY,
    balance numeric(16, 4) not null
);