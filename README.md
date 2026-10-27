# App Intelligence MCP

A Go MCP server for mobile app intelligence, designed to bring App Store Optimization (ASO) and product analytics into AI agent workflows.

> **Early development:** The project currently provides a mock keyword rankings tool. Astro and Mixpanel integrations are planned and are not yet implemented. The current tool does not return real ASO data.

## Why this project?

Mobile app teams use ASO tools to understand discovery and product analytics to understand user behavior. App Intelligence MCP aims to expose these signals through focused MCP tools, with deterministic calculations, explicit missing-data handling, and clear source information.

The project is also a hands-on exploration of Go, MCP, and reliable backend integration design.

## Current capabilities

- MCP server built with the official Go SDK
- Local stdio transport
- A mock `get_keyword_rankings` tool with structured output
- Rejection of empty or whitespace-only keywords
- Rank change calculations that distinguish missing data from unchanged rankings
- Table-driven unit tests and in-memory MCP integration tests

The in-memory tests cover tool discovery, invocation, output values, and blank-input error behavior. End-to-end verification with an external stdio client is still pending.

## Requirements

- Go 1.26.3 or later, matching the current development environment

Provider credentials are not required for the current mock implementation.

## Getting started

From the repository directory, install dependencies and run the tests:

```sh
go mod download
go test ./... -v
```

Build the server:

```sh
go build -o app-intelligence-mcp .
```

The binary is intended to be launched by an MCP client using stdio. Running it directly starts a process that waits for MCP messages; it does not provide an interactive command prompt or HTTP endpoint.

MCP protocol messages use stdout. Application logs must use stderr.

## Available tool

### `get_keyword_rankings`

Returns fixed mock rankings for the supplied keyword. It does not perform a keyword lookup or contact an external provider.

Input:

```json
{
	"keyword": "photo editor"
}
```

Structured output:

```json
{
	"keyword": "photo editor",
	"current_rank": 11,
	"previous_rank": 18,
	"rank_change": 7
}
```

The keyword is taken from the input. Current rank and previous rank are fixed at `11` and `18`; rank change is calculated as `previous_rank - current_rank`. A positive change indicates an improvement in ranking.

An empty or whitespace-only keyword produces a tool execution error.

## Architecture

The current implementation consists of a local MCP server, a mock tool handler, and a ranking calculation model.

The intended architecture is:

```text
MCP client / AI agent
         |
         | MCP
         v
App Intelligence MCP (Go)
         |
         +-- Astro connector --> Supported Astro MCP interface
         |
         +-- Mixpanel connector --> Supported reporting APIs
         |
         +-- Analytics calculations and normalized results
```

Provider connectors and unified analytics are planned. The MCP server exposes tools; it is not an LLM or an agent runtime. Any future Telegram integration will use a separate agent runtime to call these tools.

## Roadmap

- [x] Implement and test rank change calculations
- [x] Register a mock MCP tool with structured output
- [x] Test MCP discovery and invocation in memory
- [x] Add basic keyword validation
- [ ] Verify the stdio transport with an external MCP client
- [ ] Validate Astro tool schemas and integrate its supported local MCP interface
- [ ] Validate Mixpanel access and integrate supported saved-report queries
- [ ] Add source metadata, data freshness, and explicit partial-data responses
- [ ] Add provider timeouts, cancellation, rate-limit handling, and bounded retries
- [ ] Document client configuration and provider setup
- [ ] Add CI and prepare the first release
- [ ] Explore Telegram integration through a separate agent runtime

Provider capabilities, permissions, and subscription requirements will determine the final scope. Arbitrary date ranges, competitor analysis, and custom funnels are not currently supported features of this project.

## Design principles

- Keep the architecture small and understandable.
- Use supported, documented provider interfaces.
- Calculate metrics deterministically.
- Never present missing data as zero.
- Make data sources and comparison periods explicit.
- Do not present correlation as evidence of causation.
- Treat tool inputs and provider text as untrusted data.
- Keep credentials out of source control, tool results, and logs.

## Development

Format code and run tests before committing:

```sh
go fmt ./...
go test ./...
```

Run the MCP integration test on its own:

```sh
go test -run '^TestKeywordRankingsTool$' -v
```

No external provider account or Claude subscription is required to run the current tests.

## Project status

This project is under active development. APIs and tool contracts may change. It is not yet production-ready and has not been validated against live Astro or Mixpanel data.

## License

A license has not yet been selected. A license file will be added before the open-source release.
