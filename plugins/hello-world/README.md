# Hello World Plugin

A comprehensive Hello World plugin for the Go CMS that demonstrates all the core plugin functionality.

## Features

- ✅ Basic "Hello World" API endpoint
- ✅ Message management (CRUD operations)
- ✅ Configurable settings through admin interface
- ✅ Admin menu integration
- ✅ Health check endpoint
- ✅ Multiple API endpoints
- ✅ Settings persistence
- ✅ Plugin lifecycle management

## API Endpoints

The plugin registers the following endpoints under `/api/v1/plugins/hello-world/`:

### Core Endpoints

- `GET /hello` - Returns welcome message
- `GET /info` - Plugin information and settings
- `GET /status` - Plugin status and available endpoints
- `GET /health` - Health check

### Message Management

- `GET /messages` - List all messages
- `POST /messages` - Create new message
- `GET /messages/:id` - Get specific message
- `DELETE /messages/:id` - Delete message

### Settings

- `GET /settings` - Get current settings
- `PUT /settings` - Update settings

## Admin Interface

The plugin adds a "Hello World" section to the admin menu with:

- Dashboard
- Messages management
- Settings configuration

## Settings

The plugin supports the following configurable settings:

| Setting         | Type    | Default                             | Description                   |
| --------------- | ------- | ----------------------------------- | ----------------------------- |
| enabled         | boolean | true                                | Enable/disable plugin         |
| welcome_message | text    | "Hello, World! Welcome to our CMS!" | Welcome message text          |
| show_timestamp  | boolean | true                                | Show timestamps with messages |
| max_messages    | number  | 10                                  | Maximum messages to store     |
| theme_color     | select  | "blue"                              | UI theme color                |

## Installation

1. Build the plugin:
   ```bash
   make build
   ```
