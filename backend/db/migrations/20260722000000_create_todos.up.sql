CREATE TABLE todos (
  id UUID PRIMARY KEY,
  title TEXT NOT NULL CHECK (btrim(title) <> ''),
  completed BOOLEAN NOT NULL DEFAULT FALSE,
  version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
  due_date TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX todos_due_date_idx ON todos (due_date) WHERE due_date IS NOT NULL;
