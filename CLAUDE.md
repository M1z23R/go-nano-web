# go-nano-web Guidelines

## Build & Run Commands
```bash
# Build the project
go build

# Run the project
go run *.go

# Format code
go fmt *.go

# Check for issues
go vet *.go

# Create and run tests (when added)
# go test -v ./...
# go test -v -run TestName ./...
```

## Code Style Guidelines
- **Imports**: Group standard library imports first, alphabetically ordered
- **Interfaces**: Use `I` prefix for interfaces (e.g., `IStackable`, `IHandler`)
- **Error Handling**: Return explicit errors up the call stack, use `ApiError` struct
- **Naming**: 
  - CamelCase for exported types/functions
  - camelCase for unexported functions
  - Descriptive function names
- **Comments**: Add comments for exported functions and complex logic
- **Formatting**: Use standard Go formatting (go fmt)
- **Architecture**: Follow clean separation of concerns between server, router, request, and response components
- **Error Messages**: Be specific and descriptive in error messages
- **Middleware Pattern**: Follow established middleware pattern for extending functionality