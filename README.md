# Elastic REST API with JWT Authentication

This is a simple REST API built with Go (Golang), Elasticsearch, and JWT Authentication. The project demonstrates how to integrate JWT token-based authentication with a RESTful API that interacts with an Elasticsearch instance.

## Features


- **Elasticsearch Integration**: Perform CRUD operations on Elasticsearch.
- **REST API**: Simple REST endpoints for interacting with the database.
- **Go (Golang)**: Server-side logic implemented in Go.
- **JWT Authentication**: Secure the API with JWT tokens.
- **Docker-compose**: For deploying and managing database and server containers.

## Prerequisites

- **Docker**: `28.0.0` and latest
Docker is required to build, deploy, and manage the application and its dependencies in containers.

- **Docker Compose Version**: `3.9` and latest  
Docker-compose is used to define and run multi-container Docker applications. It allows you to define all your services, including the database and server, in a single file (docker-compose.yml).


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

PUT <http://localhost:8888/create/?name=index_name>


### 4. Test the Endpoints

#### 1. Index Pagination


GET <http://localhost:8888/?page=1>


#### 2. Retrieve Data from the Database Using API


GET <http://localhost:8888/api/?page=1>


#### 3. Get JWT Token


GET <http://localhost:8888/api/get_token>


This endpoint will return a JWT token for authentication.

#### 4. Get Restaurant Recommendations Based on Location

Use **Postman** to add the following header:
Authorization: Bearer "your-jwt-token"


GET <http://localhost:8888/api/recommend/?lat=55.674&lon=37.666>
