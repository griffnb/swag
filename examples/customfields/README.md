# Custom Fields Example

This example demonstrates Swagger documentation generation with custom field types and public tag filtering.

## Features Demonstrated

### 1. Custom Field Types
- `fields.StructField[T]` - Generic struct fields with type parameters
- `fields.IntConstantField[T]` - Integer constant fields with enum types
- `fields.StringConstantField[T]` - String constant fields with enum types
- `fields.StringField`, `fields.IntField`, `fields.UUIDField`, `fields.TimeField` - Basic typed fields

### 2. Public Tag Filtering
The `public:"view|edit"` struct tag controls field visibility in API documentation:
- `public:"view"` - Field is visible in public API responses (read-only)
- `public:"edit"` - Field is editable in public API requests (read/write)
- No public tag - Field is private and excluded from public API documentation

Example:
```go
type DBColumns struct {
    FirstName      *fields.StringField `public:"edit" column:"first_name"`
    ExternalID     *fields.StringField `public:"view" column:"external_id"`
    HashedPassword *fields.StringField `              column:"hashed_password"` // Private
}
```

### 3. Swagger Ignore
Fields with `swaggerignore:"true"` are completely excluded from Swagger documentation:
```go
Authentication *fields.StructField[*Authentication] `swaggerignore:"true"`
```

### 4. Model Variants
The example demonstrates automatic generation of model variants:
- Base models (e.g., `account.Account`)
- Public variants (e.g., `account.AccountPublic`) - Only includes fields with public tags
- Joined models (e.g., `account.AccountJoined`) - Includes related data

### 5. Complex Nested Structures
- Embedded structs (e.g., `base.Structure` embedded in models)
- Nested StructFields (e.g., `Properties`, `SignupProperties`)
- Map types within StructFields (e.g., `map[types.UUID]map[string]any`)

## Project Structure

```
customfields/
├── account/         # Account models with various field types
├── address/         # Address models
├── api/             # API endpoints demonstrating @Public annotation
├── base/            # Base structures shared across models
├── billing_plan/    # Billing plan models with nested StructFields
├── constants/       # Constant types used in IntConstantField/StringConstantField
├── org_member/      # Organization member models
├── response/        # API response wrappers
├── go.mod
└── main.go          # Entry point with Swagger annotations
```

## Building

```bash
cd examples/customfields
go build
```

## Generating Swagger Documentation

From the swag repository root:

```bash
swag init --dir examples/customfields --generalInfo main.go --output examples/customfields/docs
```

## Key Files to Review

1. **account/account.go** - Demonstrates various field types and public tags
2. **api/api.go** - Shows API endpoints with `@Public` annotation
3. **billing_plan/billing_plan.go** - Complex nested StructField examples
4. **base/structure.go** - Common base fields shared across models

## Dependencies

This example depends on `github.com/griffnb/core/lib` for the field type definitions. In a real project, these would be your custom field types.
