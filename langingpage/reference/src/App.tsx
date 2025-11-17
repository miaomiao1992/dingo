import { useState, useEffect } from 'react';
import { CodeComparison } from './components/CodeComparison';
import logoImage from 'figma:asset/4efa237468ffc0134271e92beb7216d2fb829847.png';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

const codeExamples = [
  {
    id: 1,
    title: 'React Class to Functional Component',
    language: 'typescript',
    description: `## Why Make This Change?

**Modern React** encourages the use of functional components with Hooks instead of class components. This transformation brings several benefits:

- **Simpler code**: Less boilerplate, no \`this\` keyword confusion
- **Better performance**: Hooks like \`useState\` are optimized for modern React
- **Easier testing**: Pure functions are simpler to test than class methods
- **Future-proof**: React team focuses development on functional components

### Key Changes

1. Replaced \`class\` with \`function\`
2. Used \`useState\` Hook instead of \`this.state\`
3. Removed constructor and \`this\` references
4. Simplified event handlers`,
    before: `import React, { Component } from 'react';

class Counter extends Component {
  constructor(props) {
    super(props);
    this.state = { count: 0 };
  }

  increment = () => {
    this.setState({ count: this.state.count + 1 });
  };

  render() {
    return (
      <div>
        <h1>Count: {this.state.count}</h1>
        <button onClick={this.increment}>
          Increment
        </button>
      </div>
    );
  }
}

export default Counter;`,
    after: `import { useState } from 'react';

function Counter() {
  const [count, setCount] = useState(0);

  const increment = () => {
    setCount(count + 1);
  };

  return (
    <div>
      <h1>Count: {count}</h1>
      <button onClick={increment}>
        Increment
      </button>
    </div>
  );
}

export default Counter;`
  },
  {
    id: 2,
    title: 'JavaScript to TypeScript',
    language: 'typescript',
    description: `## Why Add TypeScript?

Adding TypeScript to your JavaScript code provides **type safety** and better developer experience:

- **Catch errors early**: Type errors are caught during development, not runtime
- **Better IDE support**: Autocomplete, refactoring tools work better with types
- **Self-documenting**: Types serve as inline documentation
- **Safer refactoring**: Change code with confidence knowing types will catch issues

### Key Changes

1. Added \`CartItem\` interface to define the shape of cart items
2. Added type annotations to function parameters
3. Added return type annotations to functions`,
    before: `function calculateTotal(items) {
  let total = 0;
  
  for (const item of items) {
    total += item.price * item.quantity;
  }
  
  return total;
}

function applyDiscount(total, discount) {
  return total - (total * discount);
}

export { calculateTotal, applyDiscount };`,
    after: `interface CartItem {
  price: number;
  quantity: number;
  name: string;
}

function calculateTotal(items: CartItem[]): number {
  let total = 0;
  
  for (const item of items) {
    total += item.price * item.quantity;
  }
  
  return total;
}

function applyDiscount(total: number, discount: number): number {
  return total - (total * discount);
}

export { calculateTotal, applyDiscount };`
  },
  {
    id: 3,
    title: 'Callback Hell to Async/Await',
    language: 'javascript',
    description: `## Escaping Callback Hell

**Async/await** syntax makes asynchronous code look and behave like synchronous code, improving readability dramatically:

- **Linear flow**: Code reads top to bottom, not nested callbacks
- **Error handling**: Use familiar try/catch instead of callback error parameters
- **Debugging**: Stack traces are clearer and easier to follow
- **Maintainability**: Easier to add new async operations

### Key Changes

1. Replaced nested callbacks with sequential \`await\` calls
2. Code now reads like synchronous operations
3. Much easier to understand the flow of data`,
    before: `getUserData(userId, (user) => {
  getOrders(user.id, (orders) => {
    processOrders(orders, (result) => {
      console.log(result);
    });
  });
});`,
    after: `const user = await getUserData(userId);
const orders = await getOrders(user.id);
const result = await processOrders(orders);
console.log(result);`
  },
  {
    id: 4,
    title: 'CSS to Tailwind',
    language: 'css',
    description: `## Utility-First CSS

**Tailwind CSS** takes a utility-first approach, replacing custom CSS with composable utility classes:

- **Faster development**: No need to name classes or switch between files
- **Smaller bundle**: Only the utilities you use are included
- **Consistent design**: Built-in design system prevents arbitrary values
- **Easier maintenance**: Changes are localized to the component

### Key Changes

1. Replaced custom CSS classes with Tailwind utilities
2. Styles are now co-located with markup
3. No separate CSS file needed for simple components`,
    before: `.card {
  background-color: white;
  border-radius: 8px;
  padding: 24px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  max-width: 400px;
}

.card-title {
  font-size: 24px;
  font-weight: bold;
  margin-bottom: 16px;
}`,
    after: `/* Using Tailwind utility classes */
<div className="bg-white rounded-lg p-6 shadow-lg max-w-md">
  <h2 className="mb-4">Welcome</h2>
  <p>This is a card component</p>
</div>`
  }
];

export default function App() {
  const [selectedId, setSelectedId] = useState(1);
  const [readmeContent, setReadmeContent] = useState('');
  
  const selectedExample = codeExamples.find(ex => ex.id === selectedId) || codeExamples[0];

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
          {codeExamples.map((example) => (
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