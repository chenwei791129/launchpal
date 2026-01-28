# LaunchPal

A modern GUI for managing macOS LaunchAgents.

## Features

- Manage LaunchAgents with an intuitive interface
- View user-level and system-level services
- View service status
- Start/Stop user services
- View service logs
- Create user services

## Known Limitations

- Can only modify user-level services (~/Library/LaunchAgents)
- System services (/Library/LaunchDaemons, /System/Library/LaunchDaemons) are read-only
- Cannot stop services running as root
- Some system services may require Full Disk Access permission to view

## License

MIT License - see [LICENSE](LICENSE) file for details.
