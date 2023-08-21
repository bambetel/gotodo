/todos 
    GET - get all todos
    POST - create a new todo
/todos/{id}
    PUT - update by id
    DELETE - delete by id 

/todos?search=query - full text search
/todos?sort=created&dir=desc - default asc, columns: updated, created, priority ...
/todos?tag=atag -- tag or tags (?)
TODO /todos?tags=atag,btag
TODO /todos?nottags=ctag - exclude tags (?)
TODO /todos?minpriority=5&maxpriority=10
/todos?done=true OR false - default BOTH
