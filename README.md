# Elastic REST API with JWT Authentication

This is a simple REST API built with Go (Golang), Elasticsearch, and JWT Authentication. The project demonstrates how to integrate JWT token-based authentication with a RESTful API that interacts with an Elasticsearch instance.

## Features

- **JWT Authentication**: Secure the API with JWT tokens.
- **Elasticsearch Integration**: Perform CRUD operations on Elasticsearch.
- **REST API**: Simple REST endpoints for interacting with the database.
- **Go (Golang)**: Server-side logic implemented in Go.

## Prerequisites

- **Go**: Install Go version 1.18 or later. You can download it from [here](https://go.dev/dl/).
- **Elasticsearch**: Make sure you have Elasticsearch running. You can use Docker to run Elasticsearch locally.

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/aAmer0neee/elastic-rest-jwt.git
```
### 2. Deploy Application

Use docker-compose:

```bash
docker-compose up --build
```

You should get the following response:

```bash
Successfully built 8dfa01956e3c
Successfully tagged elastic-rest-jwt_server:latest
Creating elastic-rest-jwt_elasticsearch_1 ... done
Creating elastic-rest-jwt_server_1        ... done
```

### 3. Create ElasticSearch index
The restaurant database provided in the repository (taken from an open data portal) consists of more than 13 thousand restaurants in the Moscow area

using curl or postman:

```http
PUT http://localhost:8888/create/?name=<index_name>
```

### 4. Test the Endpoints

#### 1. Index Pagination


```http
http://localhost:8888/?page=1
```

#### 2. Retrieve Data from the Database Using API


```http
http://localhost:8888/api/?page=1
```

#### 3. Get JWT Token


```http
http://localhost:8888/api/get_token
```

This endpoint will return a JWT token for authentication.

#### 4. Get Restaurant Recommendations Based on Location

Use **Postman** to add the following header:
Authorization: Bearer <your-jwt-token>

```http
GET http://localhost:8888/api/recommend/?lat=55.674&lon=37.666
```