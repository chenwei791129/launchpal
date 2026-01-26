# LaunchPal

> A modern, user-friendly GUI for managing macOS LaunchAgents

## Features

- ✨ Intuitive graphical interface for LaunchAgent management
- 🔍 Real-time service status monitoring (running/stopped)
- 📝 Visual plist configuration editor
- 🚀 Quick start/stop service controls
- 📊 View service logs (stdout/stderr)
- 💾 Automatic backups (keeps last 10 versions)
- 🔄 Backup restoration functionality
- 📋 Click-to-copy for paths and labels

## Screenshots

<!-- Please add your main interface screenshot here -->
<p align="center">
  <img src="docs/screenshots/main-interface.png" alt="Main Interface" width="800">
</p>

<!-- Please add your service details page screenshot here -->
<p align="center">
  <img src="docs/screenshots/service-details.png" alt="Service Details" width="800">
</p>

> 💡 **Tip**: Place screenshots in the `docs/screenshots/` directory. Recommended screenshot size: 1600x1200 or similar aspect ratio.

## System Requirements

- macOS 10.13 or later
- Permission to manage ~/Library/LaunchAgents directory

## Installation

### Download Pre-built Release

Download the latest `.dmg` file from [Releases](https://github.com/YOUR_USERNAME/launchpal/releases).

### Build from Source

#### Prerequisites

- Go 1.23+
- Node.js 20+ and pnpm
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

#### Build Steps

```bash
# Clone the repository
git clone https://github.com/YOUR_USERNAME/launchpal.git
cd launchpal

# Install dependencies
make setup

# Build the application
make build

# The application will be located at: build/bin/launchpal.app
```

## Usage

1. Launch LaunchPal
2. The main page displays all installed LaunchAgents
3. Click on a service name to view detailed information
4. Use Start/Stop buttons to control services
5. Click "Create New Service" to add a new service

## Development Guide

### Development Mode

```bash
# Start development mode with hot reload
make dev
```

### Testing

```bash
# Run tests
make test
```

### Project Structure

```
├── app.go                 # Wails application bindings
├── main.go                # Application entry point
├── internal/
│   ├── launchctl/         # launchctl command wrapper
│   └── backup/            # Backup management
└── frontend/              # Nuxt 4 frontend
    └── app/
        ├── pages/         # Pages
        └── components/    # Vue components
```

## Known Limitations

- Only manages user-level services (~/Library/LaunchAgents)
- Cannot stop services running as root (requires sudo permissions)
- Does not support LaunchDaemons (system-level services)

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

This project uses the following excellent open-source projects:

- [Wails](https://wails.io/) - Go + Web cross-platform application framework
- [Nuxt](https://nuxt.com/) - Vue.js framework
- [TailwindCSS](https://tailwindcss.com/) - CSS framework
