import { useState, useEffect } from 'react';
import { CodeComparison } from './CodeComparison';
import logoImage from '../../assets/dingo-logo.png';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

interface AppProps {
  examples: Array<{
    id: number;
    title: string;
    language: string;
    description: string;
    before: string;
    after: string;
  }>;
}

// Mock data removed - now using real Dingo examples passed as props

export default function App({ examples }: AppProps) {
  const [selectedId, setSelectedId] = useState(1);
  const [readmeContent, setReadmeContent] = useState('');
  
  const selectedExample = examples.find(ex => ex.id === selectedId) || examples[0];

  // Fetch README.md content
  useEffect(() => {
    fetch('https://raw.githubusercontent.com/MadAppGang/dingo/refs/heads/main/README.md')
      .then(response => response.text())
      .then(data => setReadmeContent(data))
      .catch(error => console.error('Error fetching README:', error));
  }, []);

  return (
    <div className="flex h-screen bg-white">
      {/* Sidebar */}
      <div className="w-80 bg-white border-r border-gray-200 flex flex-col">
        <div className="px-6 pt-6 h-20 flex items-center gap-3 relative overflow-visible">
          <img src={logoImage} alt="Dingo Logo" className="h-24 w-24 object-contain" />
          <h1 className="text-gray-900">Dingo</h1>
        </div>
        
        <nav className="flex-1 px-6 pt-6 pb-6 space-y-2 overflow-auto">
          {examples.map((example) => (
            <button
              key={example.id}
              onClick={() => setSelectedId(example.id)}
              className={`w-full text-left px-4 py-3 rounded-lg transition-colors text-sm ${
                selectedId === example.id
                  ? 'text-blue-600 bg-blue-50'
                  : 'text-gray-700 hover:text-gray-900 hover:bg-gray-50'
              }`}
            >
              {example.title}
            </button>
          ))}
        </nav>

        {/* Description section */}
        <div className="p-6 border-t border-gray-200 bg-gray-50">
          <h3 className="text-gray-900 mb-2 text-sm">About This Tool</h3>
          <p className="text-gray-600 text-xs leading-relaxed">
            Compare code examples side by side to understand best practices and modern patterns. 
            Each example shows the transformation from older or less optimal code to improved implementations.
          </p>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex flex-col overflow-hidden bg-gray-50">
        
        <div className="flex-1 overflow-auto pt-8">
          <CodeComparison
            before={selectedExample.before}
            after={selectedExample.after}
            language={selectedExample.language}
          />
          
          {/* Description of the change */}
          <div className="p-8 bg-white">
            <div className="max-w-4xl mx-auto markdown-content">
              {readmeContent ? (
                <ReactMarkdown 
                  remarkPlugins={[remarkGfm]}
                  components={{
                    h1: ({node, ...props}) => <h1 className="mt-4 mb-2 text-sm" {...props} />,
                    h2: ({node, ...props}) => <h2 className="mt-4 mb-2 text-sm" {...props} />,
                    h3: ({node, ...props}) => <h3 className="mt-3 mb-1 text-xs" {...props} />,
                    h4: ({node, ...props}) => <h4 className="mt-3 mb-1 text-xs" {...props} />,
                    p: ({node, ...props}) => <p className="mb-2 text-gray-700 leading-relaxed text-xs" {...props} />,
                    ul: ({node, ...props}) => <ul className="mb-2 ml-4 list-disc text-gray-700 text-xs" {...props} />,
                    ol: ({node, ...props}) => <ol className="mb-2 ml-4 list-decimal text-gray-700 text-xs" {...props} />,
                    li: ({node, ...props}) => <li className="mb-1 text-xs" {...props} />,
                    code: ({node, inline, ...props}) => 
                      inline 
                        ? <code className="bg-gray-100 px-1 py-0.5 rounded text-xs text-gray-800" {...props} />
                        : <code className="block bg-gray-100 p-2 rounded text-xs overflow-x-auto" {...props} />,
                    pre: ({node, ...props}) => <pre className="mb-2 bg-gray-100 p-2 rounded overflow-x-auto text-xs" {...props} />,
                    a: ({node, ...props}) => <a className="text-blue-600 hover:underline text-xs" {...props} />,
                    blockquote: ({node, ...props}) => <blockquote className="border-l-4 border-gray-300 pl-3 italic text-gray-600 mb-2 text-xs" {...props} />,
                    img: ({node, ...props}) => <img className="max-w-full h-auto rounded my-2" {...props} />,
                    hr: ({node, ...props}) => <hr className="my-4 border-gray-200" {...props} />,
                    table: ({node, ...props}) => <table className="mb-2 border-collapse w-full text-xs" {...props} />,
                    thead: ({node, ...props}) => <thead className="bg-gray-50" {...props} />,
                    tbody: ({node, ...props}) => <tbody {...props} />,
                    tr: ({node, ...props}) => <tr className="border-b border-gray-200" {...props} />,
                    th: ({node, ...props}) => <th className="px-2 py-1 text-left text-xs" {...props} />,
                    td: ({node, ...props}) => <td className="px-2 py-1 text-gray-700 text-xs" {...props} />,
                  }}
                >
                  {readmeContent}
                </ReactMarkdown>
              ) : (
                <p className="text-gray-500">Loading README...</p>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}