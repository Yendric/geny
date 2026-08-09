package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Scaffolds a new geny site with Vite in the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		created := 0
		for _, file := range scaffold {
			if _, err := os.Stat(file.path); err == nil {
				color.Yellow("skipped %s (already exists)", file.path)
				continue
			}

			if dir := filepath.Dir(file.path); dir != "." {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					return fmt.Errorf("creating %s: %w", dir, err)
				}
			}
			if err := os.WriteFile(file.path, []byte(file.contents), 0o644); err != nil {
				return fmt.Errorf("creating %s: %w", file.path, err)
			}
			color.Green("created %s", file.path)
			created++
		}

		if err := os.MkdirAll("public", 0o755); err != nil {
			return fmt.Errorf("creating public: %w", err)
		}

		if created > 0 {
			fmt.Println()
			fmt.Println("Your site is ready. Next steps:")
			fmt.Println("  1. npm install")
			fmt.Println("  2. geny watch --serve")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

var scaffold = []struct {
	path     string
	contents string
}{
	{"geny.yaml", `vite:
  enabled: true
`},
	{"package.json", `{
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build"
  },
  "devDependencies": {
    "vite": "^7.1.0"
  }
}
`},
	{"vite.config.js", `import { resolve } from 'node:path'
import { defineConfig } from 'vite'

// Reloads the browser when geny regenerates the build directory.
const genyReload = {
  name: 'geny-reload',
  configureServer(server) {
    const buildDir = resolve('build')
    server.watcher.add(buildDir)

    let timer
    server.watcher.on('all', (event, file) => {
      if (!file.startsWith(buildDir)) return
      clearTimeout(timer)
      timer = setTimeout(() => server.ws.send({ type: 'full-reload' }), 50)
    })
  },
}

export default defineConfig({
  plugins: [genyReload],
  server: {
    // geny writes the configured dev server URL into its hot file
    strictPort: true,
  },
  build: {
    manifest: true,
    outDir: 'build',
    // geny copies public/ into build/ before running vite build.
    emptyOutDir: false,
    rollupOptions: {
      input: ['src/main.js'],
    },
  },
})
`},
	{"src/main.js", `import './style.css'

console.log('geny + vite is running')
`},
	{"src/style.css", `body {
  font-family: system-ui, sans-serif;
  max-width: 65ch;
  margin: 0 auto;
  padding: 2rem 1rem;
}
`},
	{"templates/default.html", `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>{{ .MetaData.title }}</title>
    {{ vite "src/main.js" }}
  </head>
  <body>
    <main>{{ .Content }}</main>
  </body>
</html>
`},
	{"content/index.md", `---
template: default
title: Welcome to geny
---

# Welcome to geny

Edit ` + "`content/index.md`" + ` to change this page.
`},
	{".gitignore", `/build
/node_modules
/.geny
`},
}
