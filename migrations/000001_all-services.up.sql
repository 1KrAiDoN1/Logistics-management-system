CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    time_of_registration INTEGER NOT NULL
);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_time_of_registration ON users(time_of_registration);

CREATE TABLE drivers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    email VARCHAR(255),
    license_number VARCHAR(100) NOT NULL UNIQUE,
    car TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'offline'
);
CREATE INDEX idx_drivers_status ON drivers(status);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    delivery_address TEXT NOT NULL,
    total_amount DECIMAL(12, 2) NOT NULL,
    created_at INTEGER NOT NULL
);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);

CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    quantity INTEGER NOT NULL,
    last_updated INTEGER NOT NULL
);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);
CREATE INDEX idx_order_items_price_quantity ON order_items(price, quantity);

CREATE TABLE warehouse_stock (
    product_id SERIAL PRIMARY KEY,
    product_name VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 0,
    price DECIMAL(10, 2) NOT NULL,
    last_updated INTEGER NOT NULL
);
CREATE INDEX idx_warehouse_stock_name_quantity ON warehouse_stock(product_name, quantity);
CREATE INDEX idx_warehouse_stock_quantity_price ON warehouse_stock(quantity, price);