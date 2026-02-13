package route

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/swaggo/swag/internal/parser/route/domain"
)

var paramPattern = regexp.MustCompile(`(\S+)\s+(\w+)\s+([\S.]+)\s+(\w+)\s+"([^"]+)"`)

// parseParam parses the @param annotation
// Format: @Param name paramType dataType required "description"
// Example: @Param id path int true "User ID"
func (s *Service) parseParam(op *operation, line string) error {
	matches := paramPattern.FindStringSubmatch(line)
	if len(matches) != 6 {
		return fmt.Errorf("invalid param format: %s", line)
	}

	name := matches[1]
	paramType := matches[2] // path, query, header, body, formData
	dataType := matches[3]
	requiredStr := strings.ToLower(matches[4])
	description := matches[5]

	required := requiredStr == "true" || requiredStr == "required"

	// Determine if it's an array
	isArray := strings.HasPrefix(dataType, "[]")
	if isArray {
		dataType = strings.TrimPrefix(dataType, "[]")
	}

	// Convert Go types to OpenAPI types
	schemaType, format := convertType(dataType)

	param := domain.Parameter{
		Name:        name,
		In:          paramType,
		Required:    required,
		Description: description,
		Format:      format,
	}

	if isArray {
		param.Type = "array"
		param.Items = &domain.Items{
			Type:   schemaType,
			Format: format,
		}
	} else {
		param.Type = schemaType
	}

	op.parameters = append(op.parameters, param)
	return nil
}

// convertType converts Go types to OpenAPI types
func convertType(goType string) (schemaType string, format string) {
	switch goType {
	case "int", "int32", "int64", "uint", "uint32", "uint64":
		return "integer", goType
	case "float32", "float64":
		return "number", goType
	case "bool":
		return "boolean", ""
	case "string":
		return "string", ""
	case "byte":
		return "string", "byte"
	case "rune":
		return "integer", "int32"
	case "object":
		return "object", ""
	case "array":
		return "array", ""
	default:
		// For custom types, treat as object
		return "object", ""
	}
}
