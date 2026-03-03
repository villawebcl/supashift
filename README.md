# supashift

Manage multiple Supabase CLI accounts/profiles with zero login/logout friction.

`supashift` isolates `SUPABASE_ACCESS_TOKEN` per command/session, so you can work across clients/projects safely without touching the global Supabase CLI token state.

## Why
Working with multiple Supabase accounts is painful when switching auth context repeatedly.

`supashift` solves this by:
- Storing one token per profile (secure vault backend).
- Running commands with profile-scoped environment injection.
- Keeping sessions isolated (`run`, `shell`, `tmux`, `use/unuse`).

## Key Features
- Multi-profile management: name, account label, notes, tags, aliases, favorites.
- Secure vault:
  - Primary: OS keyring (Secret Service/Keychain/Credential Manager).
  - Fallback: local encrypted file (`age` + passphrase).
- Isolated execution:
  - `supashift run <profile> -- <cmd...>`
  - `supashift shell <profile>`
  - `supashift tmux <profile>`
- Fast UX:
  - Interactive picker `supashift pick`.
  - `use/unuse` snippets for current shell.
- Project integration:
  - Detect Supabase project directories.
  - Folder-to-profile mapping.

## Installation

### Arch Linux (AUR)
```bash
yay -S supashift-bin
# or
paru -S supashift-bin
```

### Universal installer (Linux/macOS)
```bash
curl -fsSL https://raw.githubusercontent.com/villawebcl/supashift/main/install.sh | bash
```

### Build from source
```bash
go mod tidy
go build -o supashift ./cmd/supashift
```

## Quickstart
```bash
supashift init

# Add profiles (token prompt is hidden)
supashift profile add cvv-gmail --account "cristianvillalobosvv@gmail.com"
supashift profile add cvv-outlook --account "cristianvillalobosv@outlook.com"

# List profiles
supashift profile ls

# Run with isolated account token
supashift run cvv-gmail -- supabase projects list

# Optional: activate in current shell
eval "$(supashift use cvv-gmail)"
supabase projects list
eval "$(supashift unuse)"
```

## Core Commands
- `supashift init`
- `supashift profile add|edit|rm|ls`
- `supashift run <profile> -- <command>`
- `supashift shell <profile>`
- `supashift tmux <profile> [--many]`
- `supashift pick`
- `supashift doctor`
- `supashift export [--include-secrets]`
- `supashift import --input file.toml`
- `supashift migrate-from-supabase-cli <profile> [--source auto|file|env|stdin]`
- `supashift use <profile>` / `supashift unuse`
- `supashift auto [--set]`
- `supashift project bind <profile>` / `supashift project ls`
- `supashift reveal <profile>`

## Security Notes
- Does **not** modify `~/.supabase/access-token`.
- Injects `SUPABASE_ACCESS_TOKEN` only into target process/session.
- Never prints token unless explicitly requested with `reveal` + confirmation.
- `export` excludes secrets by default.
- Recommended for zsh history hygiene:
```zsh
setopt HIST_IGNORE_SPACE
```

## Docs
- Usage and distribution: `docs/USO_Y_DISTRIBUCION.md`
- Universal install details: `docs/INSTALL.md`
- AUR release flow: `docs/AUR_RELEASE.md`
- Security model: `SECURITY.md`

## Development
```bash
make fmt
make lint
make test
make build
make smoke
```

## License
MIT
