CREATE TABLE visited (
    id SERIAL PRIMARY KEY,
    url TEXT UNIQUE
);

CREATE TABLE unvisited (
    id SERIAL PRIMARY KEY,
    url TEXT UNIQUE
);

CREATE TABLE scraped_data (
    id SERIAL PRIMARY KEY,
    url TEXT UNIQUE,
    price TEXT
);
