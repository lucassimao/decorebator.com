CREATE TABLE test123(
  name text PRIMARY KEY NOT NULL,
  created_at timestamptz NOT NULL DEFAULT NOW()
);