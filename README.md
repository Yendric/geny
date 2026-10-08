# Geny

![Issues](https://img.shields.io/github/issues/Yendric/geny)

A small and efficient static site generator written in Go.

## Installation

### npm (recommended)

If your site uses npm (which it does if you use the Vite integration), install geny as a dev dependency:

```sh
npm install --save-dev @yendric/geny
```

Then run it with `npx geny`, or from package.json scripts:

```json
"scripts": {
  "build": "geny build"
}
```

### Prebuilt binaries

Download the latest release from the [releases page](https://github.com/Yendric/geny/releases) (linux/macOS/windows, amd64/arm64).

### go install

```sh
go install github.com/Yendric/geny@latest
```

### Building from source

> You need to have Golang installed in order to build Geny from source.

1. Clone: `git clone git@github.com:Yendric/geny.git`
2. Build: `go build`
3. Run: `./geny`

## Geny

Geny is a very simple static site generator written in golang. It's goal is to be very efficient and easy to use. It uses gohtml for templating, and markdown for content.\
https://yendric.be is made entirely using Geny.

## Usage instructions

### Defining templates

Templates are defined by creating a new file in the `templates` directory. The filename should be: `<name>.html`, and the file should contain valid gohtml code.
Per convention, you should put shared templates in the `templates/shared` directory.
This is only a convention, you are free to put templates anywhere you want, as long as they are in the `templates` directory and not deeper than 1 level.

#### Accessing data

The current contentFile can be accessed using:

- `{{ .Content }}` for its content as html
- `{{ .RawContent }}` for its content as markdown
- `{{ .Url }}` to link to its page
- `{{ .MetaData }}` for metadata, which you can define yourself (see [creating content](#creating-content))
- `{{ .Template }}` for info about the template (eg. `.Template.Name`)
- `{{ .Collections }}` to access all other collections.\
  A collection is a slice of contentFiles belonging to the same template.\
  For example, if you have a template called `post`, you can access all posts using `{{ .Collections.post }}`
- `{{ .Path }}` for the filepath
- `{{ .FileName }}` for the filename

#### Utility functions

Geny provides a few utility functions that can be used in templates:

- `Truncate(string)` truncates a string to 150 characters.
- `StripTags(string)` removes all html tags from a string.
- `GetCurrentYear()` returns the current year.

Collections also have helper methods on them:

- `Collection.Slice(start, end)` returns a slice of the collection.
- `Collection.SortByDate()` returns a new collection, sorted by date.

### Creating content and routes

#### Defining routes

Routes are entirely defined by the directory and file structure in the `content` directory.

For example: `content/posts/hello-world.md` will be available at `/posts/hello-world`.
The main page for a directory is defined by the `index.md` file in that directory.

It is possible to structure content in directories without defining new routes, by prefixing the directory name with an underscore.
For example: `content/_posts/hello-world.md` will be available at `/hello-world`.

#### Special routes

As mentioned above, the `index.md` file in a directory defines the main page for that directory.

If you name a file 404.md, it will be rendered as `404.html` instead of `404/index.html`.
This way, services like cloudflare pages can automatically detect the 404 page.

#### Creating content

Content files use markdown and should start with a header, which is defined by a yaml block surrounded by `---` at the top of the file. There is one required field: `template`. This field should contain the name of the template to use for rendering the content (without .html).

Other fields can be defined as desired, and can be referenced in templates using `{{ .MetaData.<field> }}`

Example:

```yaml
---
template: index
title: Yendrics blog - Home
---
# Content
```

PS: the markdown renderer also supports styled code blocks.

### Generating your site

To generate your site, run `geny build` in the root of the project. This will generate a `build` directory containing the generated site.\
You can also watch for changes using `geny watch`.\
Both commands accept an optional `--serve` flag, which will start a local server on port 8080 to serve the generated site until stopped. The port can be changed using the `--port` flag.

Example: `geny build --serve --port 3000`

### Dealing with css and static content

CSS and other static files can be placed in the `public` directory. These files will be copied to the root of the generated site.

### Using Vite for js and css

Geny integrates with [Vite](https://vite.dev) for bundling js and css. The easiest way to get started is running `geny init` in an empty directory, which scaffolds a working site (run `npm install` once afterwards). To enable it on an existing site, add the following to `geny.yaml`:

```yaml
vite:
  enabled: true
```

Then load your entry points in a template using the `vite` function:

```html
{{ vite "src/main.js" }}
```

Pass all entry points for a page to a single `vite` call (it is variadic, every call emits its own dev-server client tag).

During `geny watch`, geny starts the Vite dev server alongside it and the tag points at it, giving you hot module replacement for js/css and automatic browser reloads when content or templates change. During `geny build`, geny runs `vite build` and the tag resolves to the hashed script and stylesheet files from the build manifest.

The commands used, the dev server URL and the hot file location can be overridden in the `vite` section of `geny.yaml` (defaults shown):

```yaml
vite:
  enabled: true
  buildCommand: npm run build
  devCommand: npm run dev
  devServerURL: http://localhost:5173
  hotFile: .geny/hot
```

### Islands

Islands are interactive React components placed on otherwise static pages. Only pages that use an island load JavaScript for it, and each island loads its own code when it hydrates.

#### Setup

Islands need the Vite integration plus React. `geny init` sets this up.
If you prefer to manually set this up, you can do so as follows:

```sh
npm install react react-dom
npm install --save-dev @yendric/geny @vitejs/plugin-react typescript @types/react @types/react-dom
```

```js
// vite.config.js
import react from "@vitejs/plugin-react";
import geny from "@yendric/geny/vite";

export default defineConfig({
  plugins: [react(), geny()],
  // ...
});
```

#### Writing an island

To create an island, simply put your components in the `islands` directory as a `.tsx` or `.jsx`, with a default export.
The file path is the island name (eg. `islands/ui/Button.tsx` is `ui/Button`).
Do not use an HTML element name for your island, as it will not work.

```tsx
// islands/Counter.tsx
import { useState } from "react";

type Props = { start: number };

export default function Counter({ start }: Props) {
  const [count, setCount] = useState(start);
  return (
    <button onClick={() => setCount(count + 1)}>Clicked {count} times</button>
  );
}
```

#### Using islands in markdown

Any tag whose name matches an island is rendered as that island:

```md
<Counter client:visible start={5} label="Clicks" />

<Counter start={0}>
Loading counter...
</Counter>
```

- `"text"` or `'text'` passes a string, `{...}` passes JSON (`{5}`, `{[1, 2]}`, `{{"a": 1}}`) and a bare attribute passes `true`.
- Content between the opening and closing tag is rendered as markdown and shown until the island mounts. A tag with content must start on its own line, and the closing tag must be on its own line.
- Inside a paragraph, islands must be self-closing.

#### Using islands in templates

In go templates we use a builder pattern to render the islands, with a final render call at the end.

```html
{{ island "Counter" | render }} {{ island "Counter" | props "start" 5 | props
"label" .MetaData.title | client "visible" | fallback (include
"counter-fallback" .) | render }}
```

Just like in markdown, you can also add a fallback: `include` renders a named template to HTML, so a fallback can be defined with `{{ define "counter-fallback" }}...{{ end }}`.

#### Hydration

- `client:load` (default) mounts the island as soon as the page loads.
- `client:visible` mounts the island once it scrolls into view.

#### Type checking props

geny checks the props of every island usage against the component's `Props` type with `tsc`. During `geny build`, type errors fail the build and point to the file and line of the usage. During `geny watch`, they are printed as warnings. The check uses `tsconfig.json` at the site root when present.

#### Configuration

The islands directory can be changed in `geny.yaml`:

```yaml
islandsDir: islands
```
