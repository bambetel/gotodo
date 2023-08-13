package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type storage interface {
	GetTodos() ([]*Todo, error)
	GetTodosByTag(string) ([]*Todo, error)
	CreateTodo(*Todo) (*Todo, error)
	UpdateTodo(*Todo) error
	DeleteTodo(int) error
}

type postgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage() (*postgresStorage, error) {
	connStr := `user=golang password=JWgrep321 dbname=todo sslmode=disable`
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

func (s *postgresStorage) GetTodos() ([]*Todo, error) {
	rows, err := s.db.Query("SELECT id, label, priority, completed, created, modified FROM todos")
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	todos := []*Todo{}
	for rows.Next() {
		i := new(Todo)

		if err := rows.Scan(&i.Id, &i.Label, &i.Priority, &i.Completed, &i.Created, &i.Modified); err != nil {
			return nil, fmt.Errorf("error getting todos: %v", err)
		}
		fmt.Println(i)
		todos = append(todos, i)
	}
	return todos, nil
}

// columns, err := rows.Columns()
// 	if err != nil {
// 		fmt.Println("cannot get column names")
// 	} else {
// 		fmt.Println("the column names are ", columns)
// 	}
//
// 	columnTypes, err := rows.ColumnTypes()
// 	if err != nil {
// 		fmt.Println("cannot get column types")
// 	} else {
// 		fmt.Println("the column types are ", columnTypes)
// 		for _, t := range columnTypes {
// 			fmt.Println(t.DatabaseTypeName())
// 		}
// 	}

func (s *postgresStorage) GetTodosByTag(tag string) ([]*Todo, error) {
	// TODO
	// OR by tag id?
	// select * from todos where exists (select * from tagged where tagged.item_id=todos.id AND tagged.tag_id=3);
	// select * from todos where exists (select * from tagged where tagged.item_id=todos.id AND tagged.tag_id=(select id from tags where label='go'));
	sql := "SELECT * FROM todos WHERE EXISTS (SELECT * FROM tagged WHERE tagged.item_id=todos.id AND tagged.tag_id=(SELECT id FROM tags WHERE label=$1))"
	rows, err := s.db.Query(sql, tag)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	todos := []*Todo{}
	for rows.Next() {
		i := new(Todo)

		if err := rows.Scan(&i.Id, &i.Label, &i.Priority, &i.Completed, &i.Created, &i.Modified); err != nil {
			return nil, fmt.Errorf("error getting todos: %v", err)
		}
		fmt.Println(i)
		todos = append(todos, i)
	}
	return todos, nil
}

func (s *postgresStorage) DeleteTodo(id int) error {
	sql := "DELETE FROM todos WHERE id=$1"
	_, err := s.db.Query(sql)
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
	return &val, nil
}

func (s *postgresStorage) UpdateTodo(item *Todo) (*Todo, error) {
	var modified time.Time
	sql := "UPDATE todos SET label=$1, priority=$2, completed=$3 WHERE id=$4 RETURNING modified"
	err := s.db.QueryRow(sql, item.Label, item.Priority, item.Completed, item.Id).Scan(&modified)
	if err != nil {
		return nil, err
	}
	val := *item
	val.Modified = modified
	return &val, nil
}
