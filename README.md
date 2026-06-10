# Go Monitor


Monitor de sistema en tiempo real para la terminal.

## Instalación

```bash
git clone https://github.com/Atom1c-B1rd/Go-Monitor-Test
cd Go-Monitor-Test
go mod tidy
```

## Correr

```bash
go run .
```

## Build

```bash
go build -o Go-Monitor .
./Go-Monitor
```

## Flags

| Flag | Default | Descripción |
|------|---------|-------------|
| `-r` | `1000` | Refresh en milisegundos |
| `-n` | `10` | Cantidad de procesos |

```bash
./gomon -r 500 -n 20
```
