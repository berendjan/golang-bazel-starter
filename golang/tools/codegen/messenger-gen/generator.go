package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"strings"
	"text/template"
)

// Generator generates Go code from a MessengerSpec
type Generator struct {
	spec *MessengerSpec
}

// NewGenerator creates a new code generator
func NewGenerator(spec *MessengerSpec) *Generator {
	return &Generator{spec: spec}
}

// Spec returns the spec for template access
func (g *Generator) Spec() *MessengerSpec {
	return g.spec
}

// RoutesForHandler returns all routes where the given handler is the source
func (g *Generator) RoutesForHandler(handlerName string) []Route {
	var routes []Route
	for _, route := range g.spec.Routes {
		if route.Source == handlerName {
			routes = append(routes, route)
		}
	}
	return routes
}

// HasSendableMessages returns true if the handler has any messages it can send
func (g *Generator) HasSendableMessages(handlerName string) bool {
	routes := g.RoutesForHandler(handlerName)
	return len(routes) > 0
}

// ReceivesMessages returns true if the handler receives any messages
func (g *Generator) ReceivesMessages(handlerName string) bool {
	for _, route := range g.spec.Routes {
		for _, msg := range route.Messages {
			for _, receiver := range msg.Receivers {
				if receiver == handlerName {
					return true
				}
			}
		}
	}
	return false
}

// HandlersReceivingMessages returns only the handlers that receive messages
func (g *Generator) HandlersReceivingMessages() []Handler {
	var handlers []Handler
	for _, handler := range g.spec.Handlers {
		if g.ReceivesMessages(handler.Name) {
			handlers = append(handlers, handler)
		}
	}
	return handlers
}

// GetHandlerPackages returns a map of package aliases used by handlers
func (g *Generator) GetHandlerPackages() map[string]bool {
	packages := make(map[string]bool)
	for _, handler := range g.spec.Handlers {
		// Extract package from type like "configapi.ConfigurationApi" -> "configapi"
		parts := strings.Split(handler.Type, ".")
		if len(parts) > 1 {
			pkg := parts[0]
			// Remove generic type parameters if present
			pkg = strings.Split(pkg, "[")[0]
			packages[pkg] = true
		}
	}
	return packages
}

// Generate produces the Go source code
func (g *Generator) Generate() ([]byte, error) {
	// Create template with custom functions
	tmpl, err := template.New("messenger").Funcs(template.FuncMap{
		"title": strings.Title,
		"sub": func(a, b int) int {
			return a - b
		},
		"baseName": func(s string) string {
			// Extract base name from type like "*configpb.AccountCreationRequestProto" -> "AccountCreationRequest"
			s = strings.TrimPrefix(s, "*")
			parts := strings.Split(s, ".")
			if len(parts) > 0 {
				name := parts[len(parts)-1]
				// Remove "Proto" suffix if present
				return strings.TrimSuffix(name, "Proto")
			}
			return s
		},
	}).Parse(fileTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template with Generator as context so it can call methods
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, g); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// If formatting fails, return the unformatted code with an error
		// This helps debugging template issues
		return buf.Bytes(), fmt.Errorf("failed to format generated code: %w", err)
	}

	return formatted, nil
}

// WriteToFile generates code and writes it to the specified file
func (g *Generator) WriteToFile(filepath string) error {
	code, err := g.Generate()
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath, code, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
