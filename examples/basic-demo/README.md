# BrowserGit Basic Demo

A simple demonstration of BrowserGit's core features running in the browser.

## Features

This demo showcases:

- Repository initialization
- File creation and reading
- Git add and commit operations
- Branch creation and switching
- Commit history viewing
- Real-time status updates

## Running the Demo

```bash
# Install dependencies
yarn install

# Start development server
yarn dev

# Open browser to http://localhost:3000
```

## Building for Production

```bash
# Build the demo
yarn build

# Preview production build
yarn preview
```

## Usage

1. Click "Initialize Repository" to create a new Git repository
2. Create or edit files using the file operations section
3. Stage files with "Add Files"
4. Commit changes with "Commit Changes"
5. Create and switch between branches
6. View commit history

## Storage

The demo uses IndexedDB for storage by default, so your repository persists across page reloads.

## Learn More

- [BrowserGit Documentation](../../docs/README.md)
- [API Reference](../../docs/api-reference/repository.md)
