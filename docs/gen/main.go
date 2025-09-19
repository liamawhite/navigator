// Copyright 2025 Navigator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := generateConfigDocs(); err != nil {
		log.Fatalf("Failed to generate config documentation: %v", err)
	}
	fmt.Println("Configuration documentation generated: docs/reference/config/navctl.md")
}

func generateConfigDocs() error {
	// Parse the config package
	fset := token.NewFileSet()
	pkgPath := "./navctl/pkg/config"

	pkgs, err := parser.ParseDir(fset, pkgPath, func(info os.FileInfo) bool {
		return strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse package: %w", err)
	}

	configPkg := pkgs["config"]
	if configPkg == nil {
		return fmt.Errorf("config package not found")
	}

	// Create documentation
	docPkg := doc.New(configPkg, "github.com/liamawhite/navigator/navctl/pkg/config", 0)

	// Define the order we want types to appear
	typeOrder := []string{
		"Config",
		"ManagerConfig",
		"EdgeConfig",
		"UIConfig",
		"MetricsConfig",
		"MetricsAuth",
		"ExecConfig",
		"EnvVar",
	}

	// Generate markdown
	var content strings.Builder
	content.WriteString("# navctl Configuration Reference\n\n")
	content.WriteString("This document describes the configuration file format for navctl.\n\n")

	// Add table of contents
	content.WriteString("## Table of Contents\n\n")
	for _, typeName := range typeOrder {
		fmt.Fprintf(&content, "- [%s](#%s)\n", typeName, strings.ToLower(typeName))
	}
	content.WriteString("\n")

	// Generate documentation for each type in order
	for _, typeName := range typeOrder {
		for _, t := range docPkg.Types {
			if t.Name == typeName {
				generateTypeDoc(&content, t, fset)
				break
			}
		}
	}

	// Write to file
	outputPath := "docs/reference/config/navctl.md"
	if err := os.MkdirAll(filepath.Dir(outputPath), 0750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(content.String()), 0600); err != nil {
		return fmt.Errorf("failed to write documentation: %w", err)
	}

	return nil
}

func generateTypeDoc(content *strings.Builder, t *doc.Type, fset *token.FileSet) {
	fmt.Fprintf(content, "## %s\n\n", t.Name)

	// Add type description
	if t.Doc != "" {
		fmt.Fprintf(content, "%s\n\n", formatDocComment(t.Doc))
	}

	// Add field documentation for struct types
	if structType, ok := t.Decl.Specs[0].(*ast.TypeSpec).Type.(*ast.StructType); ok {
		content.WriteString("### Fields\n\n")

		for _, field := range structType.Fields.List {
			if len(field.Names) > 0 {
				yamlName := extractYamlFieldName(field)
				fieldType := formatType(field.Type)

				// Add field documentation
				if field.Doc != nil && field.Doc.Text() != "" {
					fmt.Fprintf(content, "#### `%s`\n\n", yamlName)

					docText := strings.TrimSpace(field.Doc.Text())
					// Clean up the field documentation and add cross-references
					processedDoc := addCrossReferences(docText, fieldType)
					lines := strings.Split(processedDoc, "\n")
					var cleanLines []string
					for _, line := range lines {
						line = strings.TrimSpace(line)
						if line != "" {
							cleanLines = append(cleanLines, line)
						}
					}
					content.WriteString(strings.Join(cleanLines, " "))

					// Add type reference for complex types
					if isComplexType(fieldType) {
						fmt.Fprintf(content, "\n\nSee [%s](#%s) for configuration details.", fieldType, strings.ToLower(fieldType))
					}
					content.WriteString("\n\n")
				}
			}
		}
	}
}

func formatDocComment(doc string) string {
	// Clean up the doc comment formatting
	lines := strings.Split(strings.TrimSpace(doc), "\n")
	var formatted []string

	inCodeBlock := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			formatted = append(formatted, "")
			continue
		}

		// Detect code blocks
		if strings.HasPrefix(line, "    ") && !inCodeBlock {
			formatted = append(formatted, "```yaml")
			inCodeBlock = true
		} else if !strings.HasPrefix(line, "    ") && inCodeBlock {
			formatted = append(formatted, "```")
			inCodeBlock = false
		}

		if inCodeBlock {
			formatted = append(formatted, strings.TrimPrefix(line, "    "))
		} else {
			formatted = append(formatted, line)
		}
	}

	if inCodeBlock {
		formatted = append(formatted, "```")
	}

	return strings.Join(formatted, "\n")
}

func extractYamlFieldName(field *ast.Field) string {
	if field.Tag == nil {
		return strings.ToLower(field.Names[0].Name)
	}

	tag := field.Tag.Value
	// Remove surrounding backticks
	tag = strings.Trim(tag, "`")

	// Look for yaml tag
	parts := strings.Fields(tag)
	for _, part := range parts {
		if strings.HasPrefix(part, "yaml:") {
			yamlTag := strings.TrimPrefix(part, "yaml:")
			yamlTag = strings.Trim(yamlTag, "\"")
			// Extract just the field name (before any options like omitempty)
			yamlName := strings.Split(yamlTag, ",")[0]
			if yamlName != "" && yamlName != "-" {
				return yamlName
			}
		}
	}

	// Fallback to lowercase field name
	return strings.ToLower(field.Names[0].Name)
}

func formatType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return formatType(t.X) // Remove pointer indicator for doc purposes
	case *ast.ArrayType:
		return formatType(t.Elt) // Show element type for arrays
	case *ast.SelectorExpr:
		return formatType(t.X) + "." + t.Sel.Name
	default:
		return "interface{}"
	}
}

func isComplexType(typeName string) bool {
	complexTypes := []string{
		"ManagerConfig", "EdgeConfig", "UIConfig",
		"MetricsConfig", "MetricsAuth", "ExecConfig", "EnvVar",
	}

	for _, complexType := range complexTypes {
		if typeName == complexType {
			return true
		}
	}
	return false
}

func addCrossReferences(docText, fieldType string) string {
	// Add references to related configuration sections
	replacements := map[string]string{
		"ManagerConfig": "[ManagerConfig](#managerconfig)",
		"EdgeConfig":    "[EdgeConfig](#edgeconfig)",
		"UIConfig":      "[UIConfig](#uiconfig)",
		"MetricsConfig": "[MetricsConfig](#metricsconfig)",
		"MetricsAuth":   "[MetricsAuth](#metricsauth)",
		"ExecConfig":    "[ExecConfig](#execconfig)",
		"EnvVar":        "[EnvVar](#envvar)",
		"prometheus":    "Prometheus",
		"kubectl":       "`kubectl`",
		"kubeconfig":    "kubeconfig",
	}

	result := docText
	for term, replacement := range replacements {
		// Only replace if it's not already a link
		if !strings.Contains(result, "["+term+"]") {
			result = strings.ReplaceAll(result, term, replacement)
		}
	}

	return result
}
