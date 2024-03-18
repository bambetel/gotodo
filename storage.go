package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type postgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(connStr string) (*postgresStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		log.Fatalf("DB ping error! %v", err)
	}
	log.Println("Connected!")
	return &postgresStorage{db: db}, nil
}

func (s *postgresStorage) GetTodos() ([]*Todo, error) {
	q := `SELECT id, label, priority, completed, created_at, modified_at,
			(SELECT coalesce(string_agg(label,','), '') FROM
				tags JOIN tagged ON tag_id=tags.id AND item_id=todos.id) tags
			FROM todos`
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos, err := scanTodos(rows)
	return todos, err
}

func (s *postgresStorage) QueryTodos(tq *TodoQuery) ([]*Todo, error) {
	cond, args := tq.SQL()
	q := `SELECT id, label, priority, completed, created_at, modified_at,
			(SELECT coalesce(string_agg(label,','), '') FROM
				tags JOIN tagged ON tag_id=tags.id AND item_id=todos.id) tags
			FROM todos`
	if len(cond) > 1 {
		q = q + " WHERE " + cond
	}
	log.Println(q)
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos, err := scanTodos(rows)
	return todos, err
}

func (s *postgresStorage) GetTodosByTag(tag string) ([]*Todo, error) {
	q := `SELECT id, label, priority, completed, created_at, modified_at, 
			(SELECT coalesce(string_agg(label,','), '') FROM
				tags JOIN tagged ON tag_id=tags.id AND item_id=todos.id) tags
			FROM todos WHERE EXISTS (SELECT * FROM tagged
			WHERE tagged.item_id=todos.id AND tagged.tag_id=(SELECT id FROM tags WHERE label=$1))`
	rows, err := s.db.Query(q, tag)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos, err := scanTodos(rows)
	return todos, nil
}

func scanTodos(rows *sql.Rows) ([]*Todo, error) {
	todos := make([]*Todo, 0, 8)
	for rows.Next() {
		i := new(Todo)

		if err := rows.Scan(&i.Id, &i.Label, &i.Priority, &i.Completed, &i.Created, &i.Modified, &i.Tags); err != nil {
			return nil, fmt.Errorf("error scanning todos: %s", err.Error())
		}
		todos = append(todos, i)
	}
	return todos, nil
}

func (s *postgresStorage) DeleteTodo(id int) error {
	q := "DELETE FROM todos WHERE id=$1"
	_, err := s.db.Query(q, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *postgresStorage) CreateTodo(item *Todo) (*Todo, error) {
	var id *int
	var modified, created time.Time
	q := "INSERT INTO todos (label, priority) VALUES ($1, $2) RETURNING id, modified_at, created_at"
	err := s.db.QueryRow(q, item.Label, item.Priority).Scan(&id, &modified, &created)
	if err != nil {
		return nil, err
	}
	val := *item
	val.Modified, val.Created = modified, created
	val.Id = *id
	log.Printf("Created todo id=%d: %+v", id, val)
	return &val, nil
}

// TODO: no tag handling?
func (s *postgresStorage) UpdateTodo(item *Todo) (*Todo, error) {
	var modified, created time.Time
	q := "UPDATE todos SET label=$1, priority=$2, completed=$3 WHERE id=$4 RETURNING modified_at, created_at"
	err := s.db.QueryRow(q, item.Label, item.Priority, item.Completed, item.Id).Scan(&modified, &created)
	if err != nil {
		log.Println("Error:", err.Error())
		return nil, err
	}
	// tags := strings.Split(item.Tags, ",")
	// ins := []string{}
	// for _, t := range tags {
	// 	ins = append(ins, fmt.Sprintf("(%d, get_tag_id('%s'))", item.Id, t))
	// }
	// if len(ins) > 0 {
	// 	q = "INSERT INTO tagged (item_id, tag_id) VALUES " + strings.Join(ins, ", ") + " ON CONFLICT DO NOTHING"
	// 	fmt.Println("tag insertion query ", q)
	// 	_, err := s.db.Query(q)
	// 	// todo feedback which tags actually added, not critical
	// 	if err != nil {
	// 		fmt.Println("couldn't add some tags but nvm")
	// 	}
	// }
	val := *item
	val.Modified, val.Created = modified, created
	return &val, nil
}

type PatchOptions struct {
	Completed bool
}

func (s *postgresStorage) PatchTodo(item *Todo, opts PatchOptions) error {
	// TODO: other fields
	if !opts.Completed {
		return nil
	}
	q := "UPDATE todos SET completed=$2 WHERE id=$1"
	err := s.db.QueryRow(q, item.Id, item.Completed).Err()
	return err
}

func (s *postgresStorage) GetTags() ([]string, error) {
	q := `SELECT string_agg(label, ',') FROM tags`
	var labels sql.NullString
	err := s.db.QueryRow(q).Scan(&labels)
	if err != nil {
		log.Printf("error getting tags: %v\n", err)
		return nil, err
	}
	if labels.Valid {
		return strings.Split(labels.String, ","), nil
	} else {
		return []string{}, nil
	}
}

func (s *postgresStorage) CreateTag(label string) error {
	q := `INSERT INTO tags (label) VALUES ($1) ON CONFLICT DO NOTHING`
	return s.db.QueryRow(q, label).Err()
}

func (s *postgresStorage) AddItemTag(item_id int, label string) error {
	q := `INSERT INTO tagged (item_id, tag_id) VALUES ($1, (SELECT id FROM tags WHERE label=$2)) ON CONFLICT DO NOTHING`
	err := s.db.QueryRow(q, item_id, label).Err()
	log.Println(err)
	return err
}

func (s *postgresStorage) RmItemTag(item_id int, label string) error {
	q := `DELETE FROM tagged WHERE item_id=$1 AND tag_id=(SELECT id FROM tags WHERE label=$2)`
	err := s.db.QueryRow(q, item_id, label).Err()
	log.Println(err)
	return err
}
