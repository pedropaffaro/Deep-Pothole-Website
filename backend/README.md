# 🕳️ Deep Pothole — Backend

API REST em Go responsável por receber denúncias de buracos nas vias, armazenar as fotos no Cloudflare R2 e salvar os dados relacionais no Cloudflare D1.



## Tecnologias

- **Cloudflare R2** — armazenamento de imagens
- **Cloudflare D1** — banco de dados SQLite na edge



## Estrutura do Projeto

```
.
├── cmd/
│   └── api/
│       └── main.go           # Entrypoint — inicializa o Fiber e registra as rotas
├── internal/
│   ├── cloudflare/
│   │   └── cloudflare.go     # Integração com R2 e D1
│   ├── handlers/
│   │   └── complaint_handler.go  # Handler do endpoint POST /complaint
│   └── models/
│       └── models.go         # Struct Complaint
├── schema.sql                # Migration para criar a tabela no D1
├── .env.example              # Exemplo de variáveis de ambiente
├── go.mod
└── go.sum
```



## Pré-requisitos

- Go 1.21+
- Conta no Cloudflare com:
  - Um bucket **R2** criado
  - Um banco **D1** criado
  - Um **API Token** com permissões `D1:Edit` e `Workers R2 Storage:Edit`



## Configuração

### 1. Clone o repositório

```bash
git clone https://github.com/pedropaffaro/deep-pothole-backend
cd backend
```

### 2. Instale as dependências

```bash
go mod tidy
```

### 3. Configure as variáveis de ambiente

Copie o arquivo de exemplo e preencha com seus valores:

```bash
cp .env.example .env
```

| Variável | Descrição |
|----| ----|
| `CF_ACCOUNT_ID` | ID da conta Cloudflare |
| `CF_API_TOKEN` | Token de autenticação da API |
| `CF_R2_BUCKET` | Nome do bucket R2 |
| `CF_D1_DATABASE_ID` | ID do banco D1 |


**Permissões necessárias no API Token:**
- `Account` → `D1` → `Edit`
- `Account` → `Workers R2 Storage` → `Edit`

### 4. Crie a tabela no D1

Via painel do Cloudflare (Workers & Pages → D1 → seu banco → Console), execute:

```sql
CREATE TABLE IF NOT EXISTS complaints (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    city      TEXT    NOT NULL,
    street    TEXT    NOT NULL,
    latitude  REAL    NOT NULL,
    longitude REAL    NOT NULL,
    photo_url TEXT    NOT NULL
);
```

## Rodando o Servidor

```bash
make run
```

O servidor sobe em `http://localhost:8080`.



## Endpoints

### `POST /api/v1/complaint`

Recebe uma denúncia de buraco via `multipart/form-data`.

**Campos do formulário:**

| Campo | Tipo | Descrição |
|---|---|---|
| `cidade` | string | Nome da cidade |
| `rua` | string | Nome da rua |
| `latitude` | float | Latitude da localização |
| `longitude` | float | Longitude da localização |
| `foto` | file | Imagem do buraco (obrigatória) |


**Respostas de erro:**

| Status | Mensagem | Causa |
|---|---|---|
| `400` | Foto é obrigatória | Nenhum arquivo enviado |
| `500` | Erro ao enviar imagem do buraco | Falha do POST da imagem |
| `500` | Erro ao salvar registro de buraco | Falha no POST no SQL |


