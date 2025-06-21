# Azure OpenAI Integration with CLI Tool

## Overview
Your CLI tool now includes Azure OpenAI integration to evaluate source code files using AI-powered analysis.

## Features Added
1. **Azure OpenAI Chat Completion** - Calls Azure OpenAI API to analyze source code
2. **Environment Variable Configuration** - Secure credential management
3. **Async Processing** - Non-blocking file processing
4. **Error Handling** - Graceful handling of API errors and missing credentials

## Configuration
Set these environment variables before running evaluations:

```bash
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com"
export AZURE_OPENAI_API_KEY="your-api-key-here"
export AZURE_OPENAI_DEPLOYMENT_NAME="gpt-4"
```

## Usage Examples

### List Available Evaluations
```bash
./target/release/cairs ls
```

### Evaluate Source Code
```bash
# Evaluate complexity of all source files in src/ directory
./target/release/cairs evaluate --target-folder src --evaluation-name complexity

# Evaluate with additional options
./target/release/cairs evaluate \
  --target-folder src \
  --evaluation-name no-pointers \
  --skip-files "*.test.rs" \
  --include-files "*.rs,*.py" \
  --junit-file-name results.xml
```

## How It Works
1. The tool loads evaluation templates from `evaluations.json`
2. Finds all source code files in the target directory
3. For each file, sends the content to Azure OpenAI with the selected evaluation prompt
4. Returns structured JSON responses with scores and explanations

## Supported Evaluations
- **complexity**: Evaluates code readability and cyclomatic complexity
- **no-pointers**: Checks for pointer usage in code
- **no-external-packages**: Identifies external package dependencies

## Dependencies Added
- `reqwest`: HTTP client for API calls
- `tokio`: Async runtime
- `serde_json`: JSON parsing for API requests/responses
