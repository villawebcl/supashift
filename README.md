# supashift

Terminal tool PRO para manejar múltiples perfiles/cuentas de Supabase CLI sin login/logout global.

## Por qué Go
Se eligió Go por portabilidad real (Linux/macOS/Windows), binarios rápidos de distribuir y muy buen ecosistema para CLI/TUI (`cobra` + `bubbletea`).

## Features
- Perfiles múltiples: nombre, account label, notas, tags, aliases, favoritos.
- Vault seguro con prioridad keyring del SO y fallback cifrado (`age` + passphrase).
- Ejecución aislada por perfil:
  - `supashift run <profile> -- <cmd ...>`
  - `supashift shell <profile>`
  - `supashift tmux <profile>`
- Selector rápido `supashift pick` con búsqueda incremental.
- `use/unuse` para `eval` en shell actual.
- Integración de proyecto: detecta Supabase repo y mapping carpeta->perfil.
- `doctor` para diagnóstico de entorno y seguridad.

## Instalación
```bash
go build -o supashift ./cmd/supashift
```

Instalación rápida (Linux/macOS):
```bash
curl -fsSL https://raw.githubusercontent.com/villawebcl/supashift/main/install.sh | bash
```

## Quickstart
```bash
supashift init
supashift profile add anagami-prod --account "Anagami Production" --tags cliente,prod --aliases anagami,anagami-prod
supashift profile add anagami-dev --account "Anagami Dev" --tags cliente,dev --aliases anagami-dev
supashift profile ls

# Ejecutar comando aislado
supashift run anagami-prod -- supabase status

# Abrir subshell aislado
supashift shell anagami-prod

# tmux aislado
supashift tmux anagami-prod

# Inyectar en shell actual
eval "$(supashift use anagami-prod)"
# limpiar
eval "$(supashift unuse)"
```

## Seguridad
- Nunca usa ni modifica `~/.supabase/access-token`.
- Inyecta `SUPABASE_ACCESS_TOKEN` solo en el proceso/sesión objetivo.
- Tokens no se imprimen, salvo `supashift reveal <profile>` con confirmación `YES`.
- `export` no incluye secretos por defecto; requiere `--include-secrets`.
- Recomendación zsh para history:
```zsh
setopt HIST_IGNORE_SPACE
# luego ejecuta comandos sensibles con prefijo de espacio
```

## Comandos
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

## Config
Ubicación (XDG):
- Linux/macOS: `${XDG_CONFIG_HOME:-~/.config}/supashift/config.toml`

Estructura general:
- `profiles` metadata (sin token)
- `project_mappings`
- `recents`
- `vault_backend = auto|keyring|file`

## Import/Export
```bash
# sin secretos (default)
supashift export -o backup.toml

# con secretos explícitamente
supashift export --include-secrets -o backup-secrets.toml

supashift import -i backup.toml
```

## Integración zsh
```bash
supashift snippet zsh
supashift completions zsh > ~/.zsh/completions/_supashift
```

## Desarrollo
```bash
make fmt
make lint
make test
make build
```

## Documentación adicional
- Guía de uso e instalación multi-PC: `docs/USO_Y_DISTRIBUCION.md`
- Instalación universal: `docs/INSTALL.md`
- Plantilla AUR (bin): `packaging/PKGBUILD.supashift-bin.template`

## Smoke test (criterios de aceptación)
```bash
./scripts/smoke.sh ./supashift
```
Valida:
- 3 perfiles operando en paralelo por `run` con token aislado.
- No modificación de `~/.supabase/access-token`.
- `use/unuse`, `doctor` y migración desde token legacy.
- Sesiones tmux simultáneas (si `tmux` está instalado).

## Rendimiento
- `pick` carga perfiles en memoria y filtra incrementalmente.
- `run` agrega overhead mínimo (inyección env + spawn del comando).
