# Offline Docs - Version Controlled Documentation

A complete documentation system with built-in version control powered by BrowserGit.

## Features

- **Markdown Editing**: Full Markdown support with live preview
- **Version Control**: Every change is tracked with Git
- **Version History**: View and restore any previous version
- **Offline First**: Works completely offline with local storage
- **Search**: Search across all documentation (planned)

## Running the Demo

```bash
# Install dependencies
yarn install

# Start development server
yarn dev

# Open browser to http://localhost:3002
```

## Building for Production

```bash
yarn build
yarn preview
```

## Usage

### Creating Documents

1. Click "New Document"
2. Enter a document name ending with `.md`
3. Write your content in Markdown
4. Click "Save & Commit" to save changes

### Editing Documents

1. Select a document from the sidebar
2. Click "Edit" to enter edit mode
3. Make your changes
4. Click "Save & Commit" and enter a commit message
5. Click "Preview" to see the rendered Markdown

### Version History

1. Select a document
2. View the version history in the right panel
3. Click any commit to restore that version
4. Confirm the restoration

## Markdown Support

Full Markdown syntax is supported including:

- Headers
- Lists (ordered and unordered)
- Code blocks with syntax highlighting
- Links and images
- Tables
- Blockquotes
- And more!

## Storage

Documents are stored using IndexedDB and persist across sessions. All changes are tracked with full Git history.

## Use Cases

- Personal documentation
- Team knowledge base
- Project documentation
- Technical specifications
- API documentation
- User guides

## Learn More

- [Marked.js - Markdown Parser](https://marked.js.org/)
- [BrowserGit Documentation](../../docs/README.md)
- [API Reference](../../docs/api-reference/repository.md)
