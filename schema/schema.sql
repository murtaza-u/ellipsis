CREATE TABLE IF NOT EXISTS user (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(50) NOT NULL UNIQUE,
    avatar_url VARCHAR(100),
    hashed_password VARCHAR(255),
    is_admin BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS client (
    id CHAR(25) PRIMARY KEY,
    secret_hash CHAR(97) NOT NULL,
    name VARCHAR(50) NOT NULL UNIQUE,
    picture_url VARCHAR(100),
    auth_callback_urls VARCHAR(1000) NOT NULL,
    logout_callback_urls VARCHAR(1000) NOT NULL,
    backchannel_logout_url VARCHAR(100),
    token_expiration bigint NOT NULL DEFAULT 28800,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS authorization_history (
    user_id BIGINT NOT NULL,
    client_id CHAR(25) NOT NULL,
    authorized_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, client_id)
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

-- CREATE EVENT delete_expired_sessions
-- ON SCHEDULE EVERY 30 MINUTE
-- STARTS CURRENT_TIMESTAMP
-- DO
--     DELETE FROM session
--     WHERE expires_at <= NOW();
