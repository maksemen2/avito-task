CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash CHAR(60) NOT NULL,
    coins BIGINT NOT NULL DEFAULT 1000 CHECK (coins >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE goods (
    id BIGSERIAL PRIMARY KEY,
    type VARCHAR(255) NOT NULL UNIQUE,
    price BIGINT NOT NULL CHECK (price > 0)
);

CREATE TABLE transactions (
    id BIGSERIAL PRIMARY KEY,
    from_user_id BIGINT NOT NULL,
    to_user_id BIGINT NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_from_user
        FOREIGN KEY (from_user_id) 
        REFERENCES users(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
        
    CONSTRAINT fk_to_user
        FOREIGN KEY (to_user_id) 
        REFERENCES users(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

CREATE TABLE purchases (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    good_id BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_user
        FOREIGN KEY (user_id) 
        REFERENCES users(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
        
    CONSTRAINT fk_good
        FOREIGN KEY (good_id) 
        REFERENCES goods(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

INSERT INTO goods (type, price)
VALUES ('t-shirt', 80),
       ('cup', 20),
       ('book', 50),
       ('pen', 10),
       ('powerbank', 200),
       ('hoody', 300),
       ('umbrella', 200),
       ('socks', 10),
       ('wallet', 50),
       ('pink-hoody', 500)
ON CONFLICT (type) DO NOTHING;

CREATE INDEX idx_transactions_from_user ON transactions(from_user_id);
CREATE INDEX idx_transactions_to_user ON transactions(to_user_id);
CREATE INDEX idx_purchases_user ON purchases(user_id);
CREATE INDEX idx_purchases_good ON purchases(good_id);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_goods_type ON goods(type);