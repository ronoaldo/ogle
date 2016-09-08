# Simple toolkit to interact with Google APIs

Warning: this is a work in progress, subject to change.

# Install

To install these tools you need the Go programming language
in your machine. Follow the setup instructions from the
official website https://golang.org/doc/install

If you have Debian/Ubuntu you can also run:

	apt-get install golang-go

Once Go is installed, you can then install the commands
by running:

	go install github.com/ronoaldo/ogle/cmd/...

# Commands

Each command is designed to be used as a standalone command
line tool, and be integrated with any Unix toolchain.

## Client ID and Client Secret

You must supply to all commands the Client ID and Client
Secret to be used when performing API calls.

Obtain these credentials from https://console.developers.google.com/

Credentials are cached per command, so you don't have to
authenticate after each invocation.

## youtube

youtube is a command line interface to interact with your
youtube channel.

For usage run `youtube --help`
