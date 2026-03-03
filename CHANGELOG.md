# Changelog

## 0.1.0 - 2026-03-03
- Primera versión funcional de `supashift`.
- Gestión de perfiles (add/edit/rm/ls) con aliases/tags/notas/favoritos.
- Vault seguro: keyring + fallback cifrado con age.
- Ejecución aislada por perfil: run/shell/tmux.
- Selector TUI `pick`.
- `use/unuse`, `doctor`, `import/export`, `auto`, `project bind`.
- `migrate-from-supabase-cli` para importar token legacy sin exponerlo.
- `scripts/smoke.sh` para validar criterios de aceptación rápidamente.
- Tests base (vault + run integration), Makefile y CI.
