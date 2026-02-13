// Package route provides parsing functionality for route annotations in Go source files.
// It extracts HTTP route information from function comments including @router, @param,
// @success, @failure, @header, and other route-related annotations.
package route

import (
	"go/ast"

	"github.com/swaggo/swag/internal/parser/route/domain"
)

// Service handles parsing of route annotations from Go source files
type Service struct {
	registry            interface{} // TODO: type this properly
	structParser        interface{} // TODO: type this properly
	codeExampleFilesDir string
	collectionFormat    string
}

// NewService creates a new route parser service
func NewService(registry interface{}, structParser interface{}) *Service {
	return &Service{
		registry:         registry,
		structParser:     structParser,
		collectionFormat: "csv", // default
	}
}

// ParseRoutes extracts all routes from an AST file
func (s *Service) ParseRoutes(astFile *ast.File) ([]*domain.Route, error) {
	var routes []*domain.Route

	// Iterate through all declarations in the file
	for _, decl := range astFile.Decls {
		// Only process function declarations
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Doc == nil {
			continue
		}

		// Parse the function's documentation comments
		operation := s.parseOperation(funcDecl)
		if operation == nil {
			continue
		}

		// Convert operation to routes (one operation can have multiple routes)
		operationRoutes := s.operationToRoutes(operation)
		routes = append(routes, operationRoutes...)
	}

	return routes, nil
}

// parseOperation parses a function declaration into an operation
func (s *Service) parseOperation(funcDecl *ast.FuncDecl) *operation {
	op := &operation{
		functionName: funcDecl.Name.Name,
		routerPaths:  []routerPath{},
		parameters:   []domain.Parameter{},
		responses:    make(map[int]domain.Response),
		security:     []map[string][]string{},
		tags:         []string{},
		consumes:     []string{},
		produces:     []string{},
	}

	// Parse each comment line
	for _, comment := range funcDecl.Doc.List {
		if err := s.parseComment(op, comment.Text); err != nil {
			// Skip comments that fail to parse
			continue
		}
	}

	// Only return if we have at least one router path
	if len(op.routerPaths) == 0 {
		return nil
	}

	return op
}

// operationToRoutes converts an operation into one or more routes
func (s *Service) operationToRoutes(op *operation) []*domain.Route {
	var routes []*domain.Route

	// Create one route for each router path
	for _, routerPath := range op.routerPaths {
		route := &domain.Route{
			Method:       routerPath.method,
			Path:         routerPath.path,
			Summary:      op.summary,
			Description:  op.description,
			Tags:         op.tags,
			Parameters:   op.parameters,
			Responses:    op.responses,
			Security:     op.security,
			Consumes:     op.consumes,
			Produces:     op.produces,
			IsPublic:     op.isPublic,
			Deprecated:   routerPath.deprecated,
			OperationID:  op.operationID,
			FunctionName: op.functionName,
		}

		routes = append(routes, route)
	}

	return routes
}
