# jsawn

A fast, interactive JSON viewer for the terminal. Navigate large JSON documents with vim-style keybindings, collapsible tree nodes, syntax highlighting, search, schema generation, and more.

![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/github/license/aryankumar07/jsawn)
![GitHub Release](https://img.shields.io/github/v/release/aryankumar07/jsawn)

## Installation

### Homebrew (macOS / Linux)

```bash
brew tap aryankumar07/tap
brew install jsawn
```

### AUR (Arch Linux)

```bash
yay -S jsawn-bin
```

### From source

```bash
go install github.com/aryankumar07/jsawn@latest
```

### Manual

Download the latest binary from [GitHub Releases](https://github.com/aryankumar07/jsawn/releases).

## Usage

```bash
cat data.json | jsawn              # pipe from stdin
jsawn data.json                    # open a local file
jsawn https://api.example.com/users  # fetch JSON from a URL
jsawn a.json,b.json,<url>         # multiple sources as tabs
```

### HTTP Options

When fetching from URLs, you can customize the request:

```bash
# Custom headers
jsawn -H 'Authorization: Bearer <token>' https://api.example.com/users

# POST with JSON body (auto-sets Content-Type: application/json)
jsawn -X POST -d '{"name":"foo"}' https://api.example.com/users

# Multiple headers
jsawn -H 'Authorization: Bearer <token>' -H 'Accept: application/json' https://api.example.com/users
```

| Flag | Description |
|---|---|
| `-H 'Key: Value'` | Add an HTTP header (repeatable) |
| `-X METHOD` | HTTP method (default: `GET`) |
| `-d '{...}'` | Request body (auto-sets `Content-Type: application/json`) |

### Multiple Sources

Comma-separate files and URLs to open them as tabs:

```bash
jsawn users.json,https://api.example.com/posts,config.json
```

Each source gets its own tab with independent navigation and state. Switch between them with `Tab` / `Shift+Tab`.

### Examples

```bash
# Explore a local JSON file
jsawn package.json

# Pipe output from any command
curl -s https://jsonplaceholder.typicode.com/todos | jsawn
docker inspect <container> | jsawn
kubectl get pods -o json | jsawn
aws s3api list-buckets | jsawn

# Inspect an API response with auth
jsawn -H 'Authorization: Bearer <token>' https://api.github.com/user

# Compare local config with remote
jsawn config.json,https://api.example.com/config

# POST and view the response
jsawn -X POST -d '{"query":"users"}' https://api.example.com/graphql
```

## Keybindings

### Navigation

| Key | Action |
|---|---|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `h` / `←` | Collapse node / jump to parent |
| `l` / `→` | Expand node / move to first child |
| `Space` | Toggle collapse/expand |
| `gg` | Jump to top |
| `G` | Jump to bottom |

### Folding

| Key | Action |
|---|---|
| `e` | Expand all nodes |
| `E` | Collapse all nodes |
| `1`–`9` | Collapse all nodes at depth >= N |

### Search

| Key | Action |
|---|---|
| `/` | Start search |
| `Enter` | Execute search |
| `n` | Next match |
| `N` | Previous match |
| `Esc` | Clear search |
| `Ctrl+u` | Clear search query |

### Views & Commands

| Key | Action |
|---|---|
| `f` | Toggle flat view (leaf nodes with full paths) |
| `:` | Enter command mode |
| `Tab` / `Shift+Tab` | Switch tabs (multi-source) |

### Schema Overlay

| Key | Action |
|---|---|
| `j` / `k` | Scroll up/down |
| `d` / `u` | Page down/up |
| `G` / `g` | Jump to end/start |
| `y` | Copy to clipboard |
| `q` / `Esc` | Close overlay |

### Quit

`q` / `Esc` / `Ctrl+c`

## Commands

Enter command mode by pressing `:`, then type a command:

| Command | Description |
|---|---|
| `schema go` | Generate a Go struct from the JSON |
| `schema ts` | Generate a TypeScript interface |
| `schema zod` | Generate a Zod schema |

Schema results open in a scrollable overlay that can be copied to clipboard with `y`.

## Features

### Syntax Highlighting

- **Keys** — cyan
- **Strings** — green
- **Numbers** — yellow
- **Booleans** — magenta
- **Null** — red

### Collapsed Summaries

Collapsed objects and arrays show a summary like `{... 3 keys}` or `[... 5 items]`.

### Breadcrumbs

The footer displays the jq-style path to the current node (e.g. `.users[0].name`), along with the scroll position percentage.

### Flat View

Press `f` to flatten the tree into a list of all leaf nodes with their full paths — useful for quickly scanning values in deeply nested JSON.

### Search

Case-insensitive search with highlighted matches across keys and values. Cycle through matches with `n` / `N`.

### Mouse Support

Scroll with the mouse wheel.

## License

[MIT](LICENSE)
