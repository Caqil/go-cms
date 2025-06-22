# Test Plugin Plugin

A test-plugin plugin for the CMS

## Features

- WordPress-like plugin architecture
- Hot-reloadable without server restart
- Dynamic route registration
- Admin interface integration
- Configurable settings

## Installation

1. Create a zip file with the plugin contents:
   ```bash
   make zip
   ```

2. Upload via admin interface or API:
   ```bash
   curl -X POST \
     -F "plugin=@test-plugin.zip" \
     http://localhost:8080/api/v1/admin/plugins/upload \
     -H "Authorization: Bearer YOUR_TOKEN"
   ```

## API Endpoints

After installation, the plugin exposes these endpoints:

- `GET /api/v1/plugins/test-plugin/` - Plugin index
- `GET /api/v1/plugins/test-plugin/info` - Plugin information
- `GET /api/v1/plugins/test-plugin/status` - Plugin status
- `POST /api/v1/plugins/test-plugin/action` - Execute actions

## Admin Interface

The plugin adds menu items to the admin dashboard:

- **Test Plugin** > Dashboard
- **Test Plugin** > Settings

## Development

1. Build the plugin:
   ```bash
   make build
   ```

2. Test the plugin:
   ```bash
   make test
   ```

3. Create distributable package:
   ```bash
   make zip
   ```

## Configuration

The plugin supports these settings:

- **enabled**: Enable/disable plugin functionality
- **auto_update**: Automatic updates when available
- **cache_ttl**: Cache time-to-live in seconds
- **debug_mode**: Enable debug logging

## Author

Plugin Developer

## License

MIT License
