create table public.balance(
    user_id uuid           not null
        primary key,
    balance numeric(16, 6) not null
);

create table public.transaction_type(
    id   smallint     not null
        primary key,
    type varchar(100) not null
);

INSERT INTO public.transaction_type(id, type)
VALUES (1::smallint, 'Деньги зарезервированы с основного баланса'),
       (2::smallint, 'Резервация подтверждена, средства списаны, оплата прошла'),
       (3::smallint, 'Резервация отменена, средства возвращены на основной счет баланса'),
       (4::smallint, 'Добавление средств к балансу');

create table public.transaction(
    id                  uuid default gen_random_uuid() not null
        primary key,
    order_id            uuid,
    user_id             uuid                           not null,
    service_id          uuid,
    transaction_type_id smallint                       not null,
    sum                 numeric(16, 4)                 not null,
    comment             varchar,
    upd_time            timestamp                      not null
);

create index transaction_order_id_user_id_service_id__index
    on public.transaction (order_id, user_id, service_id);

create table public.transaction_upd(
    upd_id              uuid default gen_random_uuid() not null
        primary key,
    id                  uuid                           not null,
    order_id            uuid,
    user_id             uuid                           not null,
    service_id          uuid,
    transaction_type_id smallint                       not null,
    sum                 numeric(16, 4)                 not null,
    comment             varchar,
    upd_time            timestamp                      not null
);

create function public.add_balance(user_id_i uuid, sum_i numeric, comment_i character varying) returns void
    language plpgsql
as
$$
begin
    IF(EXISTS(SELECT 1 FROM public.balance b WHERE b.user_id = user_id_i))THEN
        UPDATE public.balance b SET
            balance = b.balance + sum_i
        WHERE b.user_id = user_id_i;
    ELSE
        INSERT INTO public.balance(user_id, balance)
        VALUES (user_id_i, sum_i);
    END IF;

    INSERT INTO public.transaction(order_id, user_id, service_id, transaction_type_id, sum, comment, upd_time)
    VALUES (null, user_id_i, null, 4::smallint, sum_i, comment_i, CURRENT_TIMESTAMP);
end;
$$;

create function save_transaction(order_id_i uuid, user_id_i uuid, service_id_i uuid, sum_i numeric, transaction_type_id_i smallint, comment_i character varying, OUT status integer) returns integer
    language plpgsql
as
$$
DECLARE
    id_o uuid;
    order_id_o uuid;
    user_id_o uuid;
    service_id_o uuid;
    sum_o numeric;
    transaction_type_id_o smallint;
    upd_time_o timestamp;

    user_id_balance uuid;
    current_balance numeric;
begin
    SELECT id, order_id, user_id, service_id, sum, transaction_type_id, upd_time
    INTO id_o, order_id_o, user_id_o, service_id_o, sum_o, transaction_type_id_o, upd_time_o
    FROM public.transaction
    WHERE order_id = order_id_i AND user_id = user_id_i AND service_id = service_id_i;

    SELECT user_id, balance INTO user_id_balance, current_balance
    FROM public.balance
    WHERE user_id = user_id_i;

    IF(current_balance IS NULL)THEN
           status := 10;
           RETURN;
    END IF;

   IF(transaction_type_id_i = 1)THEN
       IF(transaction_type_id_o IS NOT NULL)THEN
           status := 2;
           RETURN;
       ELSEIF(current_balance < sum_i)THEN
           status := 3;
           RETURN;
        END IF;
   ELSEIF(transaction_type_id_i = 2)THEN
       IF(transaction_type_id_o IS NULL)THEN
           status := 4;
           RETURN;
       ELSEIF(transaction_type_id_o = 2)THEN
           status := 5;
           RETURN;
       ELSEIF(transaction_type_id_o = 3)THEN
           status := 6;
           RETURN;
        END IF;
   ELSEIF(transaction_type_id_i = 3)THEN
       IF(transaction_type_id_o IS NULL)THEN
           status := 7;
           RETURN;
       ELSEIF(transaction_type_id_o = 3)THEN
           status := 8;
           RETURN;
       ELSEIF(transaction_type_id_o = 2)THEN
           status := 9;
           RETURN;
        END IF;
    END IF;

   IF(transaction_type_id_i = 1)THEN
       INSERT INTO public.transaction(order_id, user_id, service_id, transaction_type_id, sum, comment, upd_time)
       VALUES (order_id_i, user_id_i, service_id_i, transaction_type_id_i, sum_i, comment_i, CURRENT_TIMESTAMP);

        UPDATE public.balance SET
            balance = balance - sum_i
        WHERE user_id = user_id_i;
    ELSE
        INSERT INTO public.transaction_upd(id, order_id, user_id, service_id, transaction_type_id, sum, comment, upd_time)
        VALUES (id_o, order_id_o, user_id_o, service_id_o, transaction_type_id_o, sum_o, comment_i, upd_time_o);

        UPDATE public.transaction SET
            transaction_type_id = transaction_type_id_i,
            comment = comment_i,
            upd_time = CURRENT_TIMESTAMP
        WHERE order_id = order_id_i AND user_id = user_id_i AND service_id = service_id_i;

        IF(transaction_type_id_i = 3)THEN
            UPDATE public.balance SET
                balance = balance + sum_o
            WHERE user_id = user_id_i;
        END IF;
    END IF;

   status := 1;
end;
$$;