package transport

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/Robotech-Org/driver-management/internal/wiring"
	"github.com/go-chi/chi/v5"
	"gopkg.in/yaml.v3"
)

// Routes that serve the docs themselves and are intentionally not in the spec.
var undocumentedRoutes = map[string]bool{
	"GET /":                     true,
	"GET /docs":                 true,
	"GET /docs/*":               true,
	"GET /swagger":              true,
	"GET /swagger.yaml":         true,
	"GET /swagger/swagger.yaml": true,
}

func loadSpec(t *testing.T) map[string]any {
	t.Helper()
	// findDocsDir works whatever the current directory is (routes_test.go
	// changes it with os.Chdir, which affects every test in this package).
	b, err := os.ReadFile(filepath.Join(findDocsDir(), "swagger.yaml"))
	if err != nil {
		t.Fatalf("read swagger.yaml: %v", err)
	}
	var spec map[string]any
	// yaml.v3 rejects duplicate keys, which also catches copy/paste mistakes.
	if err := yaml.Unmarshal(b, &spec); err != nil {
		t.Fatalf("swagger.yaml is not valid YAML: %v", err)
	}
	if spec["swagger"] != "2.0" {
		t.Fatalf(`expected swagger: "2.0", got %v`, spec["swagger"])
	}
	return spec
}

// TestSwaggerMatchesRouter fails when a route is added/removed in routes.go
// without updating docs/swagger.yaml (or the other way round).
func TestSwaggerMatchesRouter(t *testing.T) {
	spec := loadSpec(t)

	documented := map[string]bool{}
	for path, item := range spec["paths"].(map[string]any) {
		for method := range item.(map[string]any) {
			documented[strings.ToUpper(method)+" "+path] = true
		}
	}

	router := NewRoutes(&wiring.DependencyContainer{})
	registered := map[string]bool{}
	err := chi.Walk(router, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		route = strings.ReplaceAll(route, "/*/", "/") // chi sub-router mounts
		if len(route) > 1 {
			route = strings.TrimSuffix(route, "/")
		}
		key := method + " " + route
		if strings.HasPrefix(route, "/swagger/") || undocumentedRoutes[key] {
			return nil
		}
		registered[key] = true
		return nil
	})
	if err != nil {
		t.Fatalf("walk routes: %v", err)
	}

	var missing, stale []string
	for k := range registered {
		if !documented[k] {
			missing = append(missing, k)
		}
	}
	for k := range documented {
		if !registered[k] {
			stale = append(stale, k)
		}
	}
	sort.Strings(missing)
	sort.Strings(stale)
	for _, k := range missing {
		t.Errorf("route not documented in swagger.yaml: %s", k)
	}
	for _, k := range stale {
		t.Errorf("swagger.yaml documents a route that does not exist: %s", k)
	}
}

// TestSwaggerIsConsistent checks the spec's internal structure: every $ref
// resolves, every {param} in a path is declared, and every response has a
// description (required by Swagger 2.0).
func TestSwaggerIsConsistent(t *testing.T) {
	spec := loadSpec(t)
	defs, _ := spec["definitions"].(map[string]any)
	shared, _ := spec["responses"].(map[string]any)

	var walk func(where string, v any)
	walk = func(where string, v any) {
		switch n := v.(type) {
		case map[string]any:
			if ref, ok := n["$ref"].(string); ok {
				switch {
				case strings.HasPrefix(ref, "#/definitions/"):
					if _, ok := defs[strings.TrimPrefix(ref, "#/definitions/")]; !ok {
						t.Errorf("%s: unresolved $ref %s", where, ref)
					}
				case strings.HasPrefix(ref, "#/responses/"):
					if _, ok := shared[strings.TrimPrefix(ref, "#/responses/")]; !ok {
						t.Errorf("%s: unresolved $ref %s", where, ref)
					}
				default:
					t.Errorf("%s: unsupported $ref %s", where, ref)
				}
			}
			for k, child := range n {
				walk(where+"."+k, child)
			}
		case []any:
			for _, child := range n {
				walk(where, child)
			}
		}
	}
	walk("spec", spec)

	pathParam := regexp.MustCompile(`\{([^}]+)\}`)
	for path, item := range spec["paths"].(map[string]any) {
		for method, rawOp := range item.(map[string]any) {
			op := rawOp.(map[string]any)
			where := strings.ToUpper(method) + " " + path

			declared := map[string]bool{}
			params, _ := op["parameters"].([]any)
			for _, p := range params {
				pm := p.(map[string]any)
				if pm["in"] == "path" {
					declared[pm["name"].(string)] = true
				}
			}
			for _, m := range pathParam.FindAllStringSubmatch(path, -1) {
				if !declared[m[1]] {
					t.Errorf("%s: path parameter {%s} is not declared", where, m[1])
				}
			}

			resps, ok := op["responses"].(map[string]any)
			if !ok || len(resps) == 0 {
				t.Errorf("%s: no responses", where)
				continue
			}
			for code, r := range resps {
				rm := r.(map[string]any)
				if _, isRef := rm["$ref"]; isRef {
					continue
				}
				if d, _ := rm["description"].(string); d == "" {
					t.Errorf("%s: response %s has no description", where, code)
				}
			}
		}
	}
}
