# User Clarifications

## Design Decisions

### 1. Ambiguity Handling: Error (conservative)
When an ambiguous function is detected (e.g., 'Open' could be os.Open or net.Open):
- **Decision**: Throw compile error and force user to qualify
- **Rationale**: Safest approach, avoids surprising behavior
- **Implementation**: List all possible packages in error message with fix-it hints

### 2. Registry Scope: Comprehensive stdlib
Which standard library packages to include:
- **Decision**: Comprehensive stdlib coverage
- **Rationale**: Complete solution from the start
- **Implementation**: Build full registry of all stdlib exported functions

### 3. Local Function Handling: Pre-scan AST (accurate)
How to handle user-defined functions with same names:
- **Decision**: Pre-scan AST for local function definitions
- **Rationale**: Most correct, prevents false positives
- **Implementation**: Scan AST before transformation to build exclusion list

### 4. Transformation Stage: Preprocessor stage (simple)
Where transformation happens:
- **Decision**: Preprocessor stage (text-based)
- **Rationale**: Consistent with current architecture (error propagation, type annotations)
- **Implementation**: Add unqualified import processor to preprocessor pipeline

## Implementation Implications

- Conservative error handling = better debugging experience
- Comprehensive stdlib = one-time effort, complete coverage
- AST pre-scan = slightly more complex but correct
- Preprocessor stage = fits naturally into existing architecture
