CREATE DATABASE todo OWNER golang LOCALE 'pl_PL.UTF-8' TEMPLATE template0;

CREATE TABLE todos (
    id SERIAL PRIMARY KEY,
    label TEXT NOT NULL,
    priority INT DEFAULT 3,
    modified TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    -- starting TIMESTAMPTZ,
    -- duration INTERVAL,
    progress zero_to_hundred DEFAULT 0,
    completed BOOLEAN DEFAULT false
);

CREATE DOMAIN zero_to_hundred AS INT
   CHECK ( VALUE >= 0 AND VALUE <= 100);

CREATE OR REPLACE FUNCTION update_modified_todos()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE OR REPLACE TRIGGER update_todos
    BEFORE UPDATE ON todos FOR EACH ROW
        EXECUTE PROCEDURE update_modified_todos();

CREATE TABLE tags (
    id SERIAL PRIMARY KEY,
    label TEXT 
);

CREATE TABLE tagged (
    item_id INT REFERENCES todos(id) ON DELETE CASCADE,
    tag_id INT REFERENCES tags(id) ON DELETE CASCADE
);
