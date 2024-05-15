
# Jogo Connect Four RPC

Este repositório contém a implementação de um jogo Connect Four usando Chamadas de Procedimento Remoto (RPC) com gRPC em Go. O projeto é elaborado como uma submissão de trabalho para a disciplina de Sistemas Distribuídos, ministrada pelo Professor Giacomin na UESC, durante o semestre de 2024.1.

## Visão Geral

O jogo Connect Four permite que dois jogadores se conectem através de uma rede e joguem um contra o outro. A lógica do jogo é gerenciada pelo servidor, que também administra o estado do jogo e as sessões dos jogadores usando streams gRPC.

## Pré-requisitos

Para executar este projeto, você precisará de:
- Go (versão 1.15 ou superior)
- gRPC
- Protocol Buffers

## Instalação

Siga estas etapas para configurar o ambiente:

1. Clone o repositório:
   ```
   git clone https://github.com/danieljcksn/connect-four
   ```

2. Navegue até o diretório do projeto:
   ```
   cd connect-four
   ```

3. Instale os pacotes Go necessários:
   ```
   go mod tidy
   ```

## Executando o Projeto

### Iniciando o Servidor

1. Navegue até o diretório `server`:
   ```
   cd server
   ```

2. Construa e execute o servidor:
   ```
   go build
   ./server
   ```

### Iniciando o Cliente

1. Abra um novo terminal e navegue até o diretório `client`:
   ```
   cd client
   ```

2. Construa e execute o cliente:
   ```
   go build
   ./client
   ```
3. Em outro terminal, repita o passo 2 para iniciar o segundo cliente. 

4. Siga as instruções na tela para inserir seu apelido e começar a jogar.


### Compilando os arquivos proto (opcional)

Caso deseje compilar os arquivos proto manualmente, siga estas etapas:

1. Navegue até o diretório `proto`:
   ```
   cd proto
   ```

2. Compile os arquivos proto:
   ```
   protoc --go_out=. --go_opt=paths=source_relative \
   --go-grpc_out=. --go-grpc_opt=paths=source_relative \
   service.proto
   ```