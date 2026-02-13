# Basic App Example

A simple demonstration of swagger annotations for a REST API with basic CRUD operations.

## Purpose

This example demonstrates:
- Standard Go struct definitions with JSON tags
- Comprehensive swagger annotations on HTTP handlers
- Various parameter types (path, query, body)
- Multiple HTTP methods (GET, POST, PUT, DELETE)
- Proper documentation structure

## Project Structure

```
examples/basicapp/
├── main.go                 # Main entry point with API info annotations
├── models/                 # Data models
│   ├── user.go            # User model
│   ├── product.go         # Product model
│   └── error.go           # Error response model
└── handlers/              # HTTP handlers with swagger annotations
    ├── user_handler.go    # User CRUD operations
    └── product_handler.go # Product CRUD operations
```

## Building

```bash
cd examples/basicapp
go mod tidy
go build
```

## Running

```bash
./basicapp
```

The server will start on `localhost:8080`.

## API Endpoints

### Users
- `GET /api/v1/users` - List all users (with pagination)
- `GET /api/v1/users/{id}` - Get a specific user
- `POST /api/v1/users` - Create a new user
- `PUT /api/v1/users/{id}` - Update a user
- `DELETE /api/v1/users/{id}` - Delete a user
- `GET /api/v1/users/search` - Search users

### Products
- `GET /api/v1/products` - List all products (with filtering)
- `GET /api/v1/products/{id}` - Get a specific product
- `POST /api/v1/products` - Create a new product
- `PUT /api/v1/products/{id}` - Update a product
- `DELETE /api/v1/products/{id}` - Delete a product

## Swagger Annotations

This example uses standard swagger annotations including:

- `@Summary` - Brief description
- `@Description` - Detailed description
- `@Tags` - Group endpoints
- `@Accept` - Content type consumed
- `@Produce` - Content type produced
- `@Param` - Parameter definitions (path, query, body)
- `@Success` - Success response
- `@Failure` - Error responses
- `@Router` - Route definition

## Notes

This is a minimal example for demonstrating swagger parsing of standard Go structs. For more advanced features like custom fields, see the `customfields` example.
