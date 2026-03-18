package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func scaffoldReact(name, ext string, content []byte) (*ScaffoldResult, error) {
	dir, err := os.MkdirTemp("", "openberth-react-*")
	if err != nil {
		return nil, err
	}

	isTS := ext == ".tsx"
	lang := "JavaScript"
	if isTS {
		lang = "TypeScript"
	}

	// Detect if it's a full component or just JSX markup
	src := string(content)
	needsWrapper := !strings.Contains(src, "export default") &&
		!strings.Contains(src, "export function") &&
		!strings.Contains(src, "export const")

	componentFile := "App" + ext
	if needsWrapper {
		// Wrap raw JSX in a component
		wrappedContent := fmt.Sprintf(`export default function App() {
  return (
    <>
      %s
    </>
  );
}
`, strings.TrimSpace(src))
		os.WriteFile(filepath.Join(dir, componentFile), []byte(wrappedContent), 0644)
	} else {
		os.WriteFile(filepath.Join(dir, componentFile), content, 0644)
	}

	// Detect dependencies from imports
	deps := map[string]string{
		"react":     VersionReact,
		"react-dom": VersionReactDOM,
	}
	devDeps := map[string]string{
		"vite":                 VersionVite,
		"@vitejs/plugin-react": VersionVitePluginReact,
	}

	if isTS {
		devDeps["typescript"] = VersionTypeScript
		devDeps["@types/react"] = VersionTypesReact
		devDeps["@types/react-dom"] = VersionTypesReactDOM
	}

	// Scan ALL imports and add them as dependencies
	externalDeps := scanImports(src)
	for pkg, ver := range externalDeps {
		deps[pkg] = ver
	}

	// Check for Tailwind usage
	hasTailwind := strings.Contains(src, "className=")

	if hasTailwind {
		devDeps["tailwindcss"] = VersionTailwind
		devDeps["postcss"] = VersionPostCSS
		devDeps["autoprefixer"] = VersionAutoprefixer
	}

	// package.json
	depsJSON := mapToJSON(deps)
	devDepsJSON := mapToJSON(devDeps)

	pkgJSON := fmt.Sprintf(`{
  "name": "%s",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview"
  },
  "dependencies": {%s},
  "devDependencies": {%s}
}`, name, depsJSON, devDepsJSON)

	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0644)

	// index.html
	entryExt := ".jsx"
	if isTS {
		entryExt = ".tsx"
	}
	indexHTML := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>%s</title>
</head>
<body>
  <div id="root"></div>
  <script type="module" src="/main%s"></script>
</body>
</html>`, name, entryExt)

	os.WriteFile(filepath.Join(dir, "index.html"), []byte(indexHTML), 0644)

	// main entry
	mainContent := fmt.Sprintf(`import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './%s'
%s

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
)
`, componentFile, cssImport(hasTailwind))

	os.WriteFile(filepath.Join(dir, "main"+entryExt), []byte(mainContent), 0644)

	// vite.config
	viteConfig := `import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: { allowedHosts: true },
})
`
	os.WriteFile(filepath.Join(dir, "vite.config.js"), []byte(viteConfig), 0644)

	// Tailwind config if needed
	if hasTailwind {
		writeTailwindConfig(dir, []string{"./" + componentFile, "./main" + entryExt})
	}

	return &ScaffoldResult{
		Dir:       dir,
		Framework: fmt.Sprintf("React (%s) + Vite", lang),
		Cleanup:   func() { os.RemoveAll(dir) },
	}, nil
}
