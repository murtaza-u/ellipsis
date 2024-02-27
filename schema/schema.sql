CREATE TABLE IF NOT EXISTS user (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(50) NOT NULL UNIQUE,
    avatar_url VARCHAR(100),
    hashed_password VARCHAR(255),
    is_admin BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS client_key (
    client_id CHAR(25) PRIMARY KEY,
    client_secret CHAR(97) NOT NULL
);
