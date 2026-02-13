# Swag Project Refactoring Plan

## Context

The swag codebase currently suffers from architectural issues that make it difficult to maintain and test:

- **Monolithic files**: `parser.go` (2,435 lines), `operation.go` (1,314 lines), `packages.go` (788 lines)
- **Mixed responsibilities**: Single files handle multiple concerns (parsing, validation, schema building)
- **Scattered state**: Parser struct has 20+ fields managing different aspects
- **Difficult testing**: Tight coupling makes unit testing individual components hard
- **Embedded test data**: testdata/ cannot be imported by external projects

This refactoring will transform the codebase into a modular architecture with clear separation of concerns, smaller focused files, stateful services, and better testability - all while keeping existing functionality intact.

## Goals

1. **Smaller files**: No file exceeds 500 lines, grouped by related functionality
2. **Clear code flow**: Easy to trace from package loading → parsing → schema building
3. **Better testing**: TDD approach with isolated unit tests for each service
4. **Package separation**: Distinct packages with single responsibilities
5. **Stateful services**: Services own their state to avoid parameter passing hell
6. **Clear domain objects**: Route struct containing body/response, making code self-documenting
7. **Preserve all features**: Custom model parsing, public/private filtering, enums, generics
8. **Importable test data**: Move test data to real Go projects with go.mod

## New Package Structure

```
swag/
├── cmd/                          # CLI commands (unchanged)
├── format/                       # Swagger formatter (already separate)
│
├── internal/                     # Internal packages (not exposed)
│   ├── loader/                   # Package Loading Service
│   │   ├── service.go           # Package discovery and AST collection
│   │   ├── golist.go            # go list integration
│   │   ├── gopackages.go        # go/packages integration
│   │   └── dependency.go        # Dependency resolution
│   │
│   ├── registry/                 # Type & Package Registry
│   │   ├── service.go           # Main registry service
│   │   ├── packages.go          # Package definitions registry
│   │   ├── types.go             # Type spec registry
│   │   ├── enums.go             # Enum registry
│   │   └── lookup.go            # Type lookup utilities
│   │
│   ├── parser/
│   │   ├── base/                # Base Parser (General API Info)
│   │   │   ├── service.go       # @title, @version, @description parser
│   │   │   └── security.go      # Security definitions parser
│   │   │
│   │   ├── struct/              # Struct Parser
│   │   │   ├── service.go       # Struct parsing orchestrator
│   │   │   ├── standard.go      # Standard Go struct parser
│   │   │   ├── custom.go        # fields.StructField[T] parser
│   │   │   ├── field.go         # Field parser with tag handling
│   │   │   └── generics.go      # Generic type handling
│   │   │
│   │   └── route/               # Route Parser
│   │       ├── service.go       # Route parsing service
│   │       ├── operation.go     # Operation parser (@router, etc)
│   │       ├── parameter.go     # @param extraction
│   │       └── response.go      # @success, @failure extraction
│   │
│   ├── schema/                   # Schema Building
│   │   ├── builder.go           # Schema construction
│   │   ├── types.go             # Type system utilities
│   │   ├── reference.go         # Reference resolution
│   │   └── cleanup.go           # Unused definition removal
│   │
│   └── domain/                   # Domain Objects
│       ├── route.go             # Route domain object
│       ├── schema.go            # Schema wrapper
│       ├── typespec.go          # TypeSpecDef (moved from types.go)
│       └── options.go           # Parser options
│
├── examples/                     # Real importable Go projects
│   ├── basicapp/                # Simple CRUD app
│   │   ├── go.mod
│   │   ├── main.go
│   │   ├── models/
│   │   └── handlers/
│   │
│   └── customfields/            # Custom fields project
│       ├── go.mod               # (migrated from testdata/core_models)
│       ├── main.go
│       └── models/
│           ├── account.go
│           ├── address.go
│           └── billing.go
│
└── testdata/                     # Keep existing test data during migration
```

## Key Domain Objects

### Route (internal/domain/route.go)

Complete representation of an HTTP route with all metadata:

```go
type Route struct {
    Method      string
    Path        string
    OperationID string
    Summary     string
    Description string
    Tags        []string
    IsPublic    bool        // @Public annotation

    // Request
    Parameters  []Parameter
    RequestBody *RequestBody

    // Response
    Responses   map[int]Response

    // Source location
    FilePath     string
    FunctionName string
    LineNumber   int
}
```

This replaces the current `Operation` struct and makes route handling clearer.

