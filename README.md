# c3vm — C3 Language Version Manager

Manage multiple versions of the [C3 compiler](https://github.com/c3lang/c3c) on your machine.

Supports macOS and Linux only.

## Installation

```bash
go install github.com/matdexir/c3vm@latest
```

Or build from source:

```bash
git clone https://github.com/matdexir/c3vm
cd c3vm
go build -o c3vm .
```

Then add `~/.c3vm/bin` to your `PATH`:

```bash
c3vm init
```

## Usage

| Command | Description |
|---|---|
| `c3vm install <tag>` | Install a release by tag (e.g. `v0.8.0_2`) |
| `c3vm install latest` | Install the latest published release |
| `c3vm list` | List installed versions |
| `c3vm list-remote` | List available versions from GitHub |
| `c3vm use <version>` | Switch to an installed version |
| `c3vm current` | Show the active version |
| `c3vm default` | Show the default version |
| `c3vm default <version>` | Set the default version |
| `c3vm which [version]` | Show the path to the `c3c` binary |
| `c3vm remove <version>` | Uninstall a version |
| `c3vm init` | Print shell PATH setup instructions |

## How it works

All data lives in `~/.c3vm/`:

```
~/.c3vm/
├── versions/     # extracted compiler per version
├── current ->    # symlink to the active version
├── default       # plain text file with the default version
└── bin/c3c ->    # symlink to the active c3c binary
```

Run `c3vm use <version>` to switch. The `bin/c3c` symlink is updated so adding `~/.c3vm/bin` to your `PATH` is all you need.
