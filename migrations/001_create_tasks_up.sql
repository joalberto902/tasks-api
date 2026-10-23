CREATE TABLE tasks(
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    title TEXT NOT NULL,
    status TEXT NOT NULL
        CHECK (status IN ('todo', 'ongoing', 'done')),
    description TEXT NOT NULL DEFAULT ''
);
