# Devrewoh Portfolio

Modern portfolio website built with Go, Templ, Chi router, and vanilla CSS. Features mobile-first responsive design with forest green theming.

## Tech Stack

- **Backend**: Go 1.24.4 + Chi router + Templ
- **Frontend**: Vanilla CSS with CSS Grid
- **Build**: Mage (Go-based automation)
- **Deploy**: Fly.io + Docker

## Quick Start

```bash
# Clone and setup
git clone <repo-url>
cd devrewoh-portfolio
mage setup          # Creates directories and installs tools

# Development
mage dev            # Hot reload server
mage run            # Regular server
mage rebuild        # Clean rebuild
```

Visit `http://localhost:8080`

## Development Commands

```bash
mage                # Show all commands
mage setup         # Initial project setup (creates directories)
mage dev           # Development with hot reload
mage build         # Build binary (outputs to bin/)
mage rebuild       # Clean and rebuild from scratch
mage format        # Format code and templates
mage test          # Run tests
mage clean         # Clean bin/, tmp/, generated files
mage info          # Show project information
```

## Build Artifacts

- **`bin/`** - Compiled binaries
- **`tmp/`** - Air hot reload temporary files  
- **`*_templ.go`** - Generated template files

All build artifacts are cleaned with `mage clean` and ignored by Git.

## Project Structure

```
devrewoh-portfolio/
├── main.go              # Server & routes
├── components.templ     # UI components
├── magefile.go         # Build automation
├── static/css/         # Stylesheets
├── bin/               # Compiled binaries (created by build)
├── tmp/               # Hot reload temp files (created by air)
├── fly.toml           # Fly.io config
└── Dockerfile         # Container config
```

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
mage DockerBuild
mage DockerRun
```

## Key Features

- **Mobile-First**: Responsive CSS Grid layout
- **Component-Based**: Reusable Templ components
- **Type-Safe**: Go templating with compile-time checks
- **Fast**: Minimal dependencies, optimized builds
- **Secure**: Security headers, non-root container
- **Production Ready**: Health checks, graceful shutdown

## Customization

1. **Content**: Edit `components.templ`
2. **Styling**: Modify `static/css/styles.css`
3. **Routes**: Add handlers in `main.go`

## Architecture Decisions

- **Mage over Make**: Type-safe builds, cross-platform
- **Templ over html/template**: Compile-time safety
- **Chi over Gin**: Lightweight, idiomatic
- **Vanilla CSS**: No build tools, direct control
- **Alpine Docker**: Minimal attack surface

## License

MIT