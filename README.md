# Hub Dokploy 🚀

Uma aplicação extremamente leve desenvolvida em Go (Golang) que serve como **HUB central de atalhos** para todas as suas aplicações hospedadas no **Dokploy**.

Ela varre dinamicamente os arquivos de configuração do Traefik em `/etc/dokploy/traefik/dynamic/` e gera automaticamente um painel moderno e responsivo com links diretos para cada aplicação com domínio configurado.

---

## ✨ Funcionalidades

- ⚡ **Ultra Leve**: Consome menos de 10 MB de memória RAM e binary estático compilado em Go.
- 🎨 **Interface Moderna (Dark Mode)**: Painel elegante com efeito Glassmorphism, busca em tempo real e badges de protocolo (HTTPS/HTTP).
- 🔍 **Filtro em Tempo Real**: Campo de busca instantâneo para filtrar aplicações por nome ou domínio.
- 📋 **Copiar URLs com 1 Clique**: Botões práticos para copiar o link da aplicação diretamente para a área de transferência.
- 🐋 **Containerizada**: Multi-stage Dockerfile pronto para deploy direto no Dokploy ou Docker Swarm.
- 🔄 **Atualização Dinâmica**: Lê as configurações do Traefik direto da pasta montada, sem necessidade de banco de dados.

---

## 🛠️ Como Funciona

O Dokploy armazena as rotas dinâmicas do Traefik no diretório host `/etc/dokploy/traefik/dynamic/`. O **Hub Dokploy** lê esses arquivos (`.yml`, `.yaml`, `.json`), extrai as regras de domínio do Traefik (`Host(...)`), mapeia o protocolo (HTTPS/HTTP) e gera a listagem automaticamente.

---

## 🚀 Como Fazer Deploy no Dokploy

### Opção 1: Via Dockerfile (Recomendado no Dokploy)

1. Crie uma nova aplicação no Dokploy apontando para o seu repositório Git.
2. Selecione o **Provider: Dockerfile**.
3. Em **Volumes**, adicione o seguinte mapeamento de volume host:
   - **Host Path**: `/etc/dokploy/traefik/dynamic`
   - **Container Path**: `/etc/dokploy/traefik/dynamic`
   - **Mode**: `ro` (Read Only)
4. Clique em **Deploy**.

---

### Opção 2: Via Docker Compose

Você também pode utilizar o arquivo `docker-compose.yml` incluído:

```yaml
version: '3.8'

services:
  hub-dokploy:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: hub-dokploy
    restart: always
    ports:
      - "8080:8080"
    volumes:
      - /etc/dokploy/traefik/dynamic:/etc/dokploy/traefik/dynamic:ro
    environment:
      - PORT=8080
      - DYNAMIC_CONFIG_DIR=/etc/dokploy/traefik/dynamic
```

---

## ⚙️ Variáveis de Ambiente

| Variável | Padrão | Descrição |
| :--- | :--- | :--- |
| `PORT` | `8080` | Porta onde o servidor HTTP irá escutar |
| `DYNAMIC_CONFIG_DIR` | `/etc/dokploy/traefik/dynamic` | Caminho do diretório de configurações dinâmicas do Traefik |

---

## 💻 Desenvolvimento Local

```bash
# Instalar dependências
go mod download

# Executar testes unitários
go test -v ./...

# Compilar e rodar a aplicação localmente (usando diretório de teste ou padrão)
DYNAMIC_CONFIG_DIR=./test_config go run .
```
