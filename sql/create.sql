-- CREATE DATABASE todo OWNER golang LOCALE 'pl_PL.UTF-8' TEMPLATE template0;

CREATE TABLE todos (
    id SERIAL PRIMARY KEY,
    label TEXT NOT NULL,
    priority INT DEFAULT 3, -- todo limit range 1..10 asc/desc?
    modified TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    -- starting TIMESTAMPTZ,
    -- duration INTERVAL,
	-- deadline (?)
    progress zero_to_hundred DEFAULT 0,
    completed BOOLEAN DEFAULT false,
	-- completed_at (?)
	ts_index TSVECTOR GENERATED ALWAYS AS (to_tsvector('english', label)) STORED
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
    label TEXT UNIQUE 
);

CREATE TABLE tagged (
    item_id INT REFERENCES todos(id) ON DELETE CASCADE,
    tag_id INT REFERENCES tags(id) ON DELETE CASCADE,
	UNIQUE(item_id, tag_id) ON CONFLICT DO NOTHING
);

CREATE OR REPLACE FUNCTION update_tags_check()
RETURNS TRIGGER AS $$
BEGIN
	IF EXISTS (SELECT FROM tags WHERE label=LOWER(NEW.label))
		RETURN NULL;
	END IF;
    NEW.label=LOWER(NEW.label);
	RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

CREATE OR REPLACE TRIGGER update_tags
	BEFORE INSERT OR UPDATE ON tags FOR EACH ROW
		EXECUTE PROCEDURE update_tags_check();
