import React, { useState, useEffect } from 'react';
import { Repository } from '@browser-git/browser-git';

interface VersionHistoryProps {
  repo: Repository | null;
  currentDoc: string | null;
}

const VersionHistory: React.FC<VersionHistoryProps> = ({ repo, currentDoc }) => {
  const [commits, setCommits] = useState<any[]>([]);

  useEffect(() => {
    if (repo && currentDoc) {
      loadHistory();
    } else {
      setCommits([]);
    }
  }, [repo, currentDoc]);

  const loadHistory = async () => {
    if (!repo || !currentDoc) return;

    try {
      const log = await repo.log({ maxCount: 20 });
      setCommits(log);
    } catch (error) {
      console.error('Failed to load history:', error);
    }
  };

  const handleRestore = async (commit: any) => {
    if (!repo || !currentDoc) return;

    const confirmed = window.confirm(
      `Restore ${currentDoc} to version ${commit.hash.substring(0, 7)}?`
    );

    if (!confirmed) return;

    try {
      await repo.checkout(commit.hash, { paths: [currentDoc] });
      alert('Document restored! Reload the page to see changes.');
      window.location.reload();
    } catch (error) {
      alert(`Failed to restore: ${(error as Error).message}`);
    }
  };

  const formatDate = (date: string) => {
    return new Date(date).toLocaleString();
  };

  return (
    <div className="version-history">
      <h3>Version History</h3>
      {currentDoc ? (
        commits.length > 0 ? (
          commits.map(commit => (
            <div
              key={commit.hash}
              className="commit-item"
              onClick={() => handleRestore(commit)}
              title="Click to restore this version"
            >
              <div className="commit-hash">{commit.hash.substring(0, 7)}</div>
              <div className="commit-message">{commit.message.split('\n')[0]}</div>
              <div className="commit-meta">
                {commit.author.name} • {formatDate(commit.date)}
              </div>
            </div>
          ))
        ) : (
          <div style={{ color: '#999', fontSize: '0.85rem' }}>No commits yet</div>
        )
      ) : (
        <div style={{ color: '#999', fontSize: '0.85rem' }}>Select a document</div>
      )}
    </div>
  );
};

export default VersionHistory;
