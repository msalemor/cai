# CAI: Code Analysis with Intelligence

CAI (Code Analysis with Intelligence) uses AI to apply custom code checkers in GitHub Actions or Azure DevOps pipelines.

## Summary
CAI is a Go-based command-line tool that uses AI to perform custom code quality checks on source files. It can be integrated into CI/CD pipelines to automate code reviews based on predefined criteria. The tool evaluates source code files against customizable rules defined in evaluation prompts and provides scores and explanations for each file.

## Problem Statement
Traditional static code analyzers often:
1. Have limited customization capabilities
2. Require complex rule definitions
3. Cannot assess subjective aspects like code clarity or complexity
4. Provide minimal context in their recommendations

CAI addresses these limitations by leveraging AI models (OpenAI) to perform nuanced code analysis based on natural language prompts, making it easier to enforce custom coding standards and best practices.

## Solution
CAI provides:
1. A flexible framework for defining custom code quality checks using natural language prompts
2. Integration with AI models (specifically OpenAI's GPT models) to evaluate code quality
3. A CLI interface to run these evaluations on source files in a directory
4. Support for parallel processing to improve performance with large codebases
5. Clear, actionable output with color-coded results and detailed explanations
6. Pass/fail exit codes for seamless CI/CD pipeline integration

## Technical Details

### Architecture
CAI is built in Go and uses the Cobra CLI framework for command processing. It consists of:

1. **Command Layer**: CLI commands defined in the `cmd` package
   - `rootcmd.go`: Defines the base command structure
   - `evaluatecmd.go`: Handles the "evaluate" command for running code checks
   - `lscmd.go`: Handles the "ls" command for listing available evaluations

2. **Core Package**: Implementation logic in the `pkg` package
   - `evaluations.go`: Loads and manages evaluation prompts from JSON
   - `processfiles.go`: Processes files sequentially
   - `pprocessfiles.go`: Processes files in parallel using goroutines
   - `openai.go`: Interacts with OpenAI API for AI evaluations
   - `listfiles.go`: Lists source files in a directory
   - `settings.go`: Manages application settings
   - `type.go`: Defines common data structures

### Workflow
1. Load settings from environment variables
2. Load evaluation prompts from evaluation.json
3. When "evaluate" command is run:
   - Find all relevant source files in the specified directory
   - For each file:
     - Read the file content
     - Send the content to the OpenAI API with the specified evaluation prompt
     - Parse the response (score and explanation)
   - Display results with color coding (green for passing, red for failing)
   - Exit with error code if any file fails the evaluation

### Evaluation Types
The tool currently supports three evaluation types:
1. **complexity**: Evaluates code for clarity and complexity (score 1-10)
2. **no-pointers**: Checks for pointer usage in the code
3. **no-external-packages**: Verifies if code uses external packages

### Configuration
- Environment variables (`CAI_ENDPOINT`, `CAI_KEY`, `CAI_MODEL`, `CAI_TYPE`) for OpenAI API configuration
- JSON-based evaluation prompts that define custom code checks
- Command-line flags for specifying source directory, evaluation type, and parallel processing

### Integration
The tool is designed to be integrated into CI/CD pipelines:
- It returns exit code 1 when evaluations fail (any score < 5.0)
- It returns exit code 0 when all evaluations pass
- Output is color-coded and formatted for readability
