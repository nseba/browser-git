import React, { useState, useEffect } from 'react';
import { Repository } from '@browser-git/browser-git';

interface GitPanelProps {
  repo: Repository | null;
  onRefresh: () => void;
}

const GitPanel: React.FC<GitPanelProps> = ({ repo, onRefresh }) => {
  const [commitMessage, setCommitMessage] = useState('');
  const [status, setStatus] = useState<any>(null);
  const [commits, setCommits] = useState<any[]>([]);
  const [currentBranch, setCurrentBranch] = useState<string>('');

  useEffect(() => {
    if (repo) {
      refreshStatus();
      refreshCommits();
      refreshBranch();
    }
  }, [repo]);

  const refreshStatus = async () => {
    if (!repo) return;
    try {
      const st = await repo.status();
      setStatus(st);
    } catch (error) {
      console.error('Failed to get status:', error);
    }
  };

  const refreshCommits = async () => {
    if (!repo) return;
    try {
      const log = await repo.log({ maxCount: 5 });
      setCommits(log);
    } catch (error) {
      console.error('Failed to get log:', error);
    }
  };

  const refreshBranch = async () => {
    if (!repo) return;
    try {
      const branch = await repo.currentBranch();
      setCurrentBranch(branch || 'HEAD');
    } catch (error) {
      console.error('Failed to get branch:', error);
    }
  };

  const handleAdd = async () => {
    if (!repo) return;
    try {
      await repo.add(['.']);
      await refreshStatus();
      onRefresh();
      alert('Files staged successfully!');
    } catch (error) {
      alert(`Failed to stage files: ${(error as Error).message}`);
    }
  };

  const handleCommit = async () => {
    if (!repo || !commitMessage) return;
    try {
      await repo.commit(commitMessage, {
        author: {
          name: 'Mini IDE User',
          email: 'user@mini-ide.local',
        },
      });
      setCommitMessage('');
      await refreshStatus();
      await refreshCommits();
      onRefresh();
      alert('Commit created successfully!');
    } catch (error) {
      alert(`Failed to commit: ${(error as Error).message}`);
    }
  };

  return (
    <div className="git-panel">
      <h3>Git</h3>

      <div style={{ marginBottom: '15px', color: '#888', fontSize: '0.85rem' }}>
        Branch: <span style={{ color: '#4ec9b0' }}>{currentBranch}</span>
      </div>

      <button onClick={handleAdd}>Stage All Changes</button>

      <input
        type="text"
        placeholder="Commit message"
        value={commitMessage}
        onChange={(e) => setCommitMessage(e.target.value)}
      />

      <button onClick={handleCommit} disabled={!commitMessage}>
        Commit
      </button>

      {status && (
        <div className="status-section">
          <h4 style={{ fontSize: '0.85rem', marginBottom: '10px', color: '#888' }}>
            Status
          </h4>
          {status.modified.length > 0 && status.modified.map((file: string) => (
            <div key={file} className="file-status modified">
              M {file}
            </div>
          ))}
          {status.untracked.length > 0 && status.untracked.map((file: string) => (
            <div key={file} className="file-status untracked">
              ? {file}
            </div>
          ))}
          {status.modified.length === 0 && status.untracked.length === 0 && (
            <div style={{ color: '#888', fontSize: '0.85rem' }}>Clean working tree</div>
          )}
        </div>
      )}

      {commits.length > 0 && (
        <div className="commit-list">
          <h4 style={{ fontSize: '0.85rem', marginBottom: '10px', color: '#888' }}>
            Recent Commits
          </h4>
          {commits.map((commit) => (
            <div key={commit.hash} className="commit-item">
              <div className="commit-hash">{commit.hash.substring(0, 7)}</div>
              <div className="commit-message">{commit.message.split('\n')[0]}</div>
              <div className="commit-author">{commit.author.name}</div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default GitPanel;
