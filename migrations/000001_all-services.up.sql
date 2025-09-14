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
    driver_id INTEGER NOT NULL,
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
    total_price DECIMAL(12, 2) NOT NULL,
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

CREATE TABLE refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at INTEGER NOT NULL
);


INSERT INTO warehouse_stock (product_name, quantity, price, last_updated) VALUES
('Ноутбук ASUS ROG', 15, 899.99, 1757808000),  -- 14 сентября 2025, 00:00:00 UTC
('Смартфон iPhone 15', 8, 999.50, 1757721600),   -- 13 сентября 2025, 00:00:00 UTC
('Наушники Sony WH-1000XM4', 25, 349.99, 1757808000),  -- 14 сентября 2025, 00:00:00 UTC
('Монитор Dell 27"', 12, 459.75, 1757635200),    -- 12 сентября 2025, 00:00:00 UTC
('Клавиатура механическая', 30, 129.99, 1757808000); 

INSERT INTO drivers (name, phone, email, license_number, car, status) VALUES
('Иванов Петр Сергеевич', '+79161234567', 'ivanov@mail.ru', 'AB123456', 'Toyota Camry 2020, гос.номер А123АА777', 'online'),
('Смирнова Анна Викторовна', '+79169876543', 'smirnova@gmail.com', 'CD654321', 'Hyundai Solaris 2019, гос.номер В456ВВ777', 'offline'),
('Козлов Дмитрий Иванович', '+79167778899', 'kozlov@yandex.ru', 'EF789012', 'Kia Rio 2021, гос.номер С789СС777', 'online'),
('Петрова Мария Олеговна', '+79165554433', 'petrova@mail.ru', 'GH345678', 'Volkswagen Polo 2022, гос.номер Е012ЕЕ777', 'online'),
('Сидоров Алексей Николаевич', '+79162223344', 'sidorov@gmail.com', 'IJ901234', 'Skoda Octavia 2020, гос.номер К345КК777', 'offline'),
('Федорова Екатерина Дмитриевна', '+79163334455', 'fedorova@yandex.ru', 'KL567890', 'Renault Logan 2018, гос.номер М678ММ777', 'online'),
('Николаев Артем Валерьевич', '+79164445566', 'nikolaev@mail.ru', 'MN123789', 'Lada Vesta 2021, гос.номер Н901НН777', 'online'),
('Орлова Юлия Сергеевна', '+79165556677', 'orlova@gmail.com', 'OP456012', 'Chevrolet Cruze 2019, гос.номер О234ОО777', 'offline'),
('Белов Максим Андреевич', '+79166667788', 'belov@yandex.ru', 'QR789345', 'Ford Focus 2020, гос.номер П567ПП777', 'online'),
('Алексеева Светлана Игоревна', '+79167778899', 'alekseeva@mail.ru', 'ST012678', 'Nissan Almera 2022, гос.номер Р890РР777', 'online');