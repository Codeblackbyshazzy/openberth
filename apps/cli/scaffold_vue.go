package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func scaffoldVue(name string, content []byte) (*ScaffoldResult, error) {
	dir, err := os.MkdirTemp("", "openberth-vue-*")
	if err != nil {
		return nil, err
	}

	os.WriteFile(filepath.Join(dir, "App.vue"), content, 0644)

	src := string(content)
	hasTailwind := strings.Contains(src, "class=\"") && isTailwindClasses(src)

	deps := map[string]string{"vue": VersionVue}
	// Scan imports in <script> section
	for pkg, ver := range scanImports(src) {
		deps[pkg] = ver
	}
	devDeps := map[string]string{
		"vite":               VersionVite,
		"@vitejs/plugin-vue": VersionVitePluginVue,
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

	mainJS := fmt.Sprintf(`import { createApp } from 'vue'
import App from './App.vue'
%s

createApp(App).mount('#app')
`, cssImport(hasTailwind))

	os.WriteFile(filepath.Join(dir, "main.js"), []byte(mainJS), 0644)

	viteConfig := `import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: { allowedHosts: true },
})
`
	os.WriteFile(filepath.Join(dir, "vite.config.js"), []byte(viteConfig), 0644)

	if hasTailwind {
		writeTailwindConfig(dir, []string{"./App.vue"})
	}

	return &ScaffoldResult{
		Dir:       dir,
		Framework: "Vue + Vite",
		Cleanup:   func() { os.RemoveAll(dir) },
	}, nil
}
