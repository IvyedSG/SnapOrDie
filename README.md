# SnapOrDie

Snapshots instantáneos de bases Docker usando copy-on-write.

[![Release](https://img.shields.io/github/v/release/IvyedSG/SnapOrDie)](https://github.com/IvyedSG/SnapOrDie/releases)
[![Build](https://img.shields.io/github/actions/workflow/status/IvyedSG/SnapOrDie/release.yml)](https://github.com/IvyedSG/SnapOrDie/actions)
[![Go Version](https://img.shields.io/github/go-mod/go-version/IvyedSG/SnapOrDie)](https://golang.org)

---

## El problema

Restaurás un dump SQL de 2.8 GB para debuggear un bug. Tarda **3 minutos**. Encontrás el problema, lo arreglás, y necesitás probar de nuevo. Otros 3 minutos. Y otra vez. Y otra vez.

SnapOrDie guarda todo el directorio de datos de MySQL y lo restaura en **menos de un segundo** usando APFS copy-on-write (macOS) o reflink (Linux). Nada de imports SQL. Nada de esperar.

```bash
# Guardá un estado limpio una sola vez
snapordie save clean

# Debuggeá, rompé, arreglá, repetí — reset en milisegundos
snapordie reset clean
```

---

## Instalación

### Go install

```bash
go install github.com/IvyedSG/SnapOrDie@latest
```

### Descargar binary

Descargá la última versión de [GitHub Releases](https://github.com/IvyedSG/SnapOrDie/releases), extraé y mové a tu `$PATH`:

```bash
tar xzf snapordie_*.tar.gz
sudo mv snapordie /usr/local/bin/
```

### Compilar desde fuente

```bash
git clone https://github.com/IvyedSG/SnapOrDie.git
cd SnapOrDie
go build -o snapordie .
sudo mv snapordie /usr/local/bin/
```

---

## Cómo empezar

### 1. Guardá tu primer snapshot

```bash
# Auto-detecta el container MySQL/MariaDB corriendo
snapordie save clean
```

```
 ── Guardar ─────────────────────────────────────
 ◇ Container  emusa_mysql
 · Data dir   /Users/tu/proyecto/data/mysql
 ◆ Deteniendo container                       215ms
 ◆ Guardando snapshot  clean  6.3 GB          1.3s
 ◆ Iniciando container                        420ms
 ◆ Snapshot "clean" listo
```

### 2. Trabajá, rompé, ejecutá migrations

```bash
pnpm prisma:deploy
# ... test, debug, repetí ...
```

### 3. Volvé al snapshot al instante

```bash
snapordie reset clean
```

```
 ── Restaurar ───────────────────────────────────
 ◇ Snapshot  clean  (el más reciente)
 · Container  emusa_mysql
 ◆ Deteniendo container                       210ms
 ◆ Restaurando snapshot                        80ms
 ◆ Iniciando container                        410ms
 ◆ Base de datos restaurada a "clean"
 · Ejecutá migrations si cambió el schema
```

---

## Comandos

| Comando | Alias | Descripción |
|---------|-------|-------------|
| `save [nombre]` | `sv` | Guardar un snapshot |
| `reset [nombre]` | `rs` | Restaurar base a un snapshot |
| `list` | `ls` | Listar snapshots |
| `info <nombre>` | `show` | Ver detalle de un snapshot |
| `rm <nombre>` | `del` | Eliminar un snapshot |
| `version` | — | Mostrar versión |

### Flags

| Flag | Descripción |
|------|-------------|
| `--container <nombre>` | Nombre del container Docker (se auto-detecta si está vacío) |
| `--no-color` | Desactivar salida con colores |

### Ejemplos

```bash
# Guardar con nombre personalizado
snapordie sv antes-de-migration

# Restaurar un snapshot específico
snapordie rs bug-1234

# Listar snapshots en tabla
snapordie ls

# Ver detalle de un snapshot
snapordie show clean

# Eliminar snapshots viejos
snapordie del estado-experimental

# Especificar container manualmente
snapordie save --container emusa_mysql
```

---

## Cómo funciona

1. **Detecta** — busca un container MySQL/MariaDB corriendo via `docker ps` e inspecciona sus mounts para ubicar el directorio de datos en el host.

2. **Detiene** — frena el container para que el directorio de datos quede consistente.

3. **Clona** — copia el directorio de datos usando copy-on-write:
   - macOS: `cp -c` (APFS nativo, instantáneo)
   - Linux: `cp --reflink=always` (btrfs/xfs, instantáneo)
   - Fallback: `cp -a` (copia normal, más lenta)

4. **Inicia** — reinicia el container y espera a que esté healthy.

Los snapshots viven en `.snapordie/` junto al directorio de datos de MySQL. Cada snapshot es un clon completo del directorio. En filesystems CoW no ocupan espacio extra hasta que modificás datos.

```
data/mysql/
├── ...                    # datos MySQL activos
└── .snapordie/
    ├── manifest.json       # registro de snapshots
    ├── clean/              # snapshot: directorio MySQL completo
    └── antes-de-migration/ # otro snapshot
```

---

## Flujo de trabajo real

Así se usa SnapOrDie en un ciclo de debugging con Prisma + MySQL:

```bash
# 1. Restaurá fresco desde SQL dump (lento, una sola vez)
./restore-backup.sh staging_dump.sql
pnpm prisma:deploy

# 2. Guardá el estado limpio
snapordie save clean

# 3. Iterá sobre un bug
# Cambiás código → probás → encontrás bug → arreglás → probás de nuevo
# ¿Base sucia? Reseteá.
snapordie rs clean

# 4. Antes de una migration riesgosa
pnpm prisma:migrate-dev
snapordie sv antes-de-migration

# 5. ¿La migration rompió algo? Volvé.
snapordie rs clean
```

---

## Notas técnicas

- Solo MySQL/MariaDB (auto-detectado). Soporte para PostgreSQL planeado.
- Requiere Docker instalado y corriendo.
- Los WAL y redo logs están en el directorio de datos y se snapshotearn tal cual.
- Funciona con cualquier bind mount de MySQL en `/var/lib/mysql`.
- El flag `--container` sobreescribe la auto-detección para setups multi-container.

---

## Licencia

MIT
