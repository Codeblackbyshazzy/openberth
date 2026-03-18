package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ScaffoldResult struct {
	Dir       string
	Framework string
	Cleanup   func()
}

// ScaffoldSingleFile wraps a bare component file in a deployable project.
// Supported: .jsx, .tsx, .vue, .svelte, .html, .md, .ipynb
func ScaffoldSingleFile(filePath string) (*ScaffoldResult, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	name := strings.TrimSuffix(filepath.Base(filePath), ext)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	// Static HTML — no scaffolding needed, just wrap in a directory
	if ext == ".html" {
		dir, err := os.MkdirTemp("", "openberth-html-*")
		if err != nil {
			return nil, err
		}
		os.WriteFile(filepath.Join(dir, "index.html"), content, 0644)
		return &ScaffoldResult{
			Dir:       dir,
			Framework: "Static HTML",
			Cleanup:   func() { os.RemoveAll(dir) },
		}, nil
	}

	switch ext {
	case ".jsx", ".tsx":
		return scaffoldReact(name, ext, content)
	case ".vue":
		return scaffoldVue(name, content)
	case ".svelte":
		return scaffoldSvelte(name, content)
	case ".md":
		return scaffoldMarkdown(name, content)
	case ".ipynb":
		return scaffoldNotebook(name, content)
	default:
		return nil, fmt.Errorf("unsupported file type: %s (supported: .jsx, .tsx, .vue, .svelte, .html, .md, .ipynb)", ext)
	}
}

// IsSingleFile checks if the path is a single deployable file.
func IsSingleFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jsx", ".tsx", ".vue", ".svelte", ".html", ".md", ".ipynb":
		return true
	}
	return false
}

// -- Shared helpers --

func writeTailwindConfig(dir string, contentPaths []string) {
	paths := ""
	for _, p := range contentPaths {
		paths += fmt.Sprintf(`"%s", `, p)
	}

	twConfig := fmt.Sprintf(`/** @type {import('tailwindcss').Config} */
export default {
  content: [%s"./index.html"],
  theme: { extend: {} },
  plugins: [],
}
`, paths)
	os.WriteFile(filepath.Join(dir, "tailwind.config.js"), []byte(twConfig), 0644)

	postcssConfig := `export default {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
}
`
	os.WriteFile(filepath.Join(dir, "postcss.config.js"), []byte(postcssConfig), 0644)

	css := `@tailwind base;
@tailwind components;
@tailwind utilities;
`
	os.WriteFile(filepath.Join(dir, "index.css"), []byte(css), 0644)
}

func cssImport(hasTailwind bool) string {
	if hasTailwind {
		return `import './index.css'`
	}
	return ""
}

func isTailwindClasses(src string) bool {
	twIndicators := []string{"flex", "grid", "p-", "m-", "bg-", "text-", "rounded", "shadow", "w-", "h-"}
	count := 0
	for _, ind := range twIndicators {
		if strings.Contains(src, ind) {
			count++
		}
	}
	return count >= 3
}

func mapToJSON(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, fmt.Sprintf(`"%s": "%s"`, k, v))
	}
	return "\n    " + strings.Join(parts, ",\n    ") + "\n  "
}

// scanImports extracts npm package names from import/require statements.
// Skips relative imports (./foo), built-in node modules, and known framework deps.
func scanImports(src string) map[string]string {
	deps := map[string]string{}

	// Built-in/framework packages to skip (already handled elsewhere)
	skip := map[string]bool{
		"react": true, "react-dom": true, "react-dom/client": true,
		"vue": true, "svelte": true,
		"vite": true, "@vitejs/plugin-react": true, "@vitejs/plugin-vue": true,
		"@sveltejs/vite-plugin-svelte": true,
	}

	// Node built-ins
	builtins := map[string]bool{
		"fs": true, "path": true, "os": true, "url": true, "http": true,
		"https": true, "crypto": true, "stream": true, "util": true,
		"events": true, "buffer": true, "child_process": true, "net": true,
		"tls": true, "dns": true, "assert": true, "zlib": true,
	}

	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)

		var pkg string

		// import ... from "package"
		// import ... from 'package'
		if strings.Contains(line, "from ") {
			pkg = extractQuotedAfter(line, "from ")
		}

		// import "package" (side-effect import)
		if pkg == "" && strings.HasPrefix(line, "import ") && !strings.Contains(line, "from") {
			pkg = extractFirstQuoted(line)
		}

		// require("package")
		if pkg == "" && strings.Contains(line, "require(") {
			pkg = extractQuotedAfter(line, "require(")
		}

		if pkg == "" {
			continue
		}

		// Skip relative imports
		if strings.HasPrefix(pkg, ".") || strings.HasPrefix(pkg, "/") {
			continue
		}

		// Skip CSS/asset imports
		if strings.HasSuffix(pkg, ".css") || strings.HasSuffix(pkg, ".svg") ||
			strings.HasSuffix(pkg, ".png") || strings.HasSuffix(pkg, ".json") {
			continue
		}

		// Get the npm package name (handle scoped packages and subpaths)
		npmPkg := toNpmPackageName(pkg)

		if skip[npmPkg] || skip[pkg] || builtins[npmPkg] {
			continue
		}

		deps[npmPkg] = "*"
	}

	return deps
}

// toNpmPackageName extracts the npm package name from an import path.
// "lodash/merge" → "lodash"
// "@headlessui/react" → "@headlessui/react"
// "@shadcn/ui/button" → "@shadcn/ui"
func toNpmPackageName(importPath string) string {
	if strings.HasPrefix(importPath, "@") {
		// Scoped package: @scope/name or @scope/name/subpath
		parts := strings.SplitN(importPath, "/", 3)
		if len(parts) >= 2 {
			return parts[0] + "/" + parts[1]
		}
		return importPath
	}
	// Regular package: name or name/subpath
	parts := strings.SplitN(importPath, "/", 2)
	return parts[0]
}

func extractQuotedAfter(line, marker string) string {
	idx := strings.Index(line, marker)
	if idx < 0 {
		return ""
	}
	rest := line[idx+len(marker):]
	return extractFirstQuoted(rest)
}

func extractFirstQuoted(s string) string {
	for _, q := range []byte{'"', '\''} {
		start := strings.IndexByte(s, q)
		if start < 0 {
			continue
		}
		end := strings.IndexByte(s[start+1:], q)
		if end < 0 {
			continue
		}
		return s[start+1 : start+1+end]
	}
	return ""
}
