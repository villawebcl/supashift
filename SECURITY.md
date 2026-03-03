# Security Model

## Objetivo
`supashift` aísla tokens de Supabase por proceso/sesión sin alterar el estado global de Supabase CLI.

## Decisiones de seguridad
- Tokens preferentemente en keyring nativo del sistema.
- Fallback cifrado con `age` + passphrase (`SUPASHIFT_PASSPHRASE` o prompt interactivo).
- Tokens nunca en texto plano en config.
- Exportación sin secretos por defecto.
- `reveal` requiere confirmación explícita (`YES`).

## Threats considerados
- Exposición accidental en terminal/history.
- Mezcla de credenciales entre proyectos/sesiones.
- Lectura de config local por permisos laxos.

## Mitigaciones
- Archivos locales con permisos `0600` y carpetas `0700`.
- Sanitización env (sobrescribe `SUPABASE_ACCESS_TOKEN` para procesos hijos).
- Recomendación de shell para history segura (`HIST_IGNORE_SPACE`).

## Límites
- Un proceso con permisos del mismo usuario podría inspeccionar entorno de procesos activos según configuración del SO.
- Si `--token` se pasa en CLI, puede quedar en history (se recomienda prompt interactivo).

## Reporte de vulnerabilidades
Abrir issue de seguridad privado o contactar maintainers antes de disclosure pública.