## Service Architecture

### 1. LoaderService (internal/loader/service.go)

**Responsibility**: Discover and load Go packages with their AST files

**State**:
- parseVendor, parseInternal flags
- excludes map
- packagePrefix
- useGoList, useGoPackages flags

**Key Methods**:
- `LoadSearchDirs(dirs []string) (*LoadResult, error)`
- `LoadDependencies(dirs []string, depth int) (*LoadResult, error)`

### 2. RegistryService (internal/registry/service.go)

**Responsibility**: Maintain registry of all types, packages, and enums

**State**:
- packages map[string]*ParsedPackage
- uniqueDefinitions map[string]*TypeSpecDef
- enums map[string][]EnumValue

**Key Methods**:
- `RegisterPackage(pkg *ParsedPackage) error`
- `RegisterType(typeSpec *TypeSpecDef) error`
- `LookupType(pkgPath, typeName string) (*TypeSpecDef, bool)`
- `ParseTypes() error`

### 3. BaseParserService (internal/parser/base/service.go)

**Responsibility**: Parse general API info (@title, @version, @description)

**Key Methods**:
- `ParseGeneralInfo(mainFile string) error`
- `parseSecurityDefinitions(comments []string) error`

### 4. StructParserService (internal/parser/struct/service.go)

**Responsibility**: Parse Go structs into OpenAPI schemas

**State**:
- registry reference
- schemaBuilder reference
- fieldParserFactory
- propNamingStrategy
- structStack (recursion detection)
- parsedSchemas cache

**Key Methods**:
- `ParseStruct(typeSpec *TypeSpecDef) (*spec.Schema, error)`
- `ParseCustomModelStruct(typeSpec *TypeSpecDef) (*spec.Schema, error)`

**Sub-components**:
- StandardStructParser (standard.go): Regular Go structs
- CustomModelParser (custom.go): fields.StructField[T] handling
- FieldParser (field.go): Individual field parsing
- GenericsParser (generics.go): Generic type resolution

### 5. RouteParserService (internal/parser/route/service.go)

**Responsibility**: Parse function comments to extract route definitions

**State**:
- registry reference
- structParser reference
- codeExampleFilesDir
- tags map

**Key Methods**:
- `ParseRoutes(astFile *ast.File) ([]Route, error)`
- `ParseOperation(funcDecl *ast.FuncDecl) (*Route, error)`

### 6. SchemaBuilderService (internal/schema/builder.go)

**Responsibility**: Build and manage OpenAPI schemas

**State**:
- definitions map[string]spec.Schema
- parsedSchemas cache

**Key Methods**:
- `BuildSchema(typeSpec *TypeSpecDef) (string, error)`
- `BuildPublicVariant(baseSchema string) (string, error)`
- `ResolveReferences() error`
- `CleanupUnusedDefinitions(swagger *spec.Swagger)`

### 7. Main Parser (parser.go - refactored)

**Orchestrates all services** (reduced from 2,435 lines to ~300 lines):

```go
type Parser struct {
    loader       *loader.LoaderService
    registry     *registry.RegistryService
    baseParser   *base.BaseParserService
    structParser *struct.StructParserService
    routeParser  *route.RouteParserService
    schemaBuilder *schema.BuilderService
    swagger      *spec.Swagger
}
```

## Data Flow

```
1. Package Loading
   ParseAPI() → LoaderService.LoadSearchDirs()
              → LoaderService.LoadDependencies()
              → RegistryService.RegisterPackage()

2. Type Registration
   RegistryService.ParseTypes()
   → StructParserService.ParseStruct() for each type
   → SchemaBuilder.BuildSchema()

3. General API Info
   BaseParserService.ParseGeneralInfo()
   → Parse @title, @version, security definitions

4. Route Parsing
   RouteParserService.ParseRoutes()
   → RouteParserService.ParseOperation()
   → Domain.Route objects created
   → Added to swagger.Paths

5. Schema Finalization
   SchemaBuilder.ResolveReferences()
   → SchemaBuilder.CleanupUnusedDefinitions()
   → Return final Swagger spec
```

## Critical Files to Modify

1. **parser.go** (2,435 lines → ~300 lines)
   - Core orchestrator
   - Will be broken into services

2. **packages.go** (788 lines)
   - Becomes foundation for internal/registry/

3. **operation.go** (1,314 lines)
   - Refactored into internal/parser/route/ and internal/domain/route.go

