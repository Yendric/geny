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
		for _, file := range scaffold() {
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

// the npm package ships the vite plugin, so it must match this binary
func genyPackageVersion() string {
	if version == "dev" {
		return "latest"
	}
	return "^" + version
}

type scaffoldFile struct {
	path     string
	contents string
}

func scaffold() []scaffoldFile {
	return []scaffoldFile{
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
  "dependencies": {
    "react": "^19.2.0",
    "react-dom": "^19.2.0"
  },
  "devDependencies": {
    "@types/react": "^19.2.0",
    "@types/react-dom": "^19.2.0",
    "@vitejs/plugin-react": "^5.1.0",
    "@yendric/geny": "` + genyPackageVersion() + `",
    "typescript": "^5.9.0",
    "vite": "^7.1.0"
  }
}
`},
		{"vite.config.js", `import { resolve } from 'node:path'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import geny from '@yendric/geny/vite'

// Reloads the browser when geny regenerates the build directory.
const genyReload = {
  name: 'geny-reload',
  configureServer(server) {
    const buildDir = resolve('build')
    server.watcher.add(buildDir)

    let timer
    server.watcher.on('all', (event, file) => {
      if (!file.startsWith(buildDir)) return
      if (event === 'unlink' || event === 'unlinkDir') return
      clearTimeout(timer)
      timer = setTimeout(() => server.ws.send({ type: 'full-reload' }), 150)
    })
  },
}

export default defineConfig({
  plugins: [react(), geny(), genyReload],
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
      input: ['src/style.css'],
    },
  },
})
`},
		{"tsconfig.json", `{
  "compilerOptions": {
    "target": "es2022",
    "module": "esnext",
    "moduleResolution": "bundler",
    "lib": ["es2022", "dom", "dom.iterable"],
    "types": ["vite/client"],
    "jsx": "react-jsx",
    "strict": true,
    "noEmit": true,
    "skipLibCheck": true
  },
  "include": ["src", "islands"]
}
`},
		{"islands/Counter.tsx", `import { useState } from 'react'

type Props = {
  start: number
}

export default function Counter({ start }: Props) {
  const [count, setCount] = useState(start)
  return (
    <button type="button" onClick={() => setCount(count + 1)}>
      Clicked {count} {count === 1 ? 'time' : 'times'}
    </button>
  )
}
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
    {{ vite "src/style.css" }}
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

<Counter client:load start={0} />
`},
		{".gitignore", `/build
/node_modules
/.geny
`},
	}
}
