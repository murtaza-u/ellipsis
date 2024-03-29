CREATE TABLE IF NOT EXISTS user (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    avatar_url TEXT,
    hashed_password TEXT,
    is_admin BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS client (
    id TEXT PRIMARY KEY,
    secret_hash TEXT NOT NULL,
    name TEXT NOT NULL UNIQUE,
    picture_url TEXT,
    auth_callback_urls TEXT NOT NULL,
    logout_callback_urls TEXT NOT NULL,
    backchannel_logout_url TEXT,
    token_expiration bigint NOT NULL DEFAULT 28800,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS authorization_history (
    user_id TEXT NOT NULL,
    client_id TEXT NOT NULL,
    authorized_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, client_id)
);

CREATE TABLE IF NOT EXISTS session (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    client_id TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    os TEXT,
    browser TEXT,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS authorization_code (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    client_id TEXT NOT NULL,
    scopes TEXT NOT NULL,
    os TEXT,
    browser TEXT,
    expires_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE
);

-- CREATE EVENT delete_expired_sessions
-- ON SCHEDULE EVERY 30 MINUTE
-- STARTS CURRENT_TIMESTAMP
-- DO
--     DELETE FROM session
--     WHERE expires_at <= CURRENT_TIMESTAMP;
--
-- CREATE TRIGGER set_authorization_code_expiry
-- BEFORE INSERT on authorization_code
-- FOR EACH ROW BEGIN
--     IF new.expires_at IS null THEN
--         SET new.expires_at = DATE_ADD(CURRENT_TIMESTAMP, INTERVAL 5 MINUTE);
--     END IF;
-- END;
--
-- CREATE EVENT delete_expired_authorization_code
-- ON SCHEDULE EVERY 1 MINUTE
-- STARTS CURRENT_TIMESTAMP
-- DO
--     DELETE FROM authorization_code
--     WHERE expires_at <= CURRENT_TIMESTAMP;
