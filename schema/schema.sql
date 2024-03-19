CREATE TABLE IF NOT EXISTS user (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(50) NOT NULL UNIQUE,
    avatar_url VARCHAR(100),
    hashed_password VARCHAR(255),
    is_admin BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS client (
    id CHAR(25) PRIMARY KEY,
    secret_hash CHAR(97) NOT NULL,
    name VARCHAR(50) NOT NULL UNIQUE,
    picture_url VARCHAR(100),
    callback_urls VARCHAR(1000) NOT NULL,
    token_expiration bigint NOT NULL DEFAULT 28800
);

CREATE TABLE IF NOT EXISTS authorization_history (
    user_id BIGINT NOT NULL,
    client_id CHAR(25) NOT NULL,
    authorized_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS session (
    id CHAR(25) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    client_id CHAR(25),
    os VARCHAR(15),
    browser VARCHAR(50),
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE
);