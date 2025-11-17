import React from 'react';

interface EditorProps {
  filename: string | null;
  content: string;
  onChange: (content: string) => void;
}

const Editor: React.FC<EditorProps> = ({ filename, content, onChange }) => {
  return (
    <div className="editor">
      <div className="editor-header">
        {filename || 'No file selected'}
      </div>
      <textarea
        value={content}
        onChange={(e) => onChange(e.target.value)}
        placeholder="Select a file or create a new one to start editing..."
        disabled={!filename}
      />
    </div>
  );
};

export default Editor;
