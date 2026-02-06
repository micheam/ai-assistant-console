- After modifying the protobuf definition, make sure to run make protogen
- To run integration tests that require API keys: `go test -tags=integration ./...`
- GitHub Actions secrets required: `AICO_OPENAI_API_KEY` for integration tests on main branch

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


# AI-DLC and Spec-Driven Development

Kiro-style Spec Driven Development implementation on AI-DLC (AI Development Life Cycle)

## Project Context

### Paths
- Steering: `.kiro/steering/`
- Specs: `.kiro/specs/`

### Steering vs Specification

**Steering** (`.kiro/steering/`) - Guide AI with project-wide rules and context
**Specs** (`.kiro/specs/`) - Formalize development process for individual features

### Active Specifications
- Check `.kiro/specs/` for active specifications
- Use `/kiro:spec-status [feature-name]` to check progress

## Development Guidelines
- Think in English, generate responses in Japanese. All Markdown content written to project files (e.g., requirements.md, design.md, tasks.md, research.md, validation reports) MUST be written in the target language configured for this specification (see spec.json.language).

## Minimal Workflow
- Phase 0 (optional): `/kiro:steering`, `/kiro:steering-custom`
- Phase 1 (Specification):
  - `/kiro:spec-init "description"`
  - `/kiro:spec-requirements {feature}`
  - `/kiro:validate-gap {feature}` (optional: for existing codebase)
  - `/kiro:spec-design {feature} [-y]`
  - `/kiro:validate-design {feature}` (optional: design review)
  - `/kiro:spec-tasks {feature} [-y]`
- Phase 2 (Implementation): `/kiro:spec-impl {feature} [tasks]`
  - `/kiro:validate-impl {feature}` (optional: after implementation)
- Progress check: `/kiro:spec-status {feature}` (use anytime)

## Development Rules
- 3-phase approval workflow: Requirements → Design → Tasks → Implementation
- Human review required each phase; use `-y` only for intentional fast-track
- Keep steering current and verify alignment with `/kiro:spec-status`
- Follow the user's instructions precisely, and within that scope act autonomously: gather the necessary context and complete the requested work end-to-end in this run, asking questions only when essential information is missing or the instructions are critically ambiguous.

## Steering Configuration
- Load entire `.kiro/steering/` as project memory
- Default files: `product.md`, `tech.md`, `structure.md`
- Custom files are supported (managed via `/kiro:steering-custom`)
