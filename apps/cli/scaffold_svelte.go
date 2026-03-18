package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func scaffoldSvelte(name string, content []byte) (*ScaffoldResult, error) {
	dir, err := os.MkdirTemp("", "openberth-svelte-*")
	if err != nil {
		return nil, err
	}

	os.WriteFile(filepath.Join(dir, "App.svelte"), content, 0644)

	src := string(content)
	hasTailwind := strings.Contains(src, "class=\"") && isTailwindClasses(src)

	deps := map[string]string{}
	// Scan imports in <script> section
	for pkg, ver := range scanImports(src) {
		deps[pkg] = ver
	}
	devDeps := map[string]string{
		"vite":                         VersionVite,
		"@sveltejs/vite-plugin-svelte": VersionVitePluginSvelte,
		"svelte":                       VersionSvelte,
	}
	if hasTailwind {
		devDeps["tailwindcss"] = VersionTailwind
		devDeps["postcss"] = VersionPostCSS
		devDeps["autoprefixer"] = VersionAutoprefixer
	}

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
}`, name, mapToJSON(deps), mapToJSON(devDeps))

	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0644)

	indexHTML := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>%s</title>
</head>
<body>
  <div id="app"></div>
  <script type="module" src="/main.js"></script>
</body>
</html>`, name)

	os.WriteFile(filepath.Join(dir, "index.html"), []byte(indexHTML), 0644)

	mainJS := fmt.Sprintf(`import App from './App.svelte'
%s

const app = new App({
  target: document.getElementById('app'),
})

export default app
`, cssImport(hasTailwind))

	os.WriteFile(filepath.Join(dir, "main.js"), []byte(mainJS), 0644)

	viteConfig := `import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

export default defineConfig({
  plugins: [svelte()],
  server: { allowedHosts: true },
})
`
	os.WriteFile(filepath.Join(dir, "vite.config.js"), []byte(viteConfig), 0644)

	// svelte.config.js required by the plugin
	svelteConfig := `import { vitePreprocess } from '@sveltejs/vite-plugin-svelte'
export default { preprocess: vitePreprocess() }
`
	os.WriteFile(filepath.Join(dir, "svelte.config.js"), []byte(svelteConfig), 0644)

	if hasTailwind {
		writeTailwindConfig(dir, []string{"./App.svelte"})
	}

	return &ScaffoldResult{
		Dir:       dir,
		Framework: "Svelte + Vite",
		Cleanup:   func() { os.RemoveAll(dir) },
	}, nil
}
