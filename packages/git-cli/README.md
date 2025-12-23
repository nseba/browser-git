# @browser-git/git-cli

Command-line interface for BrowserGit - test and interact with browser-git from the command line.

## Installation

```bash
npm install -g @browser-git/git-cli
# or
yarn global add @browser-git/git-cli
```

## Usage

The CLI provides a `bgit` command (aliased as `browser-git`) that mirrors standard Git commands:

### Repository Operations

```bash
# Initialize a new repository
bgit init [path] [options]
  --bare                    Create a bare repository
  --initial-branch <name>   Initial branch name (default: main)
  --hash <algorithm>        Hash algorithm: sha1 or sha256 (default: sha1)
  --storage <backend>       Storage backend: indexeddb, opfs, localstorage, memory (default: indexeddb)

# Clone a repository
bgit clone <url> [directory] [options]
  --depth <depth>           Create shallow clone
  --branch <branch>         Checkout specific branch
  --bare                    Create bare repository
  --storage <backend>       Storage backend
  --username <username>     Username for authentication
  --token <token>           Token for authentication
```

### Basic Workflow

```bash
# Check status
bgit status [options]
  -s, --short              Show short format
  -b, --branch             Show branch information

# Add files to index
bgit add <paths...> [options]
  -A, --all                Add all changes
  -u, --update             Update tracked files
  -f, --force              Allow adding ignored files

# Commit changes
bgit commit [options]
  -m, --message <msg>      Commit message
  -a, --all                Commit all changed files
  --author <author>        Override author (format: "Name <email>")
  --amend                  Amend previous commit
  --allow-empty            Allow empty commit
```

### Branching

```bash
# List branches
bgit branch [options]
  -a, --all                List all branches
  -r, --remote             List remote branches

# Create branch
bgit branch <name>

# Delete branch
bgit branch -d <name>      # Safe delete
bgit branch -D <name>      # Force delete

# Rename branch
bgit branch -m <old> <new>

# Switch branches
bgit checkout <branch> [options]
  -b, --create             Create and checkout new branch
  -B, --force-create       Create/reset and checkout branch
  -f, --force              Force checkout (discard changes)
```

### History

```bash
# View commit log
bgit log [options]
  --oneline                Show condensed format
  --graph                  Show commit graph
  -n, --max-count <num>    Limit number of commits (default: 10)
  --author <pattern>       Filter by author
  --since <date>           Show commits since date
  --until <date>           Show commits until date
  --grep <pattern>         Filter by message

# View changes
bgit diff [paths...] [options]
  --cached, --staged       Show diff between index and HEAD
  --stat                   Show diffstat instead of patch
  -U, --unified <lines>    Number of context lines (default: 3)
```

### Merging

```bash
# Merge branches
bgit merge <branch> [options]
  --no-ff                  Create merge commit (no fast-forward)
  --ff-only                Only allow fast-forward
  -m, --message <msg>      Merge commit message
  --abort                  Abort current merge
```

### Remote Operations

```bash
# Fetch from remote
bgit fetch [remote] [refspec] [options]
  --all                    Fetch from all remotes
  --prune                  Remove stale remote-tracking refs
  --depth <depth>          Deepen shallow clone
  --username <username>    Username for authentication
  --token <token>          Token for authentication

# Pull from remote
bgit pull [remote] [branch] [options]
  --rebase                 Rebase instead of merge
  --ff-only                Only fast-forward
  --no-ff                  Create merge commit
  --username <username>    Username for authentication
  --token <token>          Token for authentication

# Push to remote
bgit push [remote] [refspec] [options]
  -f, --force              Force push (may lose commits)
  --all                    Push all branches
  --tags                   Push all tags
  --delete                 Delete remote branch
  --set-upstream           Set upstream for current branch
  --username <username>    Username for authentication
  --token <token>          Token for authentication
```

## Examples

```bash
# Initialize a repository
bgit init my-project
cd my-project

# Create and edit files
echo "Hello World" > README.md

# Stage and commit
bgit add README.md
bgit commit -m "Initial commit"

# Create and switch to a new branch
bgit checkout -b feature/new-feature

# View status and history
bgit status
bgit log --oneline

# Clone a repository with authentication
bgit clone https://github.com/user/repo.git \
  --username myuser \
  --token ghp_xxxxxxxxxxxx

# Fetch and merge changes
bgit fetch origin
bgit merge origin/main

# Or pull in one command
bgit pull origin main --username myuser --token ghp_xxxxxxxxxxxx
```

## Authentication

For operations that require authentication (clone, fetch, pull, push), you can provide credentials via:

- `--username` and `--token` flags
- Environment variables (planned)
- Credential storage (planned)

For GitHub, use a Personal Access Token instead of your password:

1. Go to Settings → Developer settings → Personal access tokens
2. Generate new token with `repo` scope
3. Use token with `--token` flag

## Storage Backends

BrowserGit supports multiple storage backends:

- **indexeddb** (default): Best for most use cases, good performance and storage quota
- **opfs**: Origin Private File System, best performance but limited browser support
- **localstorage**: Fallback option, limited storage (5-10MB)
- **memory**: In-memory only, useful for testing

Specify backend with `--storage` flag:

```bash
bgit init --storage opfs
bgit clone <url> --storage indexeddb
```

## Limitations

- The CLI runs in Node.js but uses browser storage APIs through polyfills
- Some operations may be slower than native Git
- Storage quota limits apply based on backend

## Development

```bash
# Build the CLI
yarn build

# Run locally
node dist/cli.js <command>

# Or install locally
yarn link
bgit <command>
```

## License

MIT
