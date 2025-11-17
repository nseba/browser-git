import React from 'react';

interface MarkdownEditorProps {
  content: string;
  onChange: (content: string) => void;
}

const MarkdownEditor: React.FC<MarkdownEditorProps> = ({ content, onChange }) => {
  return (
    <div className="markdown-editor">
      <textarea
        value={content}
        onChange={(e) => onChange(e.target.value)}
        placeholder="Write your documentation in Markdown..."
      />
    </div>
  );
};

export default MarkdownEditor;
