SELECT * FROM todos WHERE ts_index @@ to_tsquery('english', 'postgres | sql');

SELECT todos.id, todos.label, todos.priority, string_agg(tags.label, ',') tags FROM tags, todos WHERE EXISTS (SELECT * FROM tagged WHERE item_id=todos.id AND tagged.tag_id=tags.id) GROUP BY todos.id;

SELECT todos.*, string_agg(tags.label,',') tags FROM todos JOIN tagged ON todos.id=tagged.item_id JOIN tags ON tags.id=tagged.tag_id GROUP BY todos.id;

