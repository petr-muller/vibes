# Claude Instructions for vibes

## The Golden Rule  
When unsure about implementation details, ALWAYS ask the developer.  

## Project Context  
The vibes repository is an incubator repository for a collection of prototypes, helpers, utilities and tools that are
useful for OpenShift Upgrades feature development, testing and service maintenance, primarily vibe-coded to experiment
with the technology and to quickly prototype new ideas.

## Technology Considerations

1. Go is the primary language for OpenShift-related tooling; anything with sufficient complexity will be typically
   written in Go.
2. Simpler helpers and utilities can be written in fish shell scripts
3. Python or Rust can be considered for certain tasks but there should be a clear good reason to use them over Go.

## Guidelines:

## What AI Must NEVER Do  

1. **Never modify test files** - Tests encode human intent  
2. **Never commit secrets** - Use environment variables  
3. **Never assume business logic** - Always ask  
4. **Never remove AIDEV- comments** - They're there for a reason  

Remember: We optimize for maintainability over cleverness. When in doubt, choose the boring solution.

### Anchor comments  

Add specially formatted comments throughout the codebase, where appropriate, for yourself as inline knowledge that can be easily `grep`ped for.

- Use `AIDEV-NOTE:`, `AIDEV-TODO:`, or `AIDEV-QUESTION:` (all-caps prefix) for comments aimed at AI and developers.  
- **Important:** Before scanning files, always first try to **grep for existing anchors** `AIDEV-*` in relevant subdirectories.
- **Update relevant anchors** when modifying associated code.
- **Do not remove `AIDEV-NOTE`s** without explicit human instruction.  
- Make sure to add relevant anchor comments, whenever a file or piece of code is:  
  * too complex, or
  * very important, or  
  * confusing, or  
  * could have a bug

### General Practices

- Use descriptive commit messages
- Keep commits focused and atomic
- Update documentation alongside code changes
- Consider security implications, especially for network-facing tools

### Go Conventions

- Follow Go best practices and idioms
- Each tool or utility should have its own directory under `cmd/`
- The `main.go` file in each directory should contain the entry point for that tool
- Business logic should be placed in the `pkg/` directory. Each tool can have its own package under `pkg/` but some
  packages can be shared across tools.
- Use `go mod` for dependency management
- Keep dependencies minimal
- Use `go fmt` to format code before committing
- Use `gotestsum` for running tests to get enhanced output and better test reporting

### Project Structure

- Keep tool-specific documentation in their respective directories
- Update relevant README files when adding new tools or significant changes
- Follow existing conventions in each directory

## Domain Glossary (Claude, learn these!)

- CVO - Cluster Version Operator, responsible for managing updates in OpenShift
- OSUS - OpenShift Update Service, the service that provides updates to OpenShift clusters, also known as the Cincinnati service
- Cincinnati - The update graph protocol used by OpenShift to manage updates
- OSUS Operator - An OLM operator that can deploy and manage an OSUS instance. Also called Cincinnati Operator or Cincinnator

## Tools

- **fauxinnati** - Mock Cincinnati update graph server (see cmd/fauxinnati/README.md and pkg/fauxinnati/README.md for details)
- **scripts** - Various utility scripts for OpenShift CI analysis
