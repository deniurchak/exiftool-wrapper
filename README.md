# exiftool-wrapper

## Description

This is a Go HTTP server that wraps the output of the ExifTool command line tool.

It exposes a `/tags` endpoint. The endpoints return the ouput of this command:

```
exiftool -listx
```

The response is a JSON array of all available ExifTool tags with their metadata.

## Running the server

To run the server, clone the repository and run the following command to install the dependencies:

```
go mod tidy
```

Then run the server with the following command:

```
go run main.go
```

The server will listen on port 8080.

IMPORTANT: The server requires the ExifTool binary to be installed on the machine.
How to install ExifTool on a Mac:

```
https://exiftool.org/install.html
```

## Testing:

```
go test ./...
```

IMPORTANT: testing will fail if the ExifTool binary is not installed on the machine.
