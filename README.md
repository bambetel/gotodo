# GO TODO JSON API

## Routes

`/todos` - list todos
`/todos?[query]` - filter todos

Query parameters:
- q - full-text search
- l - -,,-, label only
- pmin/pmax - priority range

`/todos/{id}` - get entry by id
`/tags` - list tags
`/tags/{tag}` - get entries with a tag

