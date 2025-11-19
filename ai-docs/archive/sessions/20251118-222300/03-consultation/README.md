# Consultation Logs

## GPT (openai/gpt-4.1)
Command: cat .../gpt/prompt.md | claudish --model openai/gpt-4.1 > .../gpt/response.md
Status: Success. Detailed recommendation for multi-module monorepo with go.work.

## Gemini (google/gemini-2.0-flash-exp)
Command: cat .../gemini/prompt.md | claudish --model google/gemini-2.0-flash-exp > .../gemini/response.md
Status: Failed (404: No endpoints found).

## Grok (xai/grok-2-1212)
Command: cat .../grok/prompt.md | claudish --model xai/grok-2-1212 > .../grok/response.md
Status: Failed (400: Invalid model ID).

## Codex (openai/gpt-5.1-codex)
Command: cat .../codex/prompt.md | claudish --model openai/gpt-5.1-codex > .../codex/response.md (background ID: 8517b6)
Status: Ran but empty response. Possible model issue or timeout.

Full prompts saved in respective /prompt.md files.
Responses in /response.md files.
