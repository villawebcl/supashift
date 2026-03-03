# Supashift: Uso, Instalación y Distribución

## 1) Qué resuelve
`supashift` permite manejar múltiples perfiles de Supabase CLI sin login/logout global, inyectando `SUPABASE_ACCESS_TOKEN` por proceso/sesión.

## 2) Uso básico

### Inicializar
```bash
supashift init
```

### Agregar perfiles (1 por cuenta/correo)
```bash
supashift profile add cvv-gmail --account "cristianvillalobosvv@gmail.com"
supashift profile add cvv-outlook --account "cristianvillalobosv@outlook.com"
supashift profile add cvv-dev-gmail --account "crisvillalobosdev@gmail.com"
```

### Listar
```bash
supashift profile ls
```

### Ejecutar comando con perfil aislado
```bash
supashift run cvv-gmail -- supabase projects list
```

### Abrir shell aislado
```bash
supashift shell cvv-gmail
```

### tmux por perfil
```bash
supashift tmux cvv-gmail
supashift tmux --many
```

### Activar en shell actual
```bash
eval "$(supashift use cvv-gmail)"
# limpiar
eval "$(supashift unuse)"
```

### Selector interactivo
```bash
supashift pick
```

## 3) Migrar token existente de Supabase CLI
```bash
supashift migrate-from-supabase-cli cvv-gmail --source auto --account "cristianvillalobosvv@gmail.com"
```

Fuentes soportadas:
- `auto`: intenta archivo legacy y luego variable de entorno.
- `file`: lee token desde archivo.
- `env`: lee token desde variable (`SUPABASE_ACCESS_TOKEN` por defecto).
- `stdin`: lee token desde stdin.

## 4) Seguridad operativa
- Tokens en keyring del SO si está disponible.
- Fallback cifrado local con `age` + passphrase.
- `export` no incluye secretos salvo `--include-secrets`.
- `reveal` requiere confirmación explícita.
- Recomendado en zsh:
```zsh
setopt HIST_IGNORE_SPACE
```

## 5) Instalación en otros PCs

## Requisitos
- `supabase` CLI instalada.
- `docker` + permisos de usuario para comandos locales (`supabase start/status`).
- Opcional: `tmux`.

## Linux/macOS/Windows (manual)
1. Descargar binario de release según arquitectura.
2. Copiar a PATH (`/usr/local/bin`, `~/.local/bin`, etc).
3. Ejecutar:
```bash
supashift init
supashift doctor
```

## Compilar desde código
```bash
go mod tidy
go build -o supashift ./cmd/supashift
```

## 6) Portabilidad
Sí, puede funcionar bien en otros PCs:
- Está escrito en Go (binario portable por OS/arch).
- Usa rutas XDG y fallback seguro de vault.
- Backend keyring depende del entorno de escritorio/SO; si falla, usa archivo cifrado.

## 7) ¿Se puede publicar en AUR?
Sí.

Opciones comunes:
- `supashift-bin`: instala binarios precompilados desde GitHub Releases.
- `supashift`: compila desde fuente.

Pasos típicos:
1. Crear repo en GitHub con tags (`v0.1.0`, etc).
2. Publicar release con checksums SHA256.
3. Crear `PKGBUILD` en repo AUR (bin o source).
4. Mantener versión y checksums en cada release.

Referencia rápida de campos clave en PKGBUILD:
- `pkgname`, `pkgver`, `pkgrel`, `arch`, `url`, `license`.
- `depends` (ej. `supabase`, opcionalmente `tmux`).
- `source` + `sha256sums`.
- `package()` copiando `supashift` a `/usr/bin/supashift`.

## 8) Otros canales de distribución
- GitHub Releases (mínimo recomendado para empezar).
- Homebrew Tap (macOS/Linux).
- Scoop/Chocolatey (Windows).
- Nixpkgs (si quieres adopción dev avanzada).
- Containers (imagen CLI) para CI/CD.

## 9) ¿Marketplaces?
No hay un "market" central único para CLIs, pero sí ecosistemas:
- AUR (Arch).
- Homebrew (macOS/Linux).
- Scoop/Chocolatey (Windows).
- GitHub Marketplace aplica más a Actions/apps GitHub, no ideal como canal principal de esta CLI.

## 10) Roadmap sugerido para productizar
1. Releases automáticas multi-OS con GoReleaser.
2. Firma de artefactos (cosign/minisign).
3. Telemetría opt-in y anónima (latencia/errores).
4. Mejoras TUI: acciones con teclado, favoritos/recents persistentes más visibles.
5. Test matrix en CI (linux/mac/windows).
6. Empaquetado AUR/Homebrew/Scoop.

## 11) Flujo recomendado de versión
1. `CHANGELOG.md` actualizado.
2. Tag semver (`v0.x.y`).
3. CI verde (`make ci`, `make smoke`).
4. Publicar release con checksums.
5. Actualizar paquetes (AUR/Homebrew/Scoop).
