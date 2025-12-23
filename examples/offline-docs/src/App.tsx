import React, { useState, useEffect } from "react";
import { Repository } from "@browser-git/browser-git";
import { marked } from "marked";
import DocumentList from "./components/DocumentList";
import MarkdownEditor from "./components/MarkdownEditor";
import VersionHistory from "./components/VersionHistory";
import "./App.css";

const App: React.FC = () => {
  const [repo, setRepo] = useState<Repository | null>(null);
  const [docs, setDocs] = useState<string[]>([]);
  const [currentDoc, setCurrentDoc] = useState<string | null>(null);
  const [content, setContent] = useState<string>("");
  const [isEditing, setIsEditing] = useState<boolean>(false);

  useEffect(() => {
    initializeRepo();
  }, []);

  const initializeRepo = async () => {
    try {
      let repository: Repository;
      try {
        repository = await Repository.open("/offline-docs");
      } catch {
        repository = await Repository.init("/offline-docs", {
          storage: "indexeddb",
          initialBranch: "main",
        });

        // Create initial documentation
        await repository.fs.writeFile("README.md", getInitialReadme());
        await repository.add(["README.md"]);
        await repository.commit("Initial documentation", {
          author: { name: "Docs Admin", email: "admin@docs.local" },
        });
      }
      setRepo(repository);
      await refreshDocs(repository);
    } catch (error) {
      console.error("Failed to initialize repository:", error);
    }
  };

  const getInitialReadme = () => `# Welcome to Offline Docs

This is a version-controlled documentation system powered by BrowserGit.

## Features

- **Markdown Editing**: Write documentation in Markdown
- **Version Control**: Every change is tracked with Git
- **Offline First**: Works completely offline
- **Version History**: View and restore previous versions

## Getting Started

1. Click "New Document" to create a new page
2. Edit your content in Markdown
3. Click "Save & Commit" to save your changes
4. View version history to see all changes

## Example Content

### Code Example

\`\`\`javascript
console.log('Hello, BrowserGit!');
\`\`\`

### Lists

- Feature 1
- Feature 2
- Feature 3

### Links

[BrowserGit Documentation](https://github.com/yourusername/browser-git)
`;

  const refreshDocs = async (repository: Repository) => {
    try {
      const files = await repository.fs.readdir("/");
      const mdFiles = files.filter((f) => f.endsWith(".md"));
      setDocs(mdFiles);
    } catch (error) {
      console.error("Failed to read docs:", error);
    }
  };

  const handleDocSelect = async (docName: string) => {
    if (!repo) return;

    try {
      const docContent = await repo.fs.readFile(docName, "utf-8");
      setCurrentDoc(docName);
      setContent(docContent);
      setIsEditing(false);
    } catch (error) {
      console.error("Failed to read document:", error);
    }
  };

  const handleNewDoc = async () => {
    const docName = prompt("Enter document name (e.g., guide.md):");
    if (!docName || !repo) return;

    if (!docName.endsWith(".md")) {
      alert("Document name must end with .md");
      return;
    }

    try {
      const template = `# ${docName.replace(".md", "")}\n\nYour content here...`;
      await repo.fs.writeFile(docName, template);
      await refreshDocs(repo);
      setCurrentDoc(docName);
      setContent(template);
      setIsEditing(true);
    } catch (error) {
      alert(`Failed to create document: ${(error as Error).message}`);
    }
  };

  const handleSave = async () => {
    if (!repo || !currentDoc) return;

    const commitMsg = prompt("Commit message:", `Update ${currentDoc}`);
    if (!commitMsg) return;

    try {
      await repo.fs.writeFile(currentDoc, content);
      await repo.add([currentDoc]);
      await repo.commit(commitMsg, {
        author: { name: "Docs User", email: "user@docs.local" },
      });
      setIsEditing(false);
      alert("Document saved and committed!");
    } catch (error) {
      alert(`Failed to save: ${(error as Error).message}`);
    }
  };

  const renderMarkdown = (markdown: string): string => {
    try {
      return marked(markdown) as string;
    } catch (error) {
      return "<p>Error rendering markdown</p>";
    }
  };

  return (
    <div className="app">
      <header className="app-header">
        <h1>📚 Offline Docs</h1>
        <div className="header-actions">
          <button onClick={handleNewDoc}>New Document</button>
          {currentDoc && (
            <>
              <button onClick={() => setIsEditing(!isEditing)}>
                {isEditing ? "Preview" : "Edit"}
              </button>
              {isEditing && (
                <button onClick={handleSave} className="save-btn">
                  Save & Commit
                </button>
              )}
            </>
          )}
        </div>
      </header>

      <div className="app-content">
        <DocumentList
          docs={docs}
          currentDoc={currentDoc}
          onDocSelect={handleDocSelect}
        />

        <div className="main-content">
          {currentDoc ? (
            isEditing ? (
              <MarkdownEditor content={content} onChange={setContent} />
            ) : (
              <div
                className="markdown-preview"
                dangerouslySetInnerHTML={{ __html: renderMarkdown(content) }}
              />
            )
          ) : (
            <div className="empty-state">
              <h2>No document selected</h2>
              <p>Select a document from the sidebar or create a new one</p>
            </div>
          )}
        </div>

        <VersionHistory repo={repo} currentDoc={currentDoc} />
      </div>
    </div>
  );
};

export default App;
