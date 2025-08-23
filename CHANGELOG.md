## Update 02 - 23/08

* Instalação local e setup do [GORM](https://gorm.io/docs/)
- Preparo da geração do banco de dados, para setup do Docker.


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