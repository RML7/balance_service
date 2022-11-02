* [Запуск сервера](#запуск-сервера)
* [Описание API](#описание-API)
* [Структура БД](#структура-БД)
* [Используемые сторонние библиотеки](#используемые-сторонние-библиотеки)

# Запуск сервера

1. Запустить базу в докере 
```text
docker-compose up
```

2. Подключиться к базе (креды в docker-compose.yml). Заполнить базу выполнив последовательно скрипты из [sql файла](init_db.sql)
3. Подтянуть необходимые зависимости
4. Запустить main функцию в [файле](cmd/main.go)

P.S. я удалял локально папку проекта и базу из докера, заново клонировал и запускал базу, все должно работать.

# Описание API
Доступно по ссылке:
```text
http://localhost:8000/swagger/index.html
```

![img.png](resource/image/img.png)

# Структура БД
Таблица **balance** - хранение актуального баланса пользователя
```sql
create table public.balance(
user_id uuid           not null
   primary key,
balance numeric(16, 6) not null
);
```

Таблица **transaction_type** - классификатор типов транзакций
```sql
create table public.transaction_type(
id   smallint     not null
    primary key,
type varchar(100) not null
);
```

Таблица **transaction** - хранение транзакций проведенных с балансом
```sql
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
```

Таблица **transaction_upd** - хранение истории изменений статуса транзакци1
```sql
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
```


# Используемые сторонние библиотеки
1. [gorilla/mux](https://github.com/gorilla/mux) - http - роутер
2. [logrus](https://github.com/sirupsen/logrus) - логирование
3. [swaggo/swag](https://github.com/swaggo/swag) - генерация swagger через аннотации
4. [go-playground/validator](https://github.com/go-playground/validator) - валидация
5. [jackc/pgx](https://github.com/jackc/pgx) - работа с бд

