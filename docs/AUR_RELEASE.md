# Publicar en AUR (supashift-bin)

Este flujo evita errores de checksum y deja el paquete instalable con `yay`/`paru`.

## 1) Publicar release en GitHub
1. Asegura CI verde: `make ci && make smoke`.
2. Crea tag semver:
```bash
git tag v0.1.0
git push origin v0.1.0
```
3. El workflow `release.yml` generará artefactos y `checksums.txt`.

## 2) Generar PKGBUILD y .SRCINFO para AUR
```bash
./scripts/update-aur-bin.sh 0.1.0 villawolf supashift "Cristian Villalobos" "cristian@example.com"
```

Archivos generados:
- `packaging/aur/supashift-bin/PKGBUILD`
- `packaging/aur/supashift-bin/.SRCINFO`

## 3) Crear repo AUR y subir
```bash
cd /tmp
git clone ssh://aur@aur.archlinux.org/supashift-bin.git
cd supashift-bin
cp /home/villawolf/Documentos/workspace/supashift/packaging/aur/supashift-bin/PKGBUILD .
cp /home/villawolf/Documentos/workspace/supashift/packaging/aur/supashift-bin/.SRCINFO .
cp /home/villawolf/Documentos/workspace/supashift/LICENSE .
git add PKGBUILD .SRCINFO LICENSE
git commit -m "supashift-bin 0.1.0"
git push
```

## 4) Instalar desde AUR
```bash
yay -S supashift-bin
# o
paru -S supashift-bin
```

## 5) Checklist anti-errores
- Tag `vX.Y.Z` existe en GitHub.
- Artefactos linux `x86_64` y `arm64` publicados.
- Checksums en `PKGBUILD` coinciden con `checksums.txt` del release.
- `.SRCINFO` regenerado tras cambiar `PKGBUILD`.
