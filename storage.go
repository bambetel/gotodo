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
	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")
	return &postgresStorage{db: db}, nil
}

// TODO all args as strings?
func (s *postgresStorage) GetTodos() ([]*Todo, error) {
	sql := `SELECT id, label, priority FROM todos`
	fmt.Println(sql)

	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := []*Todo{}
	for rows.Next() {
		i := new(Todo)

		if err := rows.Scan(&i.Id, &i.Label, &i.Priority); err != nil {
			return nil, fmt.Errorf("error getting todos: %v", err)
		}
		fmt.Println(i)
		todos = append(todos, i)
	}
	return todos, nil
}

func (s *postgresStorage) GetTodosByTag(tag string) ([]*Todo, error) {
	sql := `SELECT id, label, priority, completed, created, modified, get_todo_tags(id) tags
	    FROM todos WHERE EXISTS (SELECT * FROM tagged
		WHERE tagged.item_id=todos.id AND tagged.tag_id=(SELECT id FROM tags WHERE label=$1))`
	rows, err := s.db.Query(sql, tag)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := []*Todo{}
	for rows.Next() {
		i := new(Todo)

		if err := rows.Scan(&i.Id, &i.Label, &i.Priority, &i.Completed, &i.Created, &i.Modified, &i.Tags); err != nil {
			return nil, fmt.Errorf("error getting todos: %v", err)
		}
		fmt.Println(i)
		todos = append(todos, i)
	}
	return todos, nil
}

func (s *postgresStorage) DeleteTodo(id int) error {
	sql := "DELETE FROM todos WHERE id=$1"
	_, err := s.db.Query(sql, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *postgresStorage) CreateTodo(item *Todo) (*Todo, error) {
	var id int
	sql := "INSERT INTO todos (label, priority) VALUES ($1, $2) RETURNING id"
	err := s.db.QueryRow(sql, item.Label, item.Priority).Scan(&id)
	if err != nil {
		return nil, err
	}
	val := *item
	val.Id = id
	fmt.Printf("Created todo id=%d: %+v", id, val)
	return &val, nil
}

func (s *postgresStorage) UpdateTodo(item *Todo) (*Todo, error) {
	var modified time.Time
	sql := "UPDATE todos SET label=$1, priority=$2, completed=$3 WHERE id=$4 RETURNING modified"
	err := s.db.QueryRow(sql, item.Label, item.Priority, item.Completed, item.Id).Scan(&modified)
	if err != nil {
		return nil, err
	}
	tags := strings.Split(item.Tags, ",")
	ins := []string{}
	for _, t := range tags {
		ins = append(ins, fmt.Sprintf("(%d, get_tag_id('%s'))", item.Id, t))
	}
	if len(ins) > 0 {
		sql = "INSERT INTO tagged (item_id, tag_id) VALUES " + strings.Join(ins, ", ") + " ON CONFLICT DO NOTHING"
		fmt.Println("tag insertion query ", sql)
		_, err := s.db.Query(sql)
		// todo feedback which tags actually added, not critical
		if err != nil {
			fmt.Println("couldn't add some tags but nvm")
		}
	}
	val := *item
	val.Modified = modified
	fmt.Println("Modified todo:", val)
	return &val, nil
}

func (s *postgresStorage) GetTags() ([]string, error) {
	sql := `SELECT string_agg(label, ',') FROM tags`
	var labels string
	err := s.db.QueryRow(sql).Scan(&labels)
	if err != nil {
		log.Printf("error getting tags: %v\n", err)
		return nil, err
	}
	return strings.Split(labels, ","), err
}

func (s *postgresStorage) CreateTag(label string) error {
	sql := `INSERT INTO tags (label) VALUES ($1) RETURNING id`
	return s.db.QueryRow(sql, label).Err()
}

func (s *postgresStorage) AddItemTag(item_id int, label string) error {
	sql := `INSERT INTO tagged (item_id, tag_id) VALUES ($1, (SELECT id FROM tags WHERE label=$2)) ON CONFLICT DO NOTHING`
	err := s.db.QueryRow(sql, item_id, label).Err()
	log.Println(err)
	return err
}

func (s *postgresStorage) RmItemTag(item_id int, label string) error {
	sql := `DELETE FROM tagged WHERE item_id=$1 AND tag_id=(SELECT id FROM tags WHERE label=$2)`
	err := s.db.QueryRow(sql, item_id, label).Err()
	log.Println(err)
	return err
}
