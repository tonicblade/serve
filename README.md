# serve
## simple http endpoint & resource viewer

a simple local file server for browsing a folder in the browser.

the idea is plain but useful, you start it in a directory, open it in your browsser, and browse contents without extra setup or fluff. it works amazingly for sharing files over the same network and basic uploads.

## what it does

- serves a directory over HTTP
- shows a simple directory listing with breadcrumbs and folders on the side
- supports nested folder browsing with a depth setting in the top right
- lets you upload files into the current directory
- opens the browser automatically when requested
- supports local network access for other devices on the same LAN
- works as a lightweight local file host with no extra dependencies

## quick start

from root:

```bash
go run . .
```

or you can build it:

```bash
go build -o serve ./cmd/serve
```

then open the local url shown in terminal, usually:

```text
http://localhost:8080
```

## cli usage

```bash
serve [path] [flags]
```

### flags

```text
--host <host>      bind address (default: 0.0.0.0)
--port <port>      port to listen on (default: 8080)
--open             open in browser after startup
--quiet            suppress banner output
--upload           allow file uploads
--depth <n>        folder tree depth (default: 2)
--https            serve with TLS on :8443
--once             serve a single request then exit
--expires <dur>    auto-shutdown after duration like 30m
--help             show help
```

## examples

```bash
serve .
serve ./dist
serve ./dist --port 3000
serve ./dist --host 127.0.0.1
serve . --depth 4
serve . --open
serve . --upload
```

## describe it like im 5

a small local file server for browsing a folder in the browser, with a simple tree view, folder previews, uploads, and basic LAN sharing.

## notes

- default root is the current working directory // where serve is executed
- the page is intentionally minimal and readable
- uploads are kept simple and target the folder you are currently browsing
- video files are sent with inline browser-friendly headers when possible (no guarantees)

## install it globally so you can run `serve` anywhere

### windows

1. install go if you do not already have it:
   - download and install from https://go.dev/dl/
   - make sure `go` is added to your PATH

2. build the binary:

```powershell
go build -buildvcs=false -o serve.exe ./cmd/serve
```

3. move it into a folder already on your PATH, such as:

```powershell
mkdir "$env:USERPROFILE\bin" -Force
move .\serve.exe "$env:USERPROFILE\bin\serve.exe"
```

4. add that folder to your PATH if it is not already there:

```powershell
[Environment]::SetEnvironmentVariable(
  "Path",
  "$env:USERPROFILE\bin;" + [Environment]::GetEnvironmentVariable("Path", "User"),
  "User"
)
```

5. restart your terminal and run:

```powershell
serve --help
```

You should now be able to run `serve` from any folder.

### linux

1. install go if needed:

```bash
sudo apt update
sudo apt install golang-go
```

or use your distro package manager if you prefer.

2. build the binary:

```bash
go build -buildvcs=false -o serve ./cmd/serve
```

3. move it to a directory in your PATH, commonly `/usr/local/bin`:

```bash
sudo install -m 0755 ./serve /usr/local/bin/serve
```

4. verify it is available:

```bash
serve --help
```

You should now be able to run `serve` from anywhere on the system.

### macos

```bash
go build -buildvcs=false -o serve ./cmd/serve
sudo install -m 0755 ./serve /usr/local/bin/serve
serve --help
```

this is the same pattern as linux and works well if you want it available globally on a mac too.