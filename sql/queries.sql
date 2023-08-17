SELECT * FROM todos WHERE ts_index @@ to_tsquery('english', 'postgres | sql');
