package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// InterfaceSpec defines the YAML specification structure for interface generation
type InterfaceSpec struct {
	Package string         `yaml:"package"`
	Imports []string       `yaml:"imports,omitempty"`
	Handlers []Handler     `yaml:"handlers"`
	Routes  []Route        `yaml:"routes"`
}

// Handler defines a handler with its name and type
type Handler struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

// Route defines routing for a source with multiple messages
type Route struct {
	Source   string          `yaml:"source"`
	Messages []MessageRoute  `yaml:"messages"`
}

// MessageRoute defines a specific message routing configuration
type MessageRoute struct {
	Message   string   `yaml:"message"`
	Response  string   `yaml:"response,omitempty"`
	Receivers []string `yaml:"receivers"`
}

// LoadSpec loads and validates an interface specification from YAML
func LoadSpec(filepath string) (*InterfaceSpec, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	var spec InterfaceSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := spec.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &spec, nil
}

// Validate checks if the spec is valid
func (s *InterfaceSpec) Validate() error {
	if s.Package == "" {
		return fmt.Errorf("package name is required")
	}
	if len(s.Handlers) == 0 {
		return fmt.Errorf("at least one handler is required")
	}
	if len(s.Routes) == 0 {
		return fmt.Errorf("at least one route is required")
	}

	// Validate handlers
	for i, h := range s.Handlers {
		if h.Name == "" {
			return fmt.Errorf("handler %d: name is required", i)
		}
		if h.Type == "" {
			return fmt.Errorf("handler %d: type is required", i)
		}
	}

	// Build a map of valid handler names for validation
	handlerNames := make(map[string]bool)
	for _, h := range s.Handlers {
		handlerNames[h.Name] = true
	}

	// Validate routes
	for i, r := range s.Routes {
		if r.Source == "" {
			return fmt.Errorf("route %d: source is required", i)
		}

		// Validate source handler exists
		if !handlerNames[r.Source] {
			return fmt.Errorf("route %d: unknown handler '%s' in source (available handlers: %v)", i, r.Source, getHandlerNamesList(s.Handlers))
		}

		if len(r.Messages) == 0 {
			return fmt.Errorf("route %d: at least one message is required for source %s", i, r.Source)
		}

		// Validate each message route
		for j, m := range r.Messages {
			if m.Message == "" {
				return fmt.Errorf("route %d, message %d: message type is required", i, j)
			}
			if len(m.Receivers) == 0 {
				return fmt.Errorf("route %d, message %d: at least one receiver is required", i, j)
			}

			// Validate receiver handlers exist
			for k, receiver := range m.Receivers {
				if !handlerNames[receiver] {
					return fmt.Errorf("route %d, message %d, receiver %d: unknown handler '%s' (available handlers: %v)", i, j, k, receiver, getHandlerNamesList(s.Handlers))
				}
			}
		}
	}

	return nil
}

// getHandlerNamesList returns a list of handler names for error messages
func getHandlerNamesList(handlers []Handler) []string {
	names := make([]string, len(handlers))
	for i, h := range handlers {
		names[i] = h.Name
	}
	return names
}
