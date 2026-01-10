import React, { useState, useCallback } from "react";
import { Repository } from "@browser-git/browser-git";

interface SearchResult {
  type: "document" | "version";
  docName: string;
  commitHash?: string;
  commitMessage?: string;
  commitDate?: string;
  matchContext: string;
  lineNumber?: number;
}

interface SearchPanelProps {
  repo: Repository | null;
  docs: string[];
  onResultSelect: (docName: string, commitHash?: string) => void;
}

const SearchPanel: React.FC<SearchPanelProps> = ({
  repo,
  docs,
  onResultSelect,
}) => {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<SearchResult[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [searchVersions, setSearchVersions] = useState(false);
  const [isExpanded, setIsExpanded] = useState(false);

  const getMatchContext = (
    content: string,
    query: string,
    maxLength: number = 100
  ): { context: string; lineNumber: number } | null => {
    const lowerContent = content.toLowerCase();
    const lowerQuery = query.toLowerCase();
    const index = lowerContent.indexOf(lowerQuery);

    if (index === -1) return null;

    // Find line number
    const linesBeforeMatch = content.substring(0, index).split("\n");
    const lineNumber = linesBeforeMatch.length;

    // Get context around match
    const start = Math.max(0, index - 40);
    const end = Math.min(content.length, index + query.length + 60);

    let context = content.substring(start, end);
    if (start > 0) context = "..." + context;
    if (end < content.length) context = context + "...";

    return { context: context.replace(/\n/g, " "), lineNumber };
  };

  const searchDocuments = useCallback(async () => {
    if (!repo || !query.trim()) {
      setResults([]);
      return;
    }

    setIsSearching(true);
    const searchResults: SearchResult[] = [];

    try {
      // Search current version of all documents
      for (const doc of docs) {
        try {
          const content = await repo.fs.readFile(doc, "utf-8");
          const matchInfo = getMatchContext(content, query);

          if (matchInfo) {
            searchResults.push({
              type: "document",
              docName: doc,
              matchContext: matchInfo.context,
              lineNumber: matchInfo.lineNumber,
            });
          }
        } catch (error) {
          console.error(`Failed to search ${doc}:`, error);
        }
      }

      // Search version history if enabled
      if (searchVersions) {
        try {
          const log = await repo.log({ maxCount: 50 });

          for (const commit of log) {
            // Search commit messages
            if (commit.message.toLowerCase().includes(query.toLowerCase())) {
              searchResults.push({
                type: "version",
                docName: "Commit Message",
                commitHash: commit.hash,
                commitMessage: commit.message.split("\n")[0],
                commitDate: commit.date,
                matchContext: commit.message.split("\n")[0],
              });
            }

            // Search file content at each commit (limited for performance)
            if (searchResults.length < 20) {
              for (const doc of docs) {
                try {
                  // Get file content at this commit
                  const content = await getFileAtCommit(repo, doc, commit.hash);
                  if (content) {
                    const matchInfo = getMatchContext(content, query);
                    if (matchInfo) {
                      // Check if we already have this exact match in current version
                      const isDuplicate = searchResults.some(
                        (r) =>
                          r.docName === doc &&
                          !r.commitHash &&
                          r.matchContext === matchInfo.context
                      );

                      if (!isDuplicate) {
                        searchResults.push({
                          type: "version",
                          docName: doc,
                          commitHash: commit.hash,
                          commitMessage: commit.message.split("\n")[0],
                          commitDate: commit.date,
                          matchContext: matchInfo.context,
                          lineNumber: matchInfo.lineNumber,
                        });
                      }
                    }
                  }
                } catch {
                  // File might not exist at this commit
                }
              }
            }
          }
        } catch (error) {
          console.error("Failed to search version history:", error);
        }
      }

      setResults(searchResults);
    } finally {
      setIsSearching(false);
    }
  }, [repo, docs, query, searchVersions]);

  const getFileAtCommit = async (
    repo: Repository,
    path: string,
    commitHash: string
  ): Promise<string | null> => {
    try {
      // Save current state
      const currentContent = await repo.fs.readFile(path, "utf-8");

      // Checkout file at commit
      await repo.checkout(commitHash, { paths: [path] });
      const historicalContent = await repo.fs.readFile(path, "utf-8");

      // Restore current content
      await repo.fs.writeFile(path, currentContent);

      return historicalContent;
    } catch {
      return null;
    }
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    searchDocuments();
  };

  const handleResultClick = (result: SearchResult) => {
    if (result.type === "document") {
      onResultSelect(result.docName);
    } else {
      onResultSelect(result.docName, result.commitHash);
    }
  };

  const formatDate = (date: string) => {
    return new Date(date).toLocaleDateString();
  };

  const highlightMatch = (text: string, query: string): React.ReactNode => {
    if (!query) return text;

    const lowerText = text.toLowerCase();
    const lowerQuery = query.toLowerCase();
    const index = lowerText.indexOf(lowerQuery);

    if (index === -1) return text;

    return (
      <>
        {text.substring(0, index)}
        <mark className="search-highlight">
          {text.substring(index, index + query.length)}
        </mark>
        {text.substring(index + query.length)}
      </>
    );
  };

  if (!isExpanded) {
    return (
      <button
        className="search-toggle"
        onClick={() => setIsExpanded(true)}
        title="Search documents"
      >
        🔍
      </button>
    );
  }

  return (
    <div className="search-panel">
      <div className="search-header">
        <h3>Search</h3>
        <button
          className="search-close"
          onClick={() => setIsExpanded(false)}
          title="Close search"
        >
          ×
        </button>
      </div>

      <form onSubmit={handleSearch} className="search-form">
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search documents..."
          className="search-input"
          autoFocus
        />
        <button type="submit" className="search-btn" disabled={isSearching}>
          {isSearching ? "..." : "Search"}
        </button>
      </form>

      <label className="search-option">
        <input
          type="checkbox"
          checked={searchVersions}
          onChange={(e) => setSearchVersions(e.target.checked)}
        />
        Include version history
      </label>

      <div className="search-results">
        {results.length > 0 ? (
          <>
            <div className="search-count">
              {results.length} result{results.length !== 1 ? "s" : ""} found
            </div>
            {results.map((result, index) => (
              <div
                key={index}
                className="search-result"
                onClick={() => handleResultClick(result)}
              >
                <div className="result-header">
                  <span className="result-doc">📄 {result.docName}</span>
                  {result.lineNumber && (
                    <span className="result-line">L{result.lineNumber}</span>
                  )}
                </div>
                <div className="result-context">
                  {highlightMatch(result.matchContext, query)}
                </div>
                {result.type === "version" && result.commitHash && (
                  <div className="result-version">
                    <span className="result-hash">
                      {result.commitHash.substring(0, 7)}
                    </span>
                    {result.commitDate && (
                      <span className="result-date">
                        {formatDate(result.commitDate)}
                      </span>
                    )}
                  </div>
                )}
              </div>
            ))}
          </>
        ) : query && !isSearching ? (
          <div className="no-results">No results found for "{query}"</div>
        ) : null}
      </div>
    </div>
  );
};

export default SearchPanel;
