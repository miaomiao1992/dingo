# Qwen3 Coder Analysis - MODEL TIMEOUT

## Status: FAILED

**Model**: qwen/qwen3-coder-30b-a3b-instruct
**Invocation**: Via claudish CLI
**Result**: Model timed out after 8+ minutes with no response

## Error Details

The model appeared to accept the request but never generated a response beyond the initial claudish header. The API request appears to have stalled during pre-flight checks.

**File output remained at 193 bytes** (only claudish header + warning).

**Warning message**:
```
⚠️  [BashTool] Pre-flight check is taking longer than expected.
Run with ANTHROPIC_LOG=debug to check for failed or slow API requests.
```

## Possible Causes

1. **Model availability**: qwen/qwen3-coder-30b-a3b-instruct may be temporarily unavailable
2. **Token limit**: The prompt may have exceeded the model's context window
3. **API timeout**: The OpenRouter/provider backend may have timed out
4. **Model capacity**: The model may be overloaded or rate-limited

## Recommendation

**Skip this model** and proceed with other models in the parallel investigation:
- openai/gpt-5.1-codex
- google/gemini-2.5-flash
- x-ai/grok-code-fast-1
- minimax/minimax-m2
- openrouter/sherlock-think-alpha
- z-ai/glm-4.6

## Investigation Prompt

The prompt sent to the model is documented in the user request. The investigation focused on:
- Source map translation logic in MapToOriginal
- Diagnostic position translation
- Why gopls error diagnostics are mapping to the wrong Dingo position

**This model could not provide analysis.**
