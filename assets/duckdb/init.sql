-- DuckDB initialization schema
-- DuckDB supports standard SQL with some differences from SQLite

-- Table with various column types and constraints
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    name VARCHAR NOT NULL,
    email VARCHAR UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS accounts (
    id INTEGER PRIMARY KEY,
    user_id INTEGER,
    balance DOUBLE DEFAULT 0,
    status VARCHAR DEFAULT 'active',
    FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS logs (
    id INTEGER PRIMARY KEY,
    event VARCHAR,
    ts TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS misc (
    id INTEGER PRIMARY KEY,
    flag BOOLEAN,
    data BLOB,
    value DECIMAL
);

-- Table with CHECK constraint
CREATE TABLE IF NOT EXISTS products (
    id INTEGER PRIMARY KEY,
    name VARCHAR NOT NULL,
    price DOUBLE CHECK(price >= 0),
    in_stock BOOLEAN DEFAULT true
);

-- Table with default values and various types
CREATE TABLE IF NOT EXISTS settings (
    key VARCHAR PRIMARY KEY,
    value VARCHAR,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table for join/fuzzing
CREATE TABLE IF NOT EXISTS orders (
    id INTEGER PRIMARY KEY,
    user_id INTEGER,
    product_id INTEGER,
    quantity INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
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
-- Note: DuckDB doesn't have AUTOINCREMENT, so we provide explicit IDs
INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@example.com');
INSERT INTO users (id, name, email) VALUES (2, 'Bob', 'bob@example.com');
INSERT INTO products (id, name, price) VALUES (1, 'Widget', 9.99);
INSERT INTO products (id, name, price) VALUES (2, 'Gadget', 19.99);
INSERT INTO accounts (id, user_id, balance) VALUES (1, 1, 100.0);
INSERT INTO accounts (id, user_id, balance) VALUES (2, 2, 50.0);
INSERT INTO orders (id, user_id, product_id, quantity) VALUES (1, 1, 1, 2);
INSERT INTO orders (id, user_id, product_id, quantity) VALUES (2, 2, 2, 1);
