## Update 02 - 23/08

* Instalação local e setup do [GORM](https://gorm.io/docs/)
- Preparo da geração do banco de dados, para setup do Docker.

- O comando `go mod tidy`, instala e atualiza dependências locais necessárias.
- o comando `go test ./...`, quando executado na raiz do projeto, percorre todas as pastas e roda todos os testes.

### Docker
- Implementação e atualização em
- Comando `docker run --rm hello-world` roda um teste de conteiner, sem precisar configurar nada.
- Comando `docker compose up --build` sobe o conteiner usando os parâmetros do compose. Mas para atualizar os parâmetros, é necessário reiniciar.
- Comando `docker compose down -v` encerra o conteiner.

### GORM
Por que cada import é necessário?

* `internal/config`: carrega/expõe AppConfig e o DSN do banco. Sem ele, o db.Open não sabe como conectar.

* `internal/db`: concentra a abertura do GORM, tuning do pool e migrações. Mantém o main limpo e evita duplicar setup.

* `internal/user`: contém o domínio (model), o repositório (acesso a dados via GORM) e o serviço (regras/transactions). O main cria as instâncias e injeta onde precisa.

* `internal/http`: camada de entrega. Recebe apenas o serviço (interface/struct) — não importa gorm. Facilita testes e troca de persistência.

#### Estrutura atualizada



## Update 01 - 23/08

* fix/Github Actions

* Add Go Tests 
- Comando `go mod init cmd/tests`, onde "cmd/tests" é a definição da pasta onde os testes serão buscados. Gerando o arquivo de módulos `go.mod`. O comando `go test`, deve ser executado na pasta de testes `cmd/tests`.

### Por que separar os testes?
- A classes que desejam ser testadas, devem estar na mesma pasta que o teste.
- A separação, ajuda no controle de classes nos repositórios `main`, `develop`, `tests` & `features`.
Assim, os commits podem ser submetidos, com as classes "bagunçadas", na pasta de testes e ajustadas para o merge em `develop`.


**Sprint 0 – Setup de Time, Stack e Projeto**

# User Service – Go + Gin + PostgreSQL

> Serviço base para o domínio **User**, com especificação **OpenAPI**, infraestrutura Docker, migração SQL e CI simples em GitHub Actions.