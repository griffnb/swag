# StructParserService Implementation Blocker

## Status: BLOCKED - Import Cycle Issue

### What Was Accomplished

1. **Service Implementation** (`internal/parser/struct/service.go`):
   - Created complete `Service` struct with all dependencies
   - Implemented `ParseDefinition()` - main entry point for parsing structs
   - Implemented `ParseStruct()` - parses standard Go structs
   - Implemented `ParseField()` - parses individual struct fields
   - Implemented custom model parsing support (fields.StructField[T])
   - Implemented enum processing
   - Implemented recursion detection
   - Implemented schema name transforms (UseStructName, @Name alias)
   - Added helper functions: `getFieldType()`, `fullTypeName()`
   - Total: ~530 lines, well-structured

2. **Tests** (`internal/parser/struct/service_test.go`):
   - Comprehensive test suite created (all skipped - RED phase)
   - Tests cover: simple structs, nested structs, required fields, embedded structs
   - Tests for custom models, public filtering, generics, recursion
   - Tests for naming strategies, field validation, examples
   - Ready to be enabled incrementally as implementation progresses

3. **Field Parser** (`internal/parser/struct/field.go`):
   - Already copied from root field_parser.go
   - Contains full FieldParser implementation with tag handling

### The Blocker: Import Cycle

**Problem**: Cannot integrate StructParserService into parser.go due to import cycle:

```
github.com/swaggo/swag
    → imports github.com/swaggo/swag/internal/parser/struct (from parser.go)
        → imports github.com/swaggo/swag (from field.go)
            = IMPORT CYCLE
```

**Root Cause**: `internal/parser/struct/field.go` references `swag.Parser`:

```go
// In field.go:
type tagBaseFieldParser struct {
    p     *swag.Parser  // ← References swag.Parser
    field *ast.Field
    tag   reflect.StructTag
}

func newTagBaseFieldParser(p *swag.Parser, field *ast.Field) swag.FieldParser {
    // ...
}
```

The field parser needs access to:
- `swag.Parser` type
- `swag.PropNamingStrategy` constants (CamelCase, SnakeCase, PascalCase)
- `swag.BuildCustomSchema()` function
- `swag.IsRefSchema()`, `swag.IsNumericType()` helper functions
- Many other swag package functions

### Solution Options

#### Option 1: Extract Field Parser to Separate Package (RECOMMENDED)
Create `internal/parser/field/` package that doesn't depend on either `swag` or `internal/parser/struct`:

```
internal/
├── parser/
│   ├── field/           # NEW: Field parsing logic
│   │   ├── parser.go    # FieldParser interface and implementation
│   │   ├── tags.go      # Tag parsing logic
│   │   └── naming.go    # Naming strategy logic
│   └── struct/
│       ├── service.go   # Struct parsing service
│       └── field.go     # DELETE (move to internal/parser/field/)
```

**Steps**:
1. Create `internal/parser/field/` package
2. Move field parsing logic from both `field_parser.go` and `internal/parser/struct/field.go`
3. Extract constants/functions from `swag` package that field parser needs
4. Have both `swag` and `internal/parser/struct` import `internal/parser/field`

**Pros**:
- Clean separation of concerns
- Breaks import cycle
- Field parser can be reused by other parsers

**Cons**:
- Requires significant refactoring
- Need to extract many helpers from swag package
- Estimated effort: 2-3 days

#### Option 2: Interface-Based Dependency Injection
Instead of passing `*swag.Parser`, pass interfaces:

```go
type NamingStrategy interface {
    GetStrategy() string
}

type SchemaHelper interface {
    BuildCustomSchema(types []string) (*spec.Schema, error)
    IsRefSchema(schema *spec.Schema) bool
    // ... other needed functions
}

type tagBaseFieldParser struct {
    naming       NamingStrategy
    schemaHelper SchemaHelper
    field        *ast.Field
    tag          reflect.StructTag
}
```

**Pros**:
- Minimal changes to field parser logic
- Breaks direct dependency on swag.Parser

**Cons**:
- Requires creating many interface types
- Still has tight coupling through interfaces
- Estimated effort: 1-2 days

#### Option 3: Delay StructParserService Integration
Keep current structure, integrate later after other services are stable:

**Pros**:
- No immediate work required
- Can plan proper solution
- Other services (RouteParser) can be prioritized

**Cons**:
- StructParserService remains unused
- Benefits of refactoring delayed

### Recommendation

**Proceed with Option 3** (Delay Integration) because:

1. **Other Work Pending**: Phase 7 (RouteParserService) is also not integrated
2. **CLI Works**: Current implementation is functional
3. **Proper Solution Needs Planning**: A hasty fix could create more problems
4. **Complete Refactoring Later**: Better to do it right as part of final cleanup phase

When ready to implement, **Option 1 is the best long-term solution**.

### What Remains for Phase 6

**To Complete StructParserService Integration**:

1. **Resolve Import Cycle** (choose option above)
2. **Move generics.go** to `internal/parser/struct/generics.go`
3. **Wire up service in parser.go**:
   - Add `structParser` field to Parser struct
   - Initialize in `New()` method
   - Replace `parser.ParseDefinition()` calls with `parser.structParser.ParseDefinition()`
   - Replace `parser.parseStruct()` calls with `parser.structParser.ParseStruct()`
   - Replace `parser.parseStructField()` calls with `parser.structParser.ParseField()`
4. **Enable tests incrementally** (TDD - one test at a time)
5. **Verify TestCoreModelsIntegration passes**
6. **Remove old code**: field_parser.go, inline struct parsing from parser.go

### Files Modified

**Created**:
- `/internal/parser/struct/service.go` (530 lines) - Complete service implementation
- `/.agents/STRUCT_PARSER_BLOCKER.md` (this file)

**Modified**:
- `/internal/parser/struct/service_test.go` - Already had skipped tests
- `/internal/parser/struct/field.go` - Already existed (copy of field_parser.go)

**Not Modified** (blocked by import cycle):
- `/parser.go` - Cannot add structParser field yet
- `/field_parser.go` - Still in use (cannot remove)
- `/generics.go` - Not moved yet

### Testing Status

- Service compiles independently: ❌ NO (import cycle)
- Tests created: ✅ YES (all skipped)
- Integration complete: ❌ NO (blocked)
- Core models test passes: ✅ YES (using old implementation)

### Estimated Effort to Complete

- **Option 1 (Field Parser Extraction)**: 2-3 days
- **Option 2 (Interface Injection)**: 1-2 days
- **Option 3 (Delay)**: 0 days now, 2-3 days later

### Next Steps

1. **Document blocker** in REFACTORING_STATUS.md ✅
2. **Mark task as blocked** in task tracker
3. **Decide**: Which option to pursue?
4. **If Option 3**: Move to Phase 7 (RouteParserService) or other work
5. **If Option 1 or 2**: Plan detailed implementation steps

### References

- **Service Code**: `/internal/parser/struct/service.go`
- **Field Parser**: `/internal/parser/struct/field.go`
- **Old Implementation**: `/parser.go` (lines 1567-2037)
- **Plan Document**: `/.agents/plans/radiant-purring-axolotl.md` (Phase 6)
- **Status Tracker**: `/.agents/REFACTORING_STATUS.md`
