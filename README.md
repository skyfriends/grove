<p align="center">
  <img src="assets/banner.svg" alt="grove" width="840"/>
</p>

<p align="center">
  <strong>A fast CLI for managing local git repositories and worktrees.</strong>
</p>

<p align="center">
  <a href="https://github.com/skyfriends/grove/releases"><img src="https://img.shields.io/github/v/release/skyfriends/grove?style=flat-square&color=98c379" alt="Release"></a>
  <a href="https://github.com/skyfriends/grove/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-5c6370?style=flat-square" alt="License"></a>
  <a href="https://github.com/skyfriends/grove/actions"><img src="https://img.shields.io/github/actions/workflow/status/skyfriends/grove/ci.yml?style=flat-square&color=61afef" alt="CI"></a>
</p>

---

Grove gives you a single dashboard across all your local repos. See what branch everything is on, which worktrees are lying around, what's dirty, what's stale - then clean it all up without leaving the terminal.

## Install

**Go**

```sh
go install github.com/skyfriends/grove@latest
```

**Binary**

Download from [releases](https://github.com/skyfriends/grove/releases) and put it on your PATH.

**From source**

```sh
git clone https://github.com/skyfriends/grove.git
cd grove
make install
```

## Quick start

```sh
grove init        # create config (optional - works without it)
grove             # see everything
```

## Commands

### Browse

```sh
grove                 # list repos, branches, worktrees
grove status          # ahead/behind counts relative to upstream
grove stale           # branches with no recent activity
```

### Maintain

```sh
grove sync            # pull latest on clean repos at their default branch
grove prune           # remove stale remote-tracking references
grove clean           # interactively pick worktrees to remove
grove nuke            # bulk remove all clean worktrees for selected repos
grove reset           # switch repos back to their default branch
```

### Create

```sh
grove wt cray feat/new-thing    # spin up a worktree
grove clone skyfriends/grove        # clone into your root dir
grove open cray                 # open in your editor (fuzzy matches)
```

## Config

Grove works out of the box with zero configuration. Run `grove init` to create `~/.grove.toml` if you want to customize:

```toml
# Root directory containing your git repositories
root = "~/projects"

# Editor for 'grove open' (defaults to $EDITOR, then "code")
# editor = "code"

# Branches inactive for this many days are flagged by 'grove stale'
stale_days = 30
```

| Key | Default | Description |
|-----|---------|-------------|
| `root` | `~/projects` | Directory containing your repos |
| `editor` | `$EDITOR` / `code` | Editor command for `grove open` |
| `stale_days` | `30` | Days threshold for `grove stale` |

## How it works

- **Parallel everything** - repos are scanned concurrently, worktree dirty checks run in parallel, sync pulls happen simultaneously
- **Worktree-aware** - distinguishes real repos (`.git/` directory) from worktrees (`.git` file) so nothing gets double-counted or misidentified
- **Safe by default** - dirty worktrees are never removed, dirty repos are skipped during sync and reset
- **Conventional commits** - releases are driven by commit messages, changelogs are generated automatically

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE)
