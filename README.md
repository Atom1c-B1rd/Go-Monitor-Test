# Go Monitor


Monitor de sistema en tiempo real para la terminal.

## Instalación

```bash
git clone https://github.com/tu-usuario/gomon
cd gomon
go mod tidy
```

## Correr

```bash
go run .
```

## Build

```bash
go build -o gomon .
./gomon
```

## Flags

| Flag | Default | Descripción |
|------|---------|-------------|
| `-r` | `1000` | Refresh en milisegundos |
| `-n` | `10` | Cantidad de procesos |

```bash
./gomon -r 500 -n 20
```