4. **core_models_integration_test.go**
   - CRITICAL test for custom model parsing
   - Must pass throughout refactoring

5. **testdata/core_models/** → **examples/customfields/**
   - Migrate to real Go project with go.mod
   - Demonstrates fields.StructField[T] pattern

## Implementation Phases (TDD)

### Phase 1: Preparation & Test Data Migration (Week 1)

**Goal**: Set up new structure without breaking existing code

**Tasks**:
1. Create new package directories (internal/loader, internal/registry, etc.)
2. Create internal/domain/ with domain objects
3. Migrate testdata/core_models → examples/customfields with go.mod
4. Create examples/basicapp with go.mod
5. Update TestRealProjectIntegration to use examples/

**TDD**:
- **RED**: Write test that uses examples/customfields
- **GREEN**: Create examples/customfields as real Go project
- **VERIFY**: All existing tests still pass

**Commit**: "Setup new package structure and migrate test data"

### Phase 2: Extract LoaderService (Week 1-2)

**Goal**: Extract package loading into dedicated service

**TDD Steps**:
1. **RED**: Write internal/loader/service_test.go with expected behavior
2. **GREEN**: Create LoaderService, move code from parser.go:
   - getAllGoFileInfo → LoadSearchDirs
   - loadPackagesAndDeps → LoadDependencies
   - Functions using go list and go/packages
3. **REFACTOR**: Clean up, ensure tests pass
4. Update Parser to use LoaderService

**Files Created**:
- internal/loader/service.go (~200 lines)
- internal/loader/golist.go (~150 lines)
- internal/loader/gopackages.go (~150 lines)
- internal/loader/dependency.go (~100 lines)
- internal/loader/service_test.go

**Commit**: "Extract LoaderService for package discovery"

**Verification**: Run TestRealProjectIntegration - must pass

### Phase 3: Extract RegistryService (Week 2)

**Goal**: Create type and package registry

**TDD Steps**:
1. **RED**: Write internal/registry/service_test.go
2. **GREEN**: Create RegistryService, move from packages.go:
   - PackagesDefinitions → RegistryService
   - Package registry logic
   - Type registration
   - Enum handling
3. **REFACTOR**: Connect to LoaderService output
4. Update Parser to use RegistryService

**Files Created**:
- internal/registry/service.go (~250 lines)
- internal/registry/packages.go (~200 lines)
- internal/registry/types.go (~150 lines)
- internal/registry/enums.go (~100 lines)
- internal/registry/lookup.go (~100 lines)
- internal/registry/service_test.go

**Commit**: "Extract RegistryService for type and package management"

**Verification**: Run TestRealProjectIntegration - must pass

### Phase 4: Extract SchemaBuilderService (Week 2-3)

**Goal**: Separate schema building logic

**TDD Steps**:
1. **RED**: Write internal/schema/builder_test.go
2. **GREEN**: Create SchemaBuilderService, move from parser.go:
   - Schema building utilities
   - Reference resolution
   - Definition management
3. **GREEN**: Move cleanup.go → internal/schema/cleanup.go
4. **REFACTOR**: Clean up schema utilities

**Files Created**:
- internal/schema/builder.go (~300 lines)
- internal/schema/types.go (~150 lines)
- internal/schema/reference.go (~150 lines)
- internal/schema/cleanup.go (existing file moved)
- internal/schema/builder_test.go

**Commit**: "Extract SchemaBuilderService for schema management"

**Verification**: Run TestCoreModelsIntegration - must pass (critical for custom models)

### Phase 5: Extract BaseParserService (Week 3)

**Goal**: Separate general API info parsing

**TDD Steps**:
1. **RED**: Write internal/parser/base/service_test.go
2. **GREEN**: Create BaseParserService, move from parser.go:
   - ParseGeneralAPIInfo logic
   - Security definitions parsing
   - Contact info, license, etc.
3. **REFACTOR**: Clean up and organize
4. Update Parser orchestration

**Files Created**:
- internal/parser/base/service.go (~200 lines)
- internal/parser/base/security.go (~150 lines)
- internal/parser/base/service_test.go

**Commit**: "Extract BaseParserService for general API info"

**Verification**: Run all parser tests - must pass

### Phase 6: Extract StructParserService (Week 3-4)

**Goal**: Create dedicated struct parsing service

**TDD Steps**:
1. **RED**: Write comprehensive internal/parser/struct/service_test.go
2. **GREEN**: Create StructParserService, move from parser.go:
   - ParseDefinition logic
   - Struct parsing
   - Field parsing
3. **GREEN**: Create StandardStructParser (standard.go)
4. **GREEN**: Create CustomModelParser (custom.go) for fields.StructField[T]
5. **REFACTOR**: Break into smaller files:
   - Move field_parser.go → internal/parser/struct/field.go
   - Move generics.go → internal/parser/struct/generics.go
6. Update Parser to use StructParserService

**Files Created**:
- internal/parser/struct/service.go (~300 lines)
- internal/parser/struct/standard.go (~200 lines)
- internal/parser/struct/custom.go (~200 lines)
- internal/parser/struct/field.go (existing, moved)
- internal/parser/struct/generics.go (existing, moved)
- internal/parser/struct/service_test.go
- internal/parser/struct/custom_test.go

**Commit**: "Extract StructParserService with standard and custom parsers"

**Verification**: Run TestCoreModelsIntegration - CRITICAL, must pass

### Phase 7: Extract RouteParserService (Week 4-5)

**Goal**: Create dedicated route parsing service

**TDD Steps**:
1. **RED**: Write internal/parser/route/service_test.go
2. **GREEN**: Create RouteParserService
3. **GREEN**: Create internal/domain/route.go (Route domain object)
4. **GREEN**: Move and refactor operation.go:
   - Extract Operation → Route conversion
   - Break into service.go, operation.go, parameter.go, response.go
5. **REFACTOR**: Clean up and organize into smaller files
6. Update Parser to use RouteParserService

**Files Created**:
- internal/parser/route/service.go (~250 lines)
- internal/parser/route/operation.go (~300 lines)
- internal/parser/route/parameter.go (~200 lines)
- internal/parser/route/response.go (~200 lines)
- internal/domain/route.go (~150 lines)
- internal/parser/route/service_test.go
- internal/parser/route/operation_test.go

**Commit**: "Extract RouteParserService and Route domain object"

**Verification**: Run all route/operation tests - must pass

### Phase 8: Final Integration & Cleanup (Week 5)

**Goal**: Connect all services, finalize orchestrator

**TDD Steps**:
1. Refactor parser.go to orchestrator (~300 lines)
2. Remove old code that's been moved
3. Update all imports and references
4. Run full test suite
5. Verify no file exceeds 500 lines

**Tasks**:
- Simplify parser.go to service orchestration
- Update documentation
- Clean up unused code
- Verify all tests pass

**Commit**: "Complete refactoring - orchestrator and final cleanup"

**Verification**: ALL tests must pass:
- TestRealProjectIntegration ✓
- TestCoreModelsIntegration ✓
- All service unit tests ✓
- All integration tests ✓

### Phase 9: Documentation (Week 6) ✅ COMPLETE

**Goal**: Document new architecture

**Tasks**:
1. ✅ Write package-level documentation for each internal package
2. ✅ Update README.md with new architecture
3. ✅ Add architecture documentation (ARCHITECTURE.md)
4. ✅ Document service interactions
5. ✅ Create service README files for all packages

**Commit**: "Add comprehensive documentation for refactored architecture"

**Completed**:
- Created ARCHITECTURE.md with comprehensive overview
- Updated main README.md with architecture section
- Created README.md for:
  - internal/loader/ (already existed)
  - internal/registry/ (already existed)
  - internal/schema/
  - internal/parser/base/
  - internal/parser/struct/
  - internal/parser/route/
- All documentation includes:
  - Package purpose and overview
  - File descriptions
  - Usage examples
  - Key methods and types
  - Integration details
  - Testing information
  - Design principles
  - Common patterns

## Verification Strategy

### After Each Phase

1. **Unit Tests**: All new service tests pass
2. **Integration Tests**:
   - TestRealProjectIntegration passes
   - TestCoreModelsIntegration passes (critical for Phases 4, 6)
3. **Code Quality**:
   - No file exceeds 500 lines
   - No circular dependencies
   - Clear service boundaries

### Final Verification

1. ✓ All existing tests pass
2. ✓ New service tests have 90%+ coverage
3. ✓ examples/ projects build and run
4. ✓ No breaking changes to public API
5. ✓ All features preserved:
   - Custom model parsing (fields.StructField[T])
   - Public/private field filtering
   - Generic type support
   - Enum handling
   - @Public annotation behavior
   - Schema composition (AllOf)

## Risk Mitigation

### Risk: Breaking TestCoreModelsIntegration

This test validates custom model parsing (fields.StructField[T]), which is critical functionality.

**Mitigation**:
- Run after EVERY change during Phase 6 (StructParserService extraction)
- Keep CustomModelParser separate (custom.go)
- Add additional custom model tests before refactoring
- Test public/private filtering extensively

### Risk: Complex State Management

Services need to communicate without tight coupling.

**Mitigation**:
- Use dependency injection (services passed to constructors)
- Services return data rather than modifying global state
- Clear ownership: SchemaBuilderService owns swagger.Definitions
- Registry owns types, Parser orchestrates

### Risk: Test Data Migration Issues

Moving testdata/ to examples/ could break tests.

**Mitigation**:
- Do migration in Phase 1 before any code changes
- Keep both testdata/ and examples/ during migration
- Validate examples/ are real, buildable Go projects (run go build)
- Only remove testdata/ after all tests use examples/

## Success Criteria

### Code Organization
- ✓ No file exceeds 500 lines
- ✓ Clear package boundaries
- ✓ Services own their state
- ✓ No circular dependencies

### Functionality
- ✓ All existing features work
- ✓ Custom model parsing preserved
- ✓ Public/private filtering works
- ✓ Generics handled correctly
- ✓ All swagger annotations supported

### Testing
- ✓ 90%+ test coverage for new services
- ✓ TestRealProjectIntegration passes
- ✓ TestCoreModelsIntegration passes
- ✓ Fast, isolated unit tests
- ✓ examples/ projects are real and buildable

### Developer Experience
- ✓ Easy to find code for specific functionality
- ✓ Clear service responsibilities
- ✓ Self-documenting domain objects
- ✓ Easy to add new parsers or schema types
- ✓ Clear error messages with context

## Project Status: COMPLETE ✅

All phases of the refactoring have been completed successfully:

### Phase Summary

| Phase | Status | Description |
|-------|--------|-------------|
| Phase 1 | ✅ COMPLETE | Preparation & test data migration |
| Phase 2 | ✅ COMPLETE | Extract LoaderService |
| Phase 3 | ✅ COMPLETE | Extract RegistryService |
| Phase 4 | ✅ COMPLETE | Extract SchemaBuilderService |
| Phase 5 | ✅ COMPLETE | Extract BaseParserService |
| Phase 6 | ✅ COMPLETE | Extract StructParserService |
| Phase 7 | ✅ COMPLETE | Extract RouteParserService |
| Phase 8 | ✅ COMPLETE | Final integration & cleanup |
| Phase 9 | ✅ COMPLETE | Comprehensive documentation |

### Achievements

**Code Organization**:
- Reduced parser.go from 2,435 lines to orchestrator pattern (~300 lines)
- Reduced operation.go from 1,314 lines to ~1,270 lines across 5 focused files
- Reduced packages.go from 788 lines to multiple services (~800 lines total, better organized)
- All files now under 500 lines
- Clear separation of concerns

**Architecture**:
- 6 focused services (Loader, Registry, Schema, Base Parser, Struct Parser, Route Parser)
- Each service owns its state and has clear responsibilities
- Dependency injection for testability
- Well-documented with comprehensive README files

**Testing**:
- All services have comprehensive unit tests
- Integration tests passing (TestRealProjectIntegration, etc.)
- Test coverage >90% for new services
- Examples migrated to real Go projects

**Documentation**:
- ARCHITECTURE.md provides comprehensive overview
- Each service has detailed README.md
- Main README.md updated with architecture section
- Clear data flow and integration documentation

### Next Steps for Future Work

1. **Performance Optimization**: Profile and optimize hot paths
2. **Parallel Parsing**: Parse independent files in parallel
3. **Caching**: Implement parser result caching for incremental builds
4. **Plugin System**: Allow custom parsers for special annotations
5. **Enhanced Error Messages**: Add more context and suggestions

### Key Files for Reference

- [ARCHITECTURE.md](../../ARCHITECTURE.md) - Overall architecture documentation
- [REFACTORING_STATUS.md](../../REFACTORING_STATUS.md) - Refactoring progress tracker
- Service READMEs:
  - [Loader Service](../../internal/loader/README.md)
  - [Registry Service](../../internal/registry/README.md)
  - [Schema Builder](../../internal/schema/README.md)
  - [Base Parser](../../internal/parser/base/README.md)
  - [Struct Parser](../../internal/parser/struct/README.md)
  - [Route Parser](../../internal/parser/route/README.md)
