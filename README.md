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
supashift profile add client-a --account "user1@example.com"
supashift profile add client-b --account "user2@example.com"

# List profiles
supashift profile ls

# Run with isolated account token
supashift run client-a -- supabase projects list

# Optional: activate in current shell
eval "$(supashift use client-a)"
supabase projects list
eval "$(supashift unuse)"
```

## Commands (EN / ES)

### EN
- `supashift init`: initialize Supashift config.
- `supashift profile add|edit|rm|ls`: create, edit, remove, list profiles.
- `supashift run <profile> -- <command>`: run one command with isolated token.
- `supashift shell <profile>`: open a shell with active profile.
- `supashift tmux <profile> [--many]`: create/attach tmux session(s) per profile.
- `supashift pick`: interactive profile picker.
- `supashift use <profile>` / `supashift unuse`: activate/clear profile in current shell.
- `supashift doctor`: check environment health.
- `supashift export [--include-secrets]` / `supashift import --input file.toml`: backup/restore config.
- `supashift migrate-from-supabase-cli <profile> [--source auto|file|env|stdin]`: import legacy token.
- `supashift auto [--set]`: suggest profile from current project path.
- `supashift project bind <profile>` / `supashift project ls`: map project folders to profiles.
- `supashift reveal <profile>`: reveal profile token (sensitive).

### ES
- `supashift init`: inicia la configuracion de Supashift.
- `supashift profile add|edit|rm|ls`: crea, edita, elimina y lista perfiles.
- `supashift run <perfil> -- <comando>`: ejecuta un comando con token aislado.
- `supashift shell <perfil>`: abre una shell con el perfil activo.
- `supashift tmux <perfil> [--many]`: crea/adjunta sesion(es) tmux por perfil.
- `supashift pick`: selector interactivo de perfiles.
- `supashift use <perfil>` / `supashift unuse`: activa/limpia perfil en la shell actual.
- `supashift doctor`: revisa el estado del entorno.
- `supashift export [--include-secrets]` / `supashift import --input file.toml`: respaldo/restauracion de configuracion.
- `supashift migrate-from-supabase-cli <perfil> [--source auto|file|env|stdin]`: importa token legacy.
- `supashift auto [--set]`: sugiere perfil segun la carpeta actual.
- `supashift project bind <perfil>` / `supashift project ls`: vincula carpetas de proyecto a perfiles.
- `supashift reveal <perfil>`: muestra token del perfil (sensible).

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
