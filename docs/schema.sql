-- SQL dump generated using DBML (dbml-lang.org)
-- Database: PostgreSQL
-- Generated at: 2024-02-20T18:16:49.260Z

CREATE TABLE "users" (
  "id" UUID PRIMARY KEY,
  "email" VARCHAR UNIQUE NOT NULL,
  "roles" VARCHAR[] NOT NULL,
  "hashed_password" BYTEA NOT NULL,
  "verified" BOOL NOT NULL DEFAULT false,
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT (now()),
  "updated_at" TIMESTAMPTZ
);

CREATE TABLE "sessions" (
  "id" UUID PRIMARY KEY,
  "user_id" UUID NOT NULL,
  "refresh_token" VARCHAR NOT NULL,
  "user_agent" VARCHAR NOT NULL,
  "client_ip" VARCHAR NOT NULL,
  "is_blocked" BOOLEAN NOT NULL DEFAULT false,
  "expires_at" TIMESTAMPTZ NOT NULL,
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT (now())
);

CREATE TABLE "verifications" (
  "id" UUID PRIMARY KEY,
  "user_id" UUID NOT NULL,
  "email" VARCHAR NOT NULL,
  "code" VARCHAR NOT NULL,
  "used" BOOL NOT NULL DEFAULT false,
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT (now()),
  "expired_at" TIMESTAMPTZ NOT NULL DEFAULT (now()+interval'15 minutes')
);

CREATE TABLE "topics" (
  "id" UUID PRIMARY KEY,
  "name" VARCHAR NOT NULL,
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT (now())
);

CREATE TABLE "notifications" (
  "id" UUID PRIMARY KEY,
  "topic_id" UUID NOT NULL,
  "message" TEXT,
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT (now())
);

CREATE TABLE "subscriptions" (
  "id" UUID PRIMARY KEY,
  "topic_id" UUID NOT NULL,
  "user_id" UUID NOT NULL,
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT (now())
);

CREATE UNIQUE INDEX "unique_sub_topic_user" ON "subscriptions" ("topic_id", "user_id");

ALTER TABLE "sessions" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE;

ALTER TABLE "verifications" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE;

ALTER TABLE "notifications" ADD FOREIGN KEY ("topic_id") REFERENCES "topics" ("id") ON DELETE CASCADE;

ALTER TABLE "subscriptions" ADD FOREIGN KEY ("topic_id") REFERENCES "topics" ("id") ON DELETE CASCADE;

ALTER TABLE "subscriptions" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE;
