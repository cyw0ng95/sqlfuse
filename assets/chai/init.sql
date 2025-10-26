-- SQLite-compatible initial schema for fuzzing (adapted from Postgres)
-- Simple e-commerce-like schema with several tables, constraints and sample data

CREATE TABLE users (
  id INTEGER PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  email TEXT NOT NULL,
  bio TEXT
);