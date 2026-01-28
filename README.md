# LaunchPal

A modern GUI for managing macOS LaunchAgents.

## Features

- Manage LaunchAgents with an intuitive interface
- View user-level and system-level services
- View service status
- Start/Stop user services
- View service logs
- Create user services

## Installation

### Download

Download the latest release from the [Releases](https://github.com/chenwei791129/launchpal/releases) page.

### IMPORTANT: Unsigned Application

LaunchPal is currently **not signed with an Apple Developer certificate**. macOS will block unsigned applications from running by default.

After downloading and extracting the application, you need to remove the quarantine attribute before running it:

```bash
xattr -dr com.apple.quarantine /Applications/launchpal.app
```

**Note:** Replace `/Applications/launchpal.app` with the actual path where you placed the application.

Alternatively, you can try right-clicking the app and selecting "Open" from the context menu, which may prompt you to allow the unsigned app to run.

### Building from Source

If you prefer to build from source:

```bash
# Clone the repository
git clone https://github.com/chenwei791129/launchpal.git
cd launchpal

# Install dependencies
make setup

# Build the application
make build

# The app will be in build/bin/launchpal.app
```

## Known Limitations

- Can only modify user-level services (~/Library/LaunchAgents)
- System services (/Library/LaunchDaemons, /System/Library/LaunchDaemons) are read-only
- Cannot stop services running as root
- Some system services may require Full Disk Access permission to view

## License

MIT License - see [LICENSE](LICENSE) file for details.
