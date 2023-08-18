-- get tag id for the following label - create if not exists
CREATE OR REPLACE FUNCTION get_tag_id(tag TEXT) RETURNS INT AS $$
	DECLARE id INT;
BEGIN
	id := ( SELECT tags.id FROM tags WHERE tags.label=LOWER(tag) );
	IF id IS NOT NULL THEN
		RETURN id;
	ELSE 
		INSERT INTO tags (label) VALUES ( tag ) RETURNING tags.id INTO id;
		RETURN id; 
	END IF;
END;
$$ LANGUAGE 'plpgsql';

-- get tags for a todos(id) as comma-separated string
CREATE OR REPLACE FUNCTION get_todo_tags(id INT) RETURNS TEXT AS $$ 
	DECLARE result TEXT;
BEGIN
	result := ( SELECT string_agg(tags.label,',') tags FROM todos
		JOIN tagged ON todos.id=tagged.item_id
		JOIN tags ON tags.id=tagged.tag_id
		WHERE todos.id=$1 GROUP BY todos.id );
	RETURN result;
END;
$$ LANGUAGE 'plpgsql';
