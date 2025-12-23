import React from "react";

interface DocumentListProps {
  docs: string[];
  currentDoc: string | null;
  onDocSelect: (doc: string) => void;
}

const DocumentList: React.FC<DocumentListProps> = ({
  docs,
  currentDoc,
  onDocSelect,
}) => {
  return (
    <div className="document-list">
      <h3>Documents</h3>
      {docs.length === 0 ? (
        <div style={{ color: "#999", fontSize: "0.85rem" }}>
          No documents yet
        </div>
      ) : (
        docs.map((doc) => (
          <div
            key={doc}
            className={`doc-item ${currentDoc === doc ? "active" : ""}`}
            onClick={() => onDocSelect(doc)}
          >
            📄 {doc}
          </div>
        ))
      )}
    </div>
  );
};

export default DocumentList;
