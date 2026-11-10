-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    "id" uuid NOT NULL PRIMARY KEY DEFAULT uuidv7(),
    "name" text NOT NULL,
    "email" text UNIQUE NOT NULL,
    "password" text NOT NULL,
    "image" text,
    "createdAt" timestamptz NOT NULL,
    "updatedAt" timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS "sessions" (
	"id" uuid NOT NULL PRIMARY KEY DEFAULT uuidv7(),
	"userId" uuid NOT NULL,
	"token" text NOT NULL UNIQUE,
	"expiresAt" timestamptz NOT NULL,
	"ipAddress" text,
	"userAgent" text,
	"createdAt" timestamptz NOT NULL,
	"updatedAt" timestamptz NOT NULL,
	FOREIGN KEY ("userId") REFERENCES users ("id") ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS "sessions_userId_idx" ON "sessions" ("userId");

-- +goose Down
DROP TABLE IF EXISTS "sessions";
DROP TABLE IF EXISTS users;
