-- Table with various column types and constraints
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS accounts (
    id INTEGER PRIMARY KEY,
    user_id INTEGER,
    balance REAL DEFAULT 0,
    status TEXT DEFAULT 'active',
    FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS logs (
    id INTEGER PRIMARY KEY,
    event TEXT,
    ts DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS misc (
    id INTEGER PRIMARY KEY,
    flag BOOLEAN,
    data BLOB,
    value NUMERIC
);

-- Table with CHECK constraint
CREATE TABLE IF NOT EXISTS products (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    price REAL CHECK(price >= 0),
    in_stock BOOLEAN DEFAULT 1
);

-- Table with default values and various types
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Table for join/fuzzing
CREATE TABLE IF NOT EXISTS orders (
    id INTEGER PRIMARY KEY,
    user_id INTEGER,
    product_id INTEGER,
    quantity INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id),
    FOREIGN KEY(product_id) REFERENCES products(id)
);

-- Indexes for join/fuzzing
CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_logs_event ON logs(event);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_product_id ON orders(product_id);

-- View example
CREATE VIEW IF NOT EXISTS user_balances AS
    SELECT u.id, u.name, a.balance
    FROM users u
    LEFT JOIN accounts a ON u.id = a.user_id;

-- Insert some initial data
INSERT INTO users (name, email) VALUES ('Alice', 'alice@example.com');
INSERT INTO users (name, email) VALUES ('Bob', 'bob@example.com');
INSERT INTO products (name, price) VALUES ('Widget', 9.99);
INSERT INTO products (name, price) VALUES ('Gadget', 19.99);
INSERT INTO accounts (user_id, balance) VALUES (1, 100.0);
INSERT INTO accounts (user_id, balance) VALUES (2, 50.0);
INSERT INTO orders (user_id, product_id, quantity) VALUES (1, 1, 2);
INSERT INTO orders (user_id, product_id, quantity) VALUES (2, 2, 1);
