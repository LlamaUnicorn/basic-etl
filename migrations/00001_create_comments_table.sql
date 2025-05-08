-- +goose Up
CREATE TABLE comments (
                          post_id INTEGER NOT NULL,
                          id INTEGER PRIMARY KEY,
                          name TEXT NOT NULL,
                          email TEXT NOT NULL,
                          body TEXT NOT NULL
);

-- +goose Down
DROP TABLE comments;