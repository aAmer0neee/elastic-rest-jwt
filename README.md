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
### 2. Set Up Elasticsearch
If you don't have Elasticsearch installed, you can run it using Docker. Run the following command to start the Elasticsearch container:

```bash
docker run -d --name elasticsearch \
  -e "discovery.type=single-node" \
  -e "xpack.security.enabled=false" \
  -e "xpack.security.transport.ssl.enabled=false" \
  -p 9200:9200\
  docker.elastic.co/elasticsearch/elasticsearch:8.17.2
```
3. Build the Project
Make sure you have Go installed and set up. Run the following command to build the Go application:

```bash
go build -o app .
```
4. Run the Application

Start the application using the following command:
```bash
./app
By default, the application will run on port 8888. You can change the port in the code or set the PORT environment variable.

5. Test the Endpoints
You can test the API using tools like Postman or curl.

Example requests:
Generate a JWT Token

Request:

http
Копировать
Редактировать
POST /auth/login
This endpoint will return a JWT token for authenticated users.

Get All Items (with JWT authentication)

Request:

http
Копировать
Редактировать
GET /items
Authorization: Bearer <your-jwt-token>
Add a New Item to Elasticsearch

Request:

http
Копировать
Редактировать
POST /items
Authorization: Bearer <your-jwt-token>
Content-Type: application/json
Example cURL Command:
bash
Копировать
Редактировать
curl -X GET http://localhost:8080/items -H "Authorization: Bearer <your-jwt-token>"