# Copilot Instructions for AI Coding Agents

## Project Overview
This repository is a Go-based Discord bot, organized for modularity and maintainability. The main entry point is in `cmd/bot/main.go`, with core logic split into `internal/commands` and `internal/discord`.

## Architecture & Major Components
- **cmd/bot/main.go**: Application entry; initializes the bot and orchestrates startup.
- **internal/discord/**: Handles Discord API integration, event routing, and bot lifecycle.
- **internal/commands/**: Implements bot commands (e.g., play, status) and helpers. Each command is a separate file for clarity.
- **main**: May contain static assets or configuration (verify usage before editing).

## Developer Workflows
- **Build**: Use `make` or `docker-compose build` for containerized builds. The `Dockerfile` and `Makefile` define reproducible environments.
- **Run**: Start the bot with `docker-compose up` or by running the built binary from `cmd/bot`.
- **Dependencies**: Managed via Go modules (`go.mod`, `go.sum`). Use `go get` for new packages.

## Project-Specific Patterns
- **Command Pattern**: Each bot command is a function in its own file under `internal/commands`. Shared logic goes in `helpers.go`.
- **Router**: Discord events are routed via `internal/discord/router.go`.
- **Error Handling**: Prefer returning errors up the call stack; log errors at the top level (see `discord.go`).
- **Configuration**: Environment variables are typically used for secrets and tokens. Check Dockerfile and Compose for examples.

## Integration Points
- **Discord API**: All bot interactions are via the Discord API, abstracted in `internal/discord`.
- **External Services**: If commands interact with external APIs, document the integration in the command file header.

## Conventions
- **File Naming**: Use lowercase, descriptive names for commands and helpers.
- **Logging**: Use Go's standard logging or a project-wide logger (see `discord.go`).
- **Testing**: If tests exist, they should be in a parallel `*_test.go` file in the same directory as the code.

## Example: Adding a Command
1. Create a new file in `internal/commands` (e.g., `foo.go`).
2. Implement the command as a function, exporting it if needed.
3. Register the command in the router if it should be triggered by Discord events.

## Key Files & Directories
- `cmd/bot/main.go`: Startup logic
- `internal/discord/`: Discord integration & routing
- `internal/commands/`: Bot commands
- `Dockerfile`, `docker-compose.yml`, `Makefile`: Build/run workflows

---
For questions or unclear patterns, review the referenced files or ask for clarification. Update this document as new conventions emerge.
