CREATE TABLE IF NOT EXISTS users (
    username TEXT PRIMARY KEY NOT NULL,
    password varchar(150) NOT NULL
);

CREATE TABLE IF NOT EXISTS notes (
    username TEXT references users(username),
    text TEXT
);

-- migrate -path ./database/migrate -database "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" up