import React, { useState, useEffect } from 'react';
import { Repository } from '@browser-git/browser-git';
import FileTree from './components/FileTree';
import Editor from './components/Editor';
import GitPanel from './components/GitPanel';
import './App.css';

const App: React.FC = () => {
  const [repo, setRepo] = useState<Repository | null>(null);
  const [currentFile, setCurrentFile] = useState<string | null>(null);
  const [fileContent, setFileContent] = useState<string>('');
  const [files, setFiles] = useState<string[]>([]);

  useEffect(() => {
    initializeRepo();
  }, []);

  const initializeRepo = async () => {
    try {
      let repository: Repository;
      try {
        repository = await Repository.open('/mini-ide-repo');
      } catch {
        repository = await Repository.init('/mini-ide-repo', {
          storage: 'indexeddb',
          initialBranch: 'main',
        });
      }
      setRepo(repository);
      await refreshFileList(repository);
    } catch (error) {
      console.error('Failed to initialize repository:', error);
    }
  };

  const refreshFileList = async (repository: Repository) => {
    try {
      const fileList = await repository.fs.readdir('/');
      setFiles(fileList.filter(f => !f.startsWith('.')));
    } catch (error) {
      console.error('Failed to read directory:', error);
    }
  };

  const handleFileSelect = async (filename: string) => {
    if (!repo) return;

    try {
      const content = await repo.fs.readFile(filename, 'utf-8');
      setCurrentFile(filename);
      setFileContent(content);
    } catch (error) {
      console.error('Failed to read file:', error);
    }
  };

  const handleFileChange = (content: string) => {
    setFileContent(content);
  };

  const handleFileSave = async () => {
    if (!repo || !currentFile) return;

    try {
      await repo.fs.writeFile(currentFile, fileContent);
      alert(`File ${currentFile} saved successfully!`);
    } catch (error) {
      console.error('Failed to save file:', error);
      alert(`Failed to save file: ${(error as Error).message}`);
    }
  };

  const handleNewFile = async () => {
    const filename = prompt('Enter file name:');
    if (!filename || !repo) return;

    try {
      await repo.fs.writeFile(filename, '// New file\n');
      await refreshFileList(repo);
      setCurrentFile(filename);
      setFileContent('// New file\n');
    } catch (error) {
      console.error('Failed to create file:', error);
      alert(`Failed to create file: ${(error as Error).message}`);
    }
  };

  return (
    <div className="app">
      <header className="app-header">
        <h1>🚀 BrowserGit Mini IDE</h1>
        <button onClick={handleNewFile}>New File</button>
        <button onClick={handleFileSave} disabled={!currentFile}>
          Save
        </button>
      </header>

      <div className="app-content">
        <FileTree
          files={files}
          currentFile={currentFile}
          onFileSelect={handleFileSelect}
        />

        <Editor
          filename={currentFile}
          content={fileContent}
          onChange={handleFileChange}
        />

        <GitPanel
          repo={repo}
          onRefresh={() => repo && refreshFileList(repo)}
        />
      </div>
    </div>
  );
};

export default App;
