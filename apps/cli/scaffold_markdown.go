package main

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
)

func scaffoldMarkdown(name string, content []byte) (*ScaffoldResult, error) {
	dir, err := os.MkdirTemp("", "openberth-markdown-*")
	if err != nil {
		return nil, err
	}

	escaped := html.EscapeString(string(content))

	indexHTML := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s</title>
  <style>
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
      line-height: 1.6;
      color: #24292e;
      max-width: 800px;
      margin: 0 auto;
      padding: 2rem 1rem;
    }
    h1, h2, h3, h4, h5, h6 { margin-top: 1.5em; margin-bottom: 0.5em; font-weight: 600; }
    h1 { font-size: 2em; border-bottom: 1px solid #eaecef; padding-bottom: 0.3em; }
    h2 { font-size: 1.5em; border-bottom: 1px solid #eaecef; padding-bottom: 0.3em; }
    a { color: #0366d6; text-decoration: none; }
    a:hover { text-decoration: underline; }
    code {
      background: #f6f8fa;
      padding: 0.2em 0.4em;
      border-radius: 3px;
      font-size: 85%%;
      font-family: SFMono-Regular, Consolas, "Liberation Mono", Menlo, monospace;
    }
    pre {
      background: #f6f8fa;
      padding: 1em;
      border-radius: 6px;
      overflow-x: auto;
    }
    pre code { background: none; padding: 0; font-size: 100%%; }
    blockquote {
      border-left: 4px solid #dfe2e5;
      margin: 0;
      padding: 0 1em;
      color: #6a737d;
    }
    table { border-collapse: collapse; width: 100%%; }
    th, td { border: 1px solid #dfe2e5; padding: 6px 13px; }
    th { background: #f6f8fa; font-weight: 600; }
    tr:nth-child(2n) { background: #f6f8fa; }
    img { max-width: 100%%; }
    hr { border: none; border-top: 1px solid #eaecef; margin: 1.5em 0; }
    ul, ol { padding-left: 2em; }
  </style>
</head>
<body>
  <div id="content"></div>
  <script id="md" type="text/markdown">%s</script>
  <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
  <script>
    document.getElementById('content').innerHTML = marked.parse(
      document.getElementById('md').textContent
    );
  </script>
</body>
</html>`, name, escaped)

	os.WriteFile(filepath.Join(dir, "index.html"), []byte(indexHTML), 0644)

	return &ScaffoldResult{
		Dir:       dir,
		Framework: "Markdown",
		Cleanup:   func() { os.RemoveAll(dir) },
	}, nil
}
