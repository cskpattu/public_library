# Public Library REST API

A minimal web-based REST API built in **Go** for managing a fictional public library's book collection.

## Database Initialization
`CREATE TABLE IF NOT EXISTS books (
id SERIAL PRIMARY KEY,
title TEXT,
author TEXT,
isbn TEXT
);
`


## Test with curl
# Create
curl -X POST -H "Content-Type: application/json" -d '{"title":"Dune","author":"Frank Herbert","isbn":"9780441172719"}' http://localhost:8080/books

# Update
curl -X PUT -H "Content-Type: application/json" -d '{"title":"Dune Messiah","author":"Frank Herbert","isbn":"9780441172696"}' http://localhost:8080/books/1

# Delete
curl -X DELETE http://localhost:8080/books/1

#Swagger Link
http://localhost:8080/api/v1/swagger/index.html


