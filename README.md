# Geny

![Issues](https://img.shields.io/github/issues/Yendric/geny)

A small and efficient static site generator written in Go.

## Installation

You can install geny by downloading the latest release from the [releases page](https://github.com/Yendric/geny/releases).

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

- `{{ .Content }}` for it's content (as html)
- `{{ .RawContent }}` for it's content (as markdown)
- `{{ .Url }}` to link to it's page
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

Content files use markdown should start with a header, which is defined by a yaml block surrounded by `---` at the top of the file. There is one required field: `template`. It field should contain the name of the template to use for rendering the content (without .html).

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

It it possible to configure a build step in the `--run` flag.\
For example: `geny build --run "npm run build"` will run `npm run build` in the root of the project before generating the site.
