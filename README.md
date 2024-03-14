# GO TODO JSON API

## API

This is a GO TODO list API.

The DBMS used is Postgres and the database implements relational tagging with todos, tagged and tags tables. In API tags are accessed only by labels, any internals are not exposed.

## Endpoints

`/todos` - GET all todos
`/todos?[query]` - filter todos - NIY
`/todos/{id}` - GET, DELETE, update entry by id
`/todos/{id}/tags` - POST add item tag
`/todos/{id}/tags/{tag}` - DELETE item tag
`/tags` - GET list of tags, POST
`/tags/{tag}` - GET entries with a tag

