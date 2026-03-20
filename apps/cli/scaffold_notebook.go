package main

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
)

type notebook struct {
	Cells    []nbCell `json:"cells"`
	Metadata struct {
		Kernelspec struct {
			Language string `json:"language"`
		} `json:"kernelspec"`
	} `json:"metadata"`
	// nbformat 3 wraps cells inside worksheets
	Worksheets []struct {
		Cells []nbCell `json:"cells"`
	} `json:"worksheets"`
}

type nbCell struct {
	CellType string      `json:"cell_type"`
	Source   interface{} `json:"source"`
	Input    interface{} `json:"input"`  // nbformat 3 uses "input" instead of "source" for code cells
	Level    int         `json:"level"`  // nbformat 3 heading level
	Outputs  []nbOutput  `json:"outputs"`
}

type nbOutput struct {
	OutputType string                 `json:"output_type"`
	Text       interface{}            `json:"text"`
	Data       map[string]interface{} `json:"data"`
	Ename      string                 `json:"ename"`
	Evalue     string                 `json:"evalue"`
}

// joinSource normalises a notebook source field (string or []string) into a single string.
func joinSource(v interface{}) string {
	switch s := v.(type) {
	case string:
		return s
	case []interface{}:
		var sb strings.Builder
		for _, line := range s {
			if str, ok := line.(string); ok {
				sb.WriteString(str)
			}
		}
		return sb.String()
	}
	return ""
}

