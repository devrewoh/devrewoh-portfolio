# Prerequisites
- Go 1.24+ installed
- Git

## Quick Start

```bash
# Clone and setup
git clone <repo-url>
cd devrewoh-portfolio

# Install required tools
go install github.com/magefile/mage@latest
go install github.com/a-h/templ/cmd/templ@latest

# Add Go bin to PATH if not already done
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc

mage setup          # Creates directories and installs tools

# Development
mage dev            # Hot reload server (recommended)
mage devdebug       # Debug mode with verbose output
mage watch          # Alternative watch mode
mage run            # Regular server
```

Visit `http://localhost:8080`

## Development Commands

### Core Development
```bash
mage                # Build (default command)
mage setup         # Initial project setup
mage dev           # Development with hot reload
mage devdebug      # Debug mode development
mage watch         # Alternative watch mode
mage build         # Build binary
mage buildprod     # Production build (Linux/amd64)
mage run           # Run application
mage rebuild       # Clean and rebuild
```

### Quality Assurance
```bash
mage format        # Format Go code and Templ files
mage test          # Run tests
mage testcoverage  # Run tests with HTML coverage report
mage generate      # Generate Templ files manually
```

### Utilities
```bash
mage clean         # Clean all build artifacts
mage install       # Install required tools (templ, air)
mage info          # Show project and tool status
```

### Docker & Production
```bash
mage dockerbuild   # Build Docker image
mage dockerrun     # Run Docker container locally
```

## Required Tools

The following tools are automatically installed with `mage install`:

- **templ** - Template generation (`github.com/a-h/templ/cmd/templ`)
- **air** - Hot reloading (`github.com/air-verse/air`)

Check tool status with `mage info`.

## Testing

- **Test Coverage**: 65.9% with comprehensive test suite
- **What's Tested**: All HTTP handlers, security middleware, static file serving, configuration
- **What's Not**: Server startup/shutdown (integration tests), template error paths

```bash
mage test                    # Run tests
mage testcoverage           # Generate coverage report
open coverage.html          # View detailed coverage
```

## Build Artifacts

- **`bin/`** - Compiled binaries
- **`tmp/`** - Air hot reload temporary files  
- **`*_templ.go`** - Generated template files
- **`coverage.out`** - Test coverage data
- **`coverage.html`** - HTML coverage report

All build artifacts are cleaned with `mage clean` and ignored by Git.

## Project Structure

```
devrewoh-portfolio/
├── main.go              # Server, routes & handlers
├── main_test.go         # Comprehensive test suite
├── components.templ     # UI components & pages
├── components_templ.go  # Generated template code
├── magefile.go         # Build automation
├── .air.toml           # Hot reload configuration
├── static/
│   ├── css/
│   │   └── styles.css  # Dark theme with responsive design
│   ├── images/         # Static images
│   └── js/             # Static JavaScript
├── bin/               # Compiled binaries (created by build)
├── tmp/               # Hot reload temp files (created by air)
├── fly.toml           # Fly.io deployment config
└── Dockerfile         # Container configuration
```

## Configuration

### Environment Variables
- `PORT` - Server port (default: 8080)

### Hot Reload (.air.toml)
Configured for:
- Templ generation on file changes
- Go compilation and restart
- Static file watching
- Graceful process cleanup

## Deployment

### Fly.io (Recommended)

```bash
# Install Fly CLI
curl -L https://fly.io/install.sh | sh

# Login and deploy
fly auth login
fly launch
fly deploy
```

### Docker

```bash
mage dockerbuild
mage dockerrun
```

### Manual Deployment

```bash
mage buildprod              # Build for Linux
scp bin/devrewoh-portfolio server:/path/
scp -r static/ server:/path/
```

## Key Features

### Technical
- **Type-Safe Templates**: Compile-time template checking with Templ
- **Comprehensive Testing**: 65.9% coverage focusing on critical paths
- **Security Headers**: CSP, HSTS, XSS protection
- **Rate Limiting**: Built-in request throttling
- **Graceful Shutdown**: Clean server termination
- **Static File Caching**: Optimized asset delivery

### Design
- **Mobile-First**: Responsive design starting from 320px
- **Dark Theme**: Professional dark gray with forest green accents
- **Accessible**: Proper contrast ratios and semantic HTML
- **Performance**: Minimal CSS/JS, optimized loading

## Development Workflow

1. **Start Development**: `mage dev` (starts hot reload)
2. **Make Changes**: Edit `.templ` files or `main.go`
3. **Auto Rebuild**: Air automatically rebuilds and restarts
4. **Test Changes**: `mage test` or `mage testcoverage`
5. **Format Code**: `mage format` before committing

## Customization

### Content
- **Pages**: Edit components in `components.templ`
- **Data**: Update handler functions in `main.go`
- **Routes**: Add new routes in `setupRoutes()`

### Styling
- **Theme**: Modify CSS variables in `styles.css`
- **Layout**: Update CSS Grid/Flexbox in component styles
- **Responsive**: Adjust breakpoints and mobile-first design

### Functionality
- **New Pages**: Add handler + route + template component
- **Middleware**: Add to `setupMiddleware()` in main.go
- **Static Assets**: Place in `static/` directory

## Architecture Decisions

- **Mage over Make**: Type-safe builds, cross-platform compatibility
- **Templ over html/template**: Compile-time safety, component reuse
- **Chi over Gin**: Lightweight, idiomatic Go, middleware flexibility
- **Vanilla CSS**: No build tools, direct control, fast loading
- **Air over custom**: Battle-tested hot reloading
- **Alpine Docker**: Minimal attack surface, small image size

## Performance

- **Binary Size**: ~15MB (optimized with `-ldflags "-s -w"`)
- **Memory Usage**: ~5MB typical runtime
- **Cold Start**: <100ms
- **Static Assets**: 24-hour cache headers
- **Compression**: Gzip middleware enabled

## Security

- **Headers**: CSP, HSTS, XSS protection, frame denial
- **Rate Limiting**: 100 requests per connection
- **Input Validation**: Request size limits (32KB)
- **Container**: Non-root user, minimal Alpine base
- **Dependencies**: Minimal external dependencies

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes and add tests
4. Run `mage testcoverage` to verify
5. Format with `mage format`
6. Submit a pull request

## License

MIT
