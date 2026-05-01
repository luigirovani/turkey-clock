# Turkey Clock

Servidor em Go para consultar horario via NTP e expor:

- pagina web com relogio sincronizado
- endpoint JSON de horario atual
- metricas basicas da resposta NTP (stratum, RTT, root distance, etc.)

## Funcionalidades

- consulta de horario em servidor NTP primario com fallback
- API HTTP para consumo por apps/scripts
- pagina web com atualizacao periodica do horario
- suporte a timezone por query string
- formatacao customizavel de data/hora e duracoes
- deploy simples com Docker e Docker Compose

## Estrutura do projeto

```text
.
|-- docker-compose.yaml
|-- Dockerfile
|-- go.mod
|-- server/
|   |-- main.go
|   |-- assets/
|   |   |-- index.html
|   |   |-- script.js
|   |   `-- style.css
|   `-- ntputils/
|       |-- config.go
|       `-- formats.go
```

## Requisitos

- Go (versao definida em `go.mod`)
- Docker (opcional)
- Docker Compose (opcional)

## Executando localmente (Go)

Na raiz do projeto:

```bash
go run ./server
```

Alternativa, entrando na pasta `server`:

```bash
go run main.go
```

Servidor padrao: `http://localhost:8080`

## Executando com Docker

Build e run direto:

```bash
docker build -t turkey-clock .
docker run --rm -p 8080:8080 --name turkey-clock turkey-clock
```

Com Docker Compose:

```bash
docker compose up --build
```

## Endpoints

- `GET /` pagina web
- `GET /time` API principal de horario
- `GET /current_time` alias de `/time`
- `GET /get_current_time` alias de `/time`

### Query params da API `/time`

- `timezone`: timezone IANA (ex.: `America/Sao_Paulo`)
- `timestamp`: se `true`, retorna `current_time` como UNIX timestamp
- `format`: formato de data no padrao Go (ex.: `2006-01-02 15:04:05`)
- `precision_unit`: unidade para metricas de duracao da resposta NTP
  - `auto` (padrao)
  - `ms`
  - `us`
  - `ns`

## Exemplos

Retorno com timezone:

```bash
curl "http://localhost:8080/time?timezone=America/Sao_Paulo"
```

Retorno com timestamp unix:

```bash
curl "http://localhost:8080/time?timestamp=true"
```

Formato customizado:

```bash
curl "http://localhost:8080/time?format=2006-01-02%2015:04:05"
```

## Exemplo de resposta

```json
{
  "current_time": "2026-05-01T12:34:56.789Z",
  "time_zone": "America/Sao_Paulo",
  "timestamp": 1777635296,
  "datetime": "2026-05-01T12:34:56Z",
  "ntp_response": {
    "time": "2026-05-01T12:34:56.789Z",
    "server": "pool.ntp.org",
    "unit_time": "ms",
    "offset": "1.234",
    "precision": "-20 ns",
    "root_dispersion": "2.345",
    "root_distance": "3.456",
    "rtt": "4.567",
    "stratum": 2
  }
}
```

> Observacao: se o NTP primario e fallback falharem, o servidor cai para horario local UTC e `ntp_response` pode vir `null`.

## Configuracao

### Variaveis de ambiente

- `NTP_HOST`: host do servidor NTP primario (ip/domino:porta)
- `NTP_DOMAIN`: dominio NTP padrao (default: `turkey-clock.aecrypto.io`)
- `NTP_FALLBACK`: servidor NTP fallback (default: `pool.ntp.org`)
- `HOST`: host HTTP (default: `0.0.0.0`)
- `PORT`: porta HTTP (default: `8080`)
- `GA`: Google Analytics ID
- `LOG_LEVEL`: `debug`, `info`, `warn`, `error`

### Flags de linha de comando

- `--ntp-host`
- `--ntp-domain`
- `--ntp-fallback`
- `--host`
- `--port`
- `--ga`
- `--log-level`

Exemplo:
go mod tidy
```bash
go run ./server --ntp-domain=pool.ntp.org --log-level=debug --port=8081
```

## Logs

O servidor usa `slog` e registra:

- inicio da aplicacao
- requests HTTP (IP, metodo, path, status, duracao)
- tentativas/falhas de consulta NTP

## Licenca

Defina aqui a licenca do projeto (ex.: MIT).
