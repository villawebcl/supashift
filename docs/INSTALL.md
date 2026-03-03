# Instalación rápida

## Script universal (Linux/macOS)
```bash
curl -fsSL https://raw.githubusercontent.com/villawebcl/supashift/main/install.sh | bash
```

## Opciones útiles
```bash
# versión específica
curl -fsSL https://raw.githubusercontent.com/villawebcl/supashift/main/install.sh | bash -s -- --version v0.1.2

# instalar en /usr/local/bin
curl -fsSL https://raw.githubusercontent.com/villawebcl/supashift/main/install.sh | bash -s -- --system

# directorio custom
curl -fsSL https://raw.githubusercontent.com/villawebcl/supashift/main/install.sh | bash -s -- --bin-dir "$HOME/bin"
```

El script:
- Detecta OS/arquitectura (`linux|darwin`, `x86_64|arm64`).
- Descarga binario correcto desde GitHub Releases.
- Verifica SHA256 contra `checksums.txt`.
- Instala `supashift` en tu directorio objetivo.
