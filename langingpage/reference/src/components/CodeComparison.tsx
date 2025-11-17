import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { X, Check } from 'lucide-react';

interface CodeComparisonProps {
  before: string;
  after: string;
  language: string;
}

function CodeBlock({ code, language }: { code: string; language: string }) {
  return (
    <div className="bg-[#1e1e1e] rounded-xl overflow-hidden shadow-2xl">
      {/* macOS-style window controls */}
      <div className="bg-[#323232] px-4 py-3 flex items-center gap-2">
        <div className="w-3 h-3 rounded-full bg-[#ff5f56]"></div>
        <div className="w-3 h-3 rounded-full bg-[#ffbd2e]"></div>
        <div className="w-3 h-3 rounded-full bg-[#27c93f]"></div>
      </div>
      
      {/* Code content */}
      <div className="overflow-auto">
        <SyntaxHighlighter
          language={language}
          style={vscDarkPlus}
          customStyle={{
            margin: 0,
            padding: '24px',
            background: '#1e1e1e',
            fontSize: '14px',
            lineHeight: '1.6',
          }}
          showLineNumbers={false}
        >
          {code}
        </SyntaxHighlighter>
      </div>
    </div>
  );
}

export function CodeComparison({ before, after, language }: CodeComparisonProps) {
  return (
    <div className="grid grid-cols-2 gap-8 p-8">
      {/* Before */}
      <div className="flex flex-col gap-4">
        <div className="flex items-center gap-2">
          <div className="w-6 h-6 rounded-full bg-red-100 flex items-center justify-center">
            <X className="w-4 h-4 text-red-600" />
          </div>
          <h3 className="text-gray-700">Before</h3>
        </div>
        <CodeBlock code={before} language={language} />
      </div>

      {/* After */}
      <div className="flex flex-col gap-4">
        <div className="flex items-center gap-2">
          <div className="w-6 h-6 rounded-full bg-green-100 flex items-center justify-center">
            <Check className="w-4 h-4 text-green-600" />
          </div>
          <h3 className="text-gray-700">After</h3>
        </div>
        <CodeBlock code={after} language={language} />
      </div>
    </div>
  );
}