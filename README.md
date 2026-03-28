# Gorsh 🐚

A Unix shell built from scratch in Go — because the best way to understand a tool you use every day is to build it yourself.

---

## Motivation

Every developer lives in a shell. But how many know what actually happens between pressing Enter and seeing output?

I built Gorsh to find out. Not to replace bash or zsh — but to understand the primitives underneath: how pipelines wire processes together with OS-level pipes, how builtins run _inside_ the shell process instead of forking, how a trie makes tab completion instant, and how history persists across sessions.

The result is a real, interactive shell that you can run, use, and read.

---

## Quick Start

**Requires Go 1.21+** and a Unix-like system (Linux, macOS, FreeBSD).

```bash
# Clone and build
git clone https://github.com/andrei-himself/gorsh.git
cd gorsh
go build -o gorsh

# Run with persistent history
HISTFILE=~/.gorsh_history ./gorsh
```

You're now in a Gorsh session:

```bash
$ echo hello world
hello world
$ history
    1  echo hello world
    2  history
```

---

## Usage

### Pipelines

Builtins and external commands can be freely mixed in a pipeline:

```bash
$ ls -1 | wc -l
8
$ echo "hello world" | tr a-z A-Z
HELLO WORLD
```

Gorsh handles builtins correctly in pipelines too — they run in goroutines so pipe I/O never deadlocks, even when a builtin appears in the middle of a chain.

### I/O Redirection

```bash
$ echo hello > out.txt          # overwrite stdout
$ echo world >> out.txt         # append stdout
$ cat missing 2> err.txt        # redirect stderr
$ cat missing >> out.txt 2>&1   # append both
```

Redirect operators can appear anywhere in the command line.

### Tab Completion

- **Commands** — prefix-matches all builtins and every executable on `$PATH` using a trie. A single match completes immediately. Multiple matches ring the bell on the first Tab, then print the full list on the second.
- **Arguments** — prefix-matches filenames and directories. Directories are suffixed with `/` so you can keep tabbing deeper.

### History

```bash
$ history              # show all entries
$ history 5            # show the last 5 entries
$ history -r <file>    # append entries from a file into memory
$ history -w <file>    # write all in-memory entries to a file (overwrite)
$ history -a <file>    # append only new entries to a file (since last -a)
```

Entries are numbered from 1 and reflect their absolute position in the full history — not just the slice being displayed.

### Persistent History

```bash
HISTFILE=~/.gorsh_history ./gorsh
```

- **On startup** — Gorsh loads the file into memory so previous sessions are immediately available.
- **On exit** — all in-memory entries, including the `exit` command itself, are written back to the file.

### Builtins

| Command                   | Description                                              |
| ------------------------- | -------------------------------------------------------- |
| `echo [args...]`          | Print arguments to stdout                                |
| `cd [path]`               | Change directory; `~` expands to `$HOME`                 |
| `pwd`                     | Print the current working directory                      |
| `type <name>`             | Show whether `name` is a builtin or a `$PATH` executable |
| `history [n]`             | List history; optionally limit to the last `n` entries   |
| `history -r/-w/-a <file>` | Read, write, or append history to a file                 |
| `exit`                    | Save history and exit                                    |

---

## Contributing

### Clone the repo

```bash
git clone https://github.com/andrei-himself/gorsh.git
cd gorsh
```

### Build the binary

```bash
go build -o gorsh
```

### Run it

```bash
HISTFILE=~/.gorsh_history ./gorsh
```

### Submit a pull request

If you'd like to contribute, please fork the repository and open a pull request to the `main` branch. Keep changes focused and include a short description of what you changed and why.

---

## License

MIT — see [LICENSE](LICENSE) for details.
