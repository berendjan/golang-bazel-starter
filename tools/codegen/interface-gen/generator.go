package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"strings"
	"text/template"
)

// Generator generates Go interface code from an InterfaceSpec
type Generator struct {
	spec *InterfaceSpec
}

// NewGenerator creates a new interface code generator
func NewGenerator(spec *InterfaceSpec) *Generator {
	return &Generator{spec: spec}
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

// RoutesReceivedBy returns all routes where the given handler is a receiver
func (g *Generator) RoutesReceivedBy(handlerName string) []Route {
	var routes []Route
	for _, route := range g.spec.Routes {
		// Check if this handler is a receiver for any message in this route
		hasMessages := false
		filteredMessages := []MessageRoute{}

		for _, msg := range route.Messages {
			for _, receiver := range msg.Receivers {
				if receiver == handlerName {
					filteredMessages = append(filteredMessages, msg)
					hasMessages = true
					break
				}
			}
		}

		if hasMessages {
			routes = append(routes, Route{
				Source:   route.Source,
				Messages: filteredMessages,
			})
		}
	}
	return routes
}

// Spec returns the spec for template access
func (g *Generator) Spec() *InterfaceSpec {
	return g.spec
}

// HasSendableMessages returns true if the handler has any messages it can send
func (g *Generator) HasSendableMessages(handlerName string) bool {
	routes := g.RoutesForHandler(handlerName)
	return len(routes) > 0
}

// Generate produces the Go interface source code
func (g *Generator) Generate() ([]byte, error) {
	// Create template with custom functions
	tmpl, err := template.New("interfaces").Funcs(template.FuncMap{
		"title": strings.Title,
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
