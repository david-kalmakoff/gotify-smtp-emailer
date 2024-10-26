# Gotify STMP Emailer

This Gotify plugin forwards all received messages to email through a provided SMTP server.

## Prerequisite

- An SMTP server to send messages through.

## Setup

### Installation

Build the plugin yourself or download a [binary release](https://github.com/david-kalmakoff/gotify-smtp-emailer/releases). Make the `.so` file available to Gotify in it's `pluginsdir` (default `/data/plugins`).

### Configuration

1. Launch Gotify and verify the plugin is loaded in the log:

```
Starting Gotify version 2.5.0@2024-06-23-17:12:59
Loading plugin data/plugins/gotify-smtp-emailer-linux-amd64.so
Started listening for plain connection on tcp [::]:80
```

2. Navigate to "Clients" and create a new client and copy the token
3. Navigate to "Plugins" and click the :gear: icon for the "Gotify SMTP Emailer"
4. Fill out "Configurer" information and click "Save"

```yaml
hostname: ws://localhost # Keep this as localhost because plugin is running with Gotify
token: <client_token> # Token from step 2
smtp:
  host: <smtp_host> # SMTP server host
  port: <587|465|25> # SMTP server port
  fromemail: <from_email> # Email to send message from / SMTP username
  password: <password> # Password for SMTP server
  toemails:
    - <to_email> # List of emails to send messages to
  subject: Gotify Notification # Prefix to email subjects that are send
  insecure: false # SMTP without TLS
environment: production # Used to send test messages in development
```

5. Navigate to "Plugins" and enable the "Gotify SMTP Emailer" plugin
6. Done

## Development

Development is done with Docker. You can run a development environment with the command:

```bash
make local
```

You will need to set the following "Configurer" information for the plugin. All others can stay the same.

```yaml
token: <client_token>
smtp:
  host: mailhog
  port: 1025
  insecure: true
environment: development
```

This will build the plugin and start up ephemeral instances of Gotify (with plugin loaded) and Mailhog. You can use this to manually test the plugin.

## Testing

Testing is done with Docker. You can run the test suite with the command:

```bash
make test
```

## Building

For building the plugin gotify/build docker images are used to ensure compatibility with
[gotify/server](https://github.com/gotify/server).

`GOTIFY_VERSION` can be a tag, commit or branch from the gotify/server repository.

This command builds the plugin for amd64, arm-7 and arm64.
The resulting shared object will be compatible with gotify/server version 2.0.20.

```bash
make GOTIFY_VERSION="v2.0.20" FILE_SUFFIX="for-gotify-v2.0.20" build
```
