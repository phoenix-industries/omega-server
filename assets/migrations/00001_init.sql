-- +goose Up

CREATE TABLE media (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	type TEXT NOT NULL,
	hash TEXT NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX media_hash_idx ON media(hash);

CREATE TYPE user_role AS ENUM ('root', 'admin', 'manager', 'member');
CREATE TABLE users (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE,
	phone TEXT NOT NULL UNIQUE,
	role user_role NOT NULL DEFAULT 'member',
	picture_id TEXT REFERENCES media(id),
	gender TEXT NOT NULL,
	city TEXT,
	governorate TEXT,
	address TEXT,
	password TEXT NOT NULL,
	birthdate TIMESTAMP WITH TIME ZONE NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX users_email_idx ON users (email);
CREATE INDEX users_phone_idx ON users (phone);

CREATE TABLE user_sessions (
	id TEXT NOT NULL PRIMARY KEY,
	user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
	token TEXT NOT NULL,
	ip_address TEXT NOT NULL,
	user_agent TEXT NOT NULL,
	expires_at TIMESTAMP with time zone NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX user_sessions_token_idx ON user_sessions(token);

-- +goose Down
DROP TABLE media;
DROP INDEX media_hash_idx;

DROP INDEX users_email_idx;
DROP INDEX users_phone_idx;
DROP TABLE users;
DROP TYPE user_role;

DROP TABLE user_sessions;
DROP INDEX user_sessions_token_idx;
