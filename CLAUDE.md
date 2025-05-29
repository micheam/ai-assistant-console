- After modifying the protobuf definition, make sure to run make protogen
- To run integration tests that require API keys: `go test -tags=integration ./...`
- GitHub Actions secrets required: `OPENAI_API_KEY` for integration tests on main branch

## Commit Message Guidelines

Follow Conventional Commits specification: https://www.conventionalcommits.org/

Format: `<type>(<scope>): <description>`

**Types:**
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `style` - Code style changes (formatting, etc.)
- `refactor` - Code refactoring
- `test` - Adding or updating tests
- `chore` - Maintenance tasks, build changes, etc.
- `perf` - Performance improvements
- `ci` - CI/CD changes

**Scopes for this repository:**
- `cli` - CLI interface and command-line functionality
- `api` - API integrations (OpenAI, Anthropic)
- `tui` - Terminal user interface
- `repl` - REPL functionality
- `session` - Chat session management
- `config` - Configuration management
- `plugin` - Vim plugin functionality
- `proto` - Protocol buffer definitions
- `markdown` - Markdown processing
- `logging` - Logging utilities
- `theme` - Theming and UI styling
- `spinner` - Loading indicators
- `build` - Build system and Makefile
- `deps` - Dependencies management
- `docker` - Docker-related changes
- `core` - Core application logic
- `util` - Utility functions and helpers
- `security` - Security-related changes

**Examples:**
- `feat(cli): add new chat command`
- `fix(api): handle OpenAI rate limiting`
- `feat(plugin): add Vim integration`
- `chore(ci): update release workflow`
- `docs(readme): update installation instructions`
- `refactor(session): improve chat session storage`
