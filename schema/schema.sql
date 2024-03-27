CREATE TABLE IF NOT EXISTS user (
    id CHAR(25) PRIMARY KEY,
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
    user_id CHAR(25) NOT NULL,
    client_id CHAR(25) NOT NULL,
    authorized_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, client_id)
);

CREATE TABLE IF NOT EXISTS session (
    id CHAR(25) PRIMARY KEY,
    user_id CHAR(25) NOT NULL,
    client_id CHAR(25),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    os VARCHAR(15),
    browser VARCHAR(50),
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS authorization_code (
    id CHAR(13) PRIMARY KEY,
    user_id CHAR(25) NOT NULL,
    client_id CHAR(25) NOT NULL,
    scopes VARCHAR(50) NOT NULL,
    os VARCHAR(15),
    browser VARCHAR(50),
    expires_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE
);

-- CREATE EVENT delete_expired_sessions
-- ON SCHEDULE EVERY 30 MINUTE
-- STARTS CURRENT_TIMESTAMP
-- DO
--     DELETE FROM session
--     WHERE expires_at <= NOW();
--
-- CREATE TRIGGER set_authorization_code_expiry
-- BEFORE INSERT on authorization_code
-- FOR EACH ROW BEGIN
--     IF new.expires_at IS null THEN
--         SET new.expires_at = DATE_ADD(NOW(), INTERVAL 5 MINUTE);
--     END IF;
-- END;
--
-- CREATE EVENT delete_expired_authorization_code
-- ON SCHEDULE EVERY 1 MINUTE
-- STARTS CURRENT_TIMESTAMP
-- DO
--     DELETE FROM authorization_code
--     WHERE expires_at <= NOW();
