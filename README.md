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

#Swagger Link
http://localhost:8080/api/v1/swagger/index.html


