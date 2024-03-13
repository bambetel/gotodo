-- CREATE DATABASE todo OWNER golang LOCALE 'pl_PL.UTF-8' TEMPLATE template0;

CREATE TABLE todos (
    id SERIAL PRIMARY KEY,
    label TEXT NOT NULL,
    priority INT DEFAULT 3, -- todo limit range 1..10 asc/desc?
    modified_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    -- starting TIMESTAMPTZ,
    -- duration INTERVAL,
	-- deadline (?)
    progress INT DEFAULT 0,
    completed BOOLEAN DEFAULT false,
	-- completed_at TIMESTAMPTZ DEFAULT NULL,
	ts_index TSVECTOR GENERATED ALWAYS AS (to_tsvector('english', label)) STORED
);

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
    label TEXT UNIQUE 
);

CREATE TABLE tagged (
    item_id INT REFERENCES todos(id) ON DELETE CASCADE,
    tag_id INT REFERENCES tags(id) ON DELETE CASCADE,
	UNIQUE(item_id, tag_id) 
);