func scaffoldNotebook(name string, content []byte) (*ScaffoldResult, error) {
	dir, err := os.MkdirTemp("", "openberth-notebook-*")
	if err != nil {
		return nil, err
	}

	var nb notebook
	if err := json.Unmarshal(content, &nb); err != nil {
		os.RemoveAll(dir)
		return nil, fmt.Errorf("invalid notebook JSON: %w", err)
	}

	// nbformat 3 stores cells inside worksheets; merge them into top-level cells
	if len(nb.Cells) == 0 {
		for _, ws := range nb.Worksheets {
			nb.Cells = append(nb.Cells, ws.Cells...)
		}
	}

	if len(nb.Cells) == 0 {
		os.RemoveAll(dir)
		return nil, fmt.Errorf("notebook has no cells")
	}

	// Detect notebook language for syntax highlighting
	lang := strings.ToLower(nb.Metadata.Kernelspec.Language)
	if lang == "" {
		lang = "python" // vast majority of notebooks
	}

	var body strings.Builder
	for i, cell := range nb.Cells {
		// nbformat 3 uses "input" for code cells; fall back to it if "source" is empty
		src := joinSource(cell.Source)
		if src == "" {
			src = joinSource(cell.Input)
		}
		switch cell.CellType {
		case "markdown":
			escaped := html.EscapeString(src)
			fmt.Fprintf(&body, "<div class=\"cell cell-md\" data-cell=\"%d\"><textarea class=\"nb-md\" style=\"display:none\">%s</textarea></div>\n", i, escaped)
		case "heading":
			// nbformat 3 heading cell — convert to markdown heading
			level := cell.Level
			if level < 1 || level > 6 {
				level = 1
			}
			md := strings.Repeat("#", level) + " " + src
			escaped := html.EscapeString(md)
			fmt.Fprintf(&body, "<div class=\"cell cell-md\" data-cell=\"%d\"><textarea class=\"nb-md\" style=\"display:none\">%s</textarea></div>\n", i, escaped)
		case "code":
			fmt.Fprintf(&body, "<div class=\"cell cell-code\" data-cell=\"%d\">\n", i)
			fmt.Fprintf(&body, "  <div class=\"cell-label\">In [%d]:</div>\n", i+1)
			fmt.Fprintf(&body, "  <pre><code class=\"language-%s\">%s</code></pre>\n", lang, html.EscapeString(src))
			for _, out := range cell.Outputs {
				renderOutput(&body, out, i)
			}
			fmt.Fprintf(&body, "</div>\n")
		default:
			// raw cells — render as preformatted text
			if src != "" {
				fmt.Fprintf(&body, "<div class=\"cell\" data-cell=\"%d\"><pre>%s</pre></div>\n", i, html.EscapeString(src))
			}
		}
	}

	var page strings.Builder
	page.WriteString(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>`)
	page.WriteString(html.EscapeString(name))
	page.WriteString(`</title>
  <style>
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
      line-height: 1.6;
      color: #24292e;
      max-width: 900px;
      margin: 0 auto;
      padding: 2rem 1rem;
      background: #fff;
    }
    .cell { margin-bottom: 1rem; }
    .cell-code {
      border: 1px solid #e1e4e8;
      border-radius: 6px;
      overflow: hidden;
    }
    .cell-label {
      background: #f6f8fa;
      padding: 4px 12px;
      font-size: 12px;
      color: #6a737d;
      font-family: SFMono-Regular, Consolas, "Liberation Mono", Menlo, monospace;
      border-bottom: 1px solid #e1e4e8;
    }
    .cell-code pre {
      margin: 0;
      padding: 12px;
      background: #f8f9fa;
      overflow-x: auto;
    }
    .cell-code code {
      font-family: SFMono-Regular, Consolas, "Liberation Mono", Menlo, monospace;
      font-size: 14px;
    }
    .output {
      padding: 8px 12px;
      border-top: 1px solid #e1e4e8;
    }
    .output-label {
      font-size: 12px;
      color: #6a737d;
      font-family: SFMono-Regular, Consolas, "Liberation Mono", Menlo, monospace;
      margin-bottom: 4px;
    }
    .output pre {
      margin: 0;
      font-family: SFMono-Regular, Consolas, "Liberation Mono", Menlo, monospace;
      font-size: 14px;
      white-space: pre-wrap;
    }
    .output-error pre {
      color: #d73a49;
    }
    .output img { max-width: 100%; }
    /* Markdown cell styling */
    .cell-md h1, .cell-md h2, .cell-md h3, .cell-md h4 { margin-top: 1em; margin-bottom: 0.5em; font-weight: 600; }
    .cell-md h1 { font-size: 2em; border-bottom: 1px solid #eaecef; padding-bottom: 0.3em; }
    .cell-md h2 { font-size: 1.5em; border-bottom: 1px solid #eaecef; padding-bottom: 0.3em; }
    .cell-md a { color: #0366d6; }
    .cell-md code {
      background: #f6f8fa; padding: 0.2em 0.4em; border-radius: 3px;
      font-size: 85%; font-family: SFMono-Regular, Consolas, "Liberation Mono", Menlo, monospace;
    }
    .cell-md pre { background: #f6f8fa; padding: 1em; border-radius: 6px; overflow-x: auto; }
    .cell-md pre code { background: none; padding: 0; font-size: 100%; }
    .cell-md blockquote { border-left: 4px solid #dfe2e5; margin: 0; padding: 0 1em; color: #6a737d; }
    .cell-md table { border-collapse: collapse; }
    .cell-md th, .cell-md td { border: 1px solid #dfe2e5; padding: 6px 13px; }
    .cell-md th { background: #f6f8fa; }
    .cell-md img { max-width: 100%; }
  </style>
</head>
<body>
  `)
	page.WriteString(body.String())
	page.WriteString(`
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11/build/styles/github.min.css">
  <script src="https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11/build/highlight.min.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
  <script>
    document.querySelectorAll('.nb-md').forEach(function(el) {
      var div = document.createElement('div');
      div.innerHTML = marked.parse(el.value);
      el.parentNode.appendChild(div);
    });
    hljs.highlightAll();
  </script>
</body>
</html>`)
	indexHTML := page.String()

	os.WriteFile(filepath.Join(dir, "index.html"), []byte(indexHTML), 0644)

	return &ScaffoldResult{
		Dir:       dir,
		Framework: "Jupyter Notebook",
		Cleanup:   func() { os.RemoveAll(dir) },
	}, nil
}

func renderOutput(body *strings.Builder, out nbOutput, cellIdx int) {
	switch out.OutputType {
	case "stream", "execute_result":
		text := ""
		if out.Text != nil {
			text = joinSource(out.Text)
		}
		if out.Data != nil {
			if plain, ok := out.Data["text/plain"]; ok {
				text = joinSource(plain)
			}
		}
		if text != "" {
			fmt.Fprintf(body, "  <div class=\"output\"><div class=\"output-label\">Out [%d]:</div><pre>%s</pre></div>\n", cellIdx+1, html.EscapeString(text))
		}
	case "display_data":
		if out.Data == nil {
			return
		}
		// Prefer HTML, then image, then plain text
		if htmlData, ok := out.Data["text/html"]; ok {
			fmt.Fprintf(body, "  <div class=\"output\">%s</div>\n", joinSource(htmlData))
		} else if pngData, ok := out.Data["image/png"]; ok {
			if s, ok := pngData.(string); ok {
				fmt.Fprintf(body, "  <div class=\"output\"><img src=\"data:image/png;base64,%s\"></div>\n", s)
			}
		} else if plain, ok := out.Data["text/plain"]; ok {
			fmt.Fprintf(body, "  <div class=\"output\"><pre>%s</pre></div>\n", html.EscapeString(joinSource(plain)))
		}
	case "error":
		fmt.Fprintf(body, "  <div class=\"output output-error\"><pre>%s: %s</pre></div>\n", html.EscapeString(out.Ename), html.EscapeString(out.Evalue))
	}
}
