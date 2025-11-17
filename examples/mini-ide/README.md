# BrowserGit Mini IDE

A minimal IDE with full Git integration powered by BrowserGit.

## Features

- **File Management**: Create, edit, and save files
- **Git Integration**: Stage, commit, and view history
- **Branch Management**: View current branch and switch branches
- **Real-time Status**: See modified and untracked files
- **Commit History**: View recent commits

## Running the Demo

```bash
# Install dependencies
yarn install

# Start development server
yarn dev

# Open browser to http://localhost:3001
```

## Building for Production

```bash
yarn build
yarn preview
```

## Usage

1. The IDE initializes a repository automatically
2. Click "New File" to create files
3. Edit files in the central editor
4. Click "Save" to persist changes
5. Use the Git panel to stage and commit changes
6. View commit history in the right panel

## Technology Stack

- React 18
- TypeScript
- Vite
- BrowserGit

## Learn More

- [BrowserGit Documentation](../../docs/README.md)
- [API Reference](../../docs/api-reference/repository.md)
