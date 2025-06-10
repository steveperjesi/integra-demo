# Demo User App

This is a demo user management API built with Go, Echo, and PostgreSQL. It provides RESTful endpoints to create, read, update, and delete user records. Swagger UI is included for easy API exploration and testing.

## 🚀 Features

- RESTful API using Echo
- Swagger (OpenAPI) documentation
- PostgreSQL database
- Fully containerized with Docker Compose
- Ginkgo unit tests with mocks

## 📦 Prerequisites

- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)

## 🔧 Running the App

To start the application and database:

```bash
docker-compose up --build -d
```

This will:

- Build the Go application image
- Start the app and PostgreSQL service
- Expose the app on http://localhost:8080

## 📖 Accessing the Swagger UI

Once the app is running, you can access the API docs via:

➡️ http://localhost:8080/swagger/index.html

This page includes all available endpoints, request/response formats, and allows you to test the API directly from the browser.

🧪 Running Tests
From your local machine (outside Docker):

```bash
go test ./... -cover
```
or

```bash
ginkgo -r -cover
```

## 📬 API Endpoints

Common endpoints include:

- GET /users
- GET /users/:user_id
- POST /users
- PUT /users
- DELETE /users/:user_id

For full details, see the [Swagger UI](http://localhost:8080/swagger/index.html).