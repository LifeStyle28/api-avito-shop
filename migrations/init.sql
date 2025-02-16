CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    price NUMERIC(10, 2) NOT NULL
);

INSERT INTO products (name, description, price) VALUES
('t-shirt', 'Description', 80),
('cup', 'Description', 20),
('book', 'Description', 50),
('pen', 'Description', 10),
('powerbank', 'Description', 200),
('hoody', 'Description', 300),
('umbrella', 'Description', 200),
('socks', 'Description', 10),
('wallet', 'Description', 50),
('pink-hoody', 'Description', 500);

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    md5 TEXT NOT NULL,
    balance NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    UNIQUE(name, md5)
);
CREATE INDEX IF NOT EXISTS idx_name ON users (name);

CREATE TABLE IF NOT EXISTS inventory (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    purchase_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, product_id)
);
CREATE INDEX IF NOT EXISTS idx_user_id_product_id ON inventory (user_id, product_id);

CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    src INTEGER NOT NULL,
    dst INTEGER NOT NULL,
    amount NUMERIC(10, 2) NOT NULL DEFAULT 0.00
);
CREATE INDEX IF NOT EXISTS idx_src ON transactions (src);
CREATE INDEX IF NOT EXISTS idx_dst ON transactions (dst);
