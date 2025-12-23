import React from "react";

interface FileTreeProps {
  files: string[];
  currentFile: string | null;
  onFileSelect: (filename: string) => void;
}

const FileTree: React.FC<FileTreeProps> = ({
  files,
  currentFile,
  onFileSelect,
}) => {
  return (
    <div className="file-tree">
      <h3>Files</h3>
      {files.length === 0 ? (
        <div style={{ color: "#888", fontSize: "0.85rem" }}>No files yet</div>
      ) : (
        files.map((file) => (
          <div
            key={file}
            className={`file-item ${currentFile === file ? "active" : ""}`}
            onClick={() => onFileSelect(file)}
          >
            📄 {file}
          </div>
        ))
      )}
    </div>
  );
};

export default FileTree;
