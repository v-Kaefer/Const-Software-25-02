.PHONY: help localstack-start localstack-stop localstack-status localstack-logs localstack-clean infra-up infra-down infra-test infra-debug cognito-local-start cognito-local-stop cognito-local-setup cognito-local-test cognito-local-clean cognito-local-ready cognito-local-passwords infra-prod-passwords tflocal-init tflocal-plan tflocal-apply tflocal-destroy infra-prod-init infra-prod-plan infra-prod-apply infra-prod-destroy docker-compose-up docker-compose-down swagger-only build test go-test test-db-up test-db-down test-workspace test-http test-api

# Default target
help:
	@echo "==================================================================="
	@echo "Makefile para gerenciar LocalStack e Terraform"
	@echo "==================================================================="
	@echo ""
	@echo "Comandos LocalStack:"
	@echo "  make localstack-start    - Inicia o LocalStack"
	@echo "  make localstack-stop     - Para o LocalStack"
	@echo "  make localstack-status   - Verifica o status do LocalStack"
	@echo "  make localstack-logs     - Mostra os logs do LocalStack"
	@echo "  make localstack-clean    - Remove containers e volumes do LocalStack"
	@echo ""
	@echo "Comandos cognito-local (Alternativa Free ao Cognito):"
	@echo "  make cognito-local-start     - Inicia cognito-local"
	@echo "  make cognito-local-setup     - Configura cognito-local com Terraform"
	@echo "  make cognito-local-test      - Testa configuração do cognito-local"
	@echo "  make cognito-local-passwords - Exibe senhas dos usuários (cognito-local)"
	@echo "  make cognito-local-stop      - Para cognito-local"
	@echo "  make cognito-local-clean     - Remove cognito-local e dados"
	@echo ""
	@echo "Comandos Docker Compose (API, Database e Swagger UI):"
	@echo "  make swagger-only        - Inicia APENAS o Swagger UI (mais rápido)"
	@echo "  make docker-compose-up   - RESET FORÇADO + inicia serviços (db, api, swagger)"
	@echo "  make docker-compose-down - RESET FORÇADO + para serviços e limpa volumes/imagens"
	@echo ""
	@echo "Comandos Terraform Local (infra com tflocal para testes):"
	@echo "  make tflocal-init        - Inicializa o Terraform Local"
	@echo "  make tflocal-plan        - Executa tflocal plan"
	@echo "  make tflocal-apply       - Aplica a infraestrutura com tflocal"
	@echo "  make tflocal-destroy     - Destrói a infraestrutura com tflocal"
	@echo ""
	@echo "Comandos Terraform Produção (infra):"
	@echo "  make infra-prod-init      - Inicializa o Terraform (produção)"
	@echo "  make infra-prod-plan      - Executa terraform plan (produção)"
	@echo "  make infra-prod-apply     - Aplica a infraestrutura (produção)"
	@echo "  make infra-prod-passwords - Exibe senhas geradas pelo Terraform (produção)"
	@echo "  make infra-prod-destroy   - Destrói a infraestrutura (produção)"
	@echo ""
	@echo "Comandos combinados:"
	@echo "  make infra-up           - RESET FORÇADO + LocalStack + cognito-local + API + Swagger"
	@echo "  make infra-down         - RESET FORÇADO + para tudo e limpa volumes/imagens"
	@echo "  make infra-test         - Testa a infraestrutura criada"
	@echo "  make infra-debug        - Debug da infraestrutura (lista todos os recursos)"
	@echo ""
	@echo "Comandos de build/teste da API:"
	@echo "  make build              - Compila ./cmd/api dentro do container local"
	@echo "  make test               - Sobe dependências necessárias e executa go test ./..."
	@echo "  make test-api           - Testa a API com requisições HTTP (simula Swagger UI)"
	@echo ""
	@echo "==================================================================="
	@echo "IMPORTANTE: Cognito - Integrado automaticamente!"
	@echo "==================================================================="
	@echo "O LocalStack free tier NÃO suporta Cognito."
	@echo ""
	@echo "✅ SOLUÇÃO IMPLEMENTADA: cognito-local integrado no pipeline"
	@echo "O comando 'make infra-up' já inicia cognito-local automaticamente!"
	@echo "tflocal exclui recursos Cognito e usa cognito-local no lugar."
	@echo ""
	@echo "Para testar Cognito manualmente:"
	@echo "  1. make cognito-local-start  # Inicia o emulador"
	@echo "  2. make cognito-local-setup  # Configura igual ao Terraform"
	@echo "  3. make cognito-local-test   # Testa a configuração"
	@echo ""
	@echo "Para testar sem Cognito (apenas S3 e DynamoDB):"
	@echo "  - Comente os recursos Cognito no cognito.tf temporariamente"
	@echo "  - Execute: make infra-up"
	@echo "==================================================================="

# LocalStack commands
localstack-start:
	@echo "🚀 Iniciando LocalStack..."
	@localstack start -d
	@echo "⏳ Aguardando LocalStack ficar pronto..."
	@sleep 10
	@localstack status
	@echo "✅ LocalStack iniciado!"

localstack-stop:
	@echo "🛑 Parando LocalStack..."
	@localstack stop
	@echo "✅ LocalStack parado!"

localstack-status:
	@echo "📊 Status do LocalStack:"
	@localstack status || echo "❌ LocalStack não está rodando"

localstack-logs:
	@echo "📋 Logs do LocalStack:"
	@localstack logs

localstack-clean:
	@echo "🧹 Limpando containers e volumes do LocalStack..."
	@docker ps -a | grep localstack | awk '{print $$1}' | xargs -r docker rm -f
	@docker volume ls | grep localstack | awk '{print $$2}' | xargs -r docker volume rm
	@echo "✅ Limpeza concluída!"

# Combined commands
infra-up: localstack-start cognito-local-start tflocal-init cognito-local-setup tflocal-apply docker-compose-up
	@echo "✅ Infraestrutura completa iniciada (com reset forçado)!"
	@echo ""
	@echo "📊 Recursos disponíveis:"
	@echo "  - S3: http://localhost:4566"
	@echo "  - DynamoDB: http://localhost:4566"
	@echo "  - Cognito: http://localhost:9229 (cognito-local)"
	@echo "  - API: http://localhost:8080"
	@echo "  - Swagger UI: http://localhost:8081"
	@echo "  - Health check: http://localhost:8080/api/v1/health"
	@echo ""
	@echo "Para testar os recursos:"
	@echo "  make infra-test"

infra-down: tflocal-destroy cognito-local-clean localstack-stop docker-compose-down
	@echo "✅ Infraestrutura completa parada (volumes e imagens limpos)!"

infra-test: cognito-local-ready
	@echo "🧪 Testando infraestrutura LocalStack + cognito-local..."
	@echo ""
	@echo "1️⃣ Testando S3..."
	@AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test aws --endpoint-url=http://localhost:4566 --region us-east-1 s3 ls s3://grupo-l-terraform >/dev/null 2>&1 && echo "✅ Bucket S3 'grupo-l-terraform' existe" || echo "❌ Bucket S3 não encontrado"
	@echo ""
	@echo "2️⃣ Testando DynamoDB..."
	@AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test aws --endpoint-url=http://localhost:4566 --region us-east-1 dynamodb describe-table --table-name GrupoLConstSoftSprint1DynamoDB >/dev/null 2>&1 && echo "✅ Tabela DynamoDB 'GrupoLConstSoftSprint1DynamoDB' existe" || echo "❌ Tabela DynamoDB não encontrada"
	@echo ""
	@echo "3️⃣ Testando IAM Roles..."
	@AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test aws --endpoint-url=http://localhost:4566 --region us-east-1 iam get-role --role-name ec2_role >/dev/null 2>&1 && echo "✅ IAM Role 'ec2_role' existe" || echo "❌ IAM Role não encontrada"
	@echo ""
	@echo "4️⃣ Testando VPC Security Groups..."
	@AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test aws --endpoint-url=http://localhost:4566 --region us-east-1 ec2 describe-security-groups --group-names allow-http >/dev/null 2>&1 && echo "✅ Security Group 'allow-http' existe" || echo "❌ Security Group não encontrado"
	@echo ""
	@echo "5️⃣ Testando EC2 Key Pair..."
	@AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test aws --endpoint-url=http://localhost:4566 --region us-east-1 ec2 describe-key-pairs --key-names grupo-l-key >/dev/null 2>&1 && echo "✅ Key Pair 'grupo-l-key' existe" || echo "❌ Key Pair não encontrado"
	@echo ""
	@echo "6️⃣ Testando EC2 Instance..."
	@AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test aws --endpoint-url=http://localhost:4566 --region us-east-1 ec2 describe-instances --filters "Name=tag:Name,Values=grupo-l-sprint1" 2>&1 | grep -q "Instances" && echo "✅ EC2 Instance 'grupo-l-sprint1' existe" || echo "❌ EC2 Instance não encontrada"
	@echo ""
	@echo "7️⃣ Testando Cognito (cognito-local)..."
	@AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test aws --endpoint-url=http://localhost:9229 --region us-east-1 cognito-idp list-user-pools --max-results 10 >/dev/null 2>&1 && echo "✅ Cognito User Pool disponível (cognito-local)" || echo "❌ Cognito não disponível"
	@echo ""
	@echo "8️⃣ Testando configuração detalhada do cognito-local..."
	@cd infra && ./test-cognito-local.sh
	@echo ""
	@echo "✅ Teste concluído!"
	@echo ""
	@echo "💡 Resumo dos recursos testados:"
	@echo "   - S3 Bucket (LocalStack)"
	@echo "   - DynamoDB Table (LocalStack)"
	@echo "   - IAM Roles (LocalStack)"
	@echo "   - VPC Security Groups (LocalStack)"
	@echo "   - EC2 Key Pair (LocalStack)"
	@echo "   - EC2 Instance (LocalStack)"
	@echo "   - Cognito User Pool (cognito-local)"
	@echo ""
	@echo "✅ Teste concluído!"

infra-debug:
	@echo "🔍 Debugando infraestrutura..."
	@echo ""
	@echo "📊 LocalStack Status:"
	@localstack status 2>&1 || echo "LocalStack não está rodando"
	@echo ""
	@echo "📊 Cognito-local Status:"
	@docker ps | grep cognito-local || echo "cognito-local não está rodando"
	@echo ""
	@echo "📦 Listando todos os recursos S3:"
	@AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test aws --endpoint-url=http://localhost:4566 --region us-east-1 s3 ls 2>&1 || echo "Erro ao listar S3"
	@echo ""
	@echo "📦 Listando todas as tabelas DynamoDB:"
	@AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test aws --endpoint-url=http://localhost:4566 --region us-east-1 dynamodb list-tables 2>&1 || echo "Erro ao listar DynamoDB"
	@echo ""
	@echo "📦 Listando todos os IAM roles:"
	@AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test aws --endpoint-url=http://localhost:4566 --region us-east-1 iam list-roles 2>&1 | head -20 || echo "Erro ao listar IAM"
	@echo ""
	@echo "📦 Listando todos os security groups:"
	@AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test aws --endpoint-url=http://localhost:4566 --region us-east-1 ec2 describe-security-groups 2>&1 | head -20 || echo "Erro ao listar Security Groups"
	@echo ""
	@echo "📦 Listando todos os key pairs:"
	@AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test aws --endpoint-url=http://localhost:4566 --region us-east-1 ec2 describe-key-pairs 2>&1 || echo "Erro ao listar Key Pairs"
	@echo ""
	@echo "📦 Listando todas as instâncias EC2:"
	@AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test aws --endpoint-url=http://localhost:4566 --region us-east-1 ec2 describe-instances 2>&1 | head -20 || echo "Erro ao listar EC2"

# cognito-local commands
cognito-local-ready:
	@if docker ps --format '{{.Names}}' | grep -q "^cognito-local$$"; then \
		echo "✅ cognito-local já está em execução"; \
	else \
		echo "⚙️  cognito-local não está rodando. Iniciando agora..."; \
		$(MAKE) --no-print-directory cognito-local-start; \
	fi
	@if [ ! -f infra/cognito-local-config/config.json ]; then \
		echo "⚙️  Configuração do cognito-local não encontrada. Executando setup..."; \
		$(MAKE) --no-print-directory cognito-local-setup; \
	else \
		echo "✅ Configuração do cognito-local encontrada (infra/cognito-local-config/config.json)"; \
	fi

cognito-local-start:
	@echo "🚀 Iniciando cognito-local..."
	@docker-compose -f docker-compose.cognito-local.yaml up -d
	@echo "⏳ Aguardando cognito-local ficar pronto..."
	@sleep 10
	@echo "🔍 Verificando status do container..."
	@docker ps | grep cognito-local || (echo "❌ Container não está rodando" && docker logs cognito-local && exit 1)
	@echo "✅ cognito-local iniciado em http://localhost:9229"
	@echo ""
	@echo "💡 Próximo passo: make cognito-local-setup"

cognito-local-stop:
	@echo "🛑 Parando cognito-local..."
	@docker-compose -f docker-compose.cognito-local.yaml down
	@echo "✅ cognito-local parado!"

cognito-local-setup:
	@echo "🔧 Configurando cognito-local com base no Terraform..."
	@cd infra && ./setup-cognito-local.sh
	@echo "✅ Configuração concluída!"

cognito-local-test:
	@echo "🧪 Testando configuração do cognito-local..."
	@cd infra && ./test-cognito-local.sh

cognito-local-passwords:
	@echo "🔑 Senhas dos usuários (cognito-local - ambiente de desenvolvimento):"
	@echo ""
	@echo "👤 Usuários e senhas padrão:"
	@echo "   admin@example.com    -> AdminTemp123!"
	@echo "   reviewer@example.com -> PassTemp123!"
	@echo "   user@example.com     -> PassTemp123!"
	@echo ""
	@echo "⚠️  Se você usou senhas customizadas no cognito-local-setup, consulte o valor definido."
	@echo ""
	@echo "💡 Para personalizar senhas, defina variáveis de ambiente antes de cognito-local-setup:"
	@echo "   ADMIN_PASSWORD=SuaSenha123! REVIEWER_PASSWORD=OutraSenha123! USER_PASSWORD=Senha123! make cognito-local-setup"
	@echo ""

cognito-local-clean:
	@echo "🧹 Limpando cognito-local..."
	@docker-compose -f docker-compose.cognito-local.yaml down -v
	@rm -rf infra/cognito-local-config/*.json
	@echo "✅ Limpeza concluída!"

# Docker Compose commands for API, Database and Swagger UI
docker-compose-up:
	@echo "🚀 Iniciando serviços com Docker Compose..."
	@echo "🧹 Limpando containers e volumes existentes..."
	@docker compose down --remove-orphans --volumes 2>/dev/null || true
	@docker rm -f swagger userdb usersvc 2>/dev/null || true
	@echo "🗑️ Removendo imagens antigas da API para forçar rebuild..."
	@docker rmi -f $$(docker images -q 'const-software*' 2>/dev/null) 2>/dev/null || true
	@docker rmi -f $$(docker images -q '*usersvc*' 2>/dev/null) 2>/dev/null || true
	@sleep 1
	@echo "🔨 Reconstruindo imagens com código mais recente..."
	@docker compose up -d --build --remove-orphans --force-recreate
	@echo "⏳ Aguardando serviços ficarem prontos..."
	@sleep 5
	@echo "✅ Serviços iniciados com código atualizado!"
	@echo "  - Database: http://localhost:5432"
	@echo "  - API: http://localhost:8080"
	@echo "  - Swagger UI: http://localhost:8081"
	@echo "  - Health check: http://localhost:8080/api/v1/health"

docker-compose-down:
	@echo "🛑 Parando serviços do Docker Compose..."
	@docker compose down --remove-orphans --volumes
	@docker rm -f swagger userdb usersvc 2>/dev/null || true
	@echo "🧹 Removendo imagens antigas da API..."
	@docker rmi -f $$(docker images -q 'const-software*' 2>/dev/null) 2>/dev/null || true
	@docker rmi -f $$(docker images -q '*usersvc*' 2>/dev/null) 2>/dev/null || true
	@echo "✅ Serviços parados e volumes/imagens limpos!"

# Comando simplificado para apenas visualizar o Swagger (sem API)
swagger-only:
	@echo "🚀 Iniciando apenas o Swagger UI..."
	@echo "🧹 Limpando containers existentes..."
	@docker compose down --remove-orphans 2>/dev/null || true
	@docker rm -f swagger userdb usersvc 2>/dev/null || true
	@sleep 1
	@docker compose up -d --remove-orphans swagger
	@echo "⏳ Aguardando Swagger ficar pronto..."
	@sleep 3
	@echo "✅ Swagger UI iniciado!"
	@echo "  - Swagger UI: http://localhost:8081"
	@echo ""
	@echo "💡 Para visualizar a página do Swagger, acesse: http://localhost:8081"

build:
	@echo "🔨 Compilando aplicação Go..."
	@go build -o cmd/api/usersvc ./cmd/api

# Go test workflow
GO_TEST_CACHE ?= $(CURDIR)/.cache
GO_MOD_CACHE ?= $(CURDIR)/.gomodcache
GO_TEST_FLAGS ?=
GO_TEST_TARGETS ?= ./...
TEST_DB_SENTINEL ?= $(CURDIR)/.tmp/.db-started-for-test

test: go-test

go-test: test-db-up
	@set -euo pipefail; \
	  trap '$(MAKE) --no-print-directory test-db-down' EXIT; \
	  echo "🧪 Executando testes Go com dependências locais..."; \
	  GOCACHE="$(GO_TEST_CACHE)" GOMODCACHE="$(GO_MOD_CACHE)" go test $(GO_TEST_FLAGS) $(GO_TEST_TARGETS)

test-workspace:
	@$(MAKE) --no-print-directory GO_TEST_TARGETS=./pkg/workspace test

test-http:
	@$(MAKE) --no-print-directory GO_TEST_TARGETS=./internal/http test

# Teste de API simulando requisições do Swagger UI (inclui CORS preflight)
test-api:
	@echo "🧪 Testando API com requisições HTTP (simula Swagger UI)..."
	@echo ""
	@echo "1️⃣ Verificando se a API está rodando..."
	@if ! curl -s http://localhost:8080/api/v1/health > /dev/null 2>&1; then \
		echo "❌ API não está rodando. Execute 'make docker-compose-up' primeiro."; \
		exit 1; \
	fi
	@echo "✅ API está rodando!"
	@echo ""
	@echo "2️⃣ Testando endpoint /api/v1/health..."
	@HEALTH_RESPONSE=$$(curl -s http://localhost:8080/api/v1/health); \
	echo "   Resposta: $$HEALTH_RESPONSE"; \
	if echo "$$HEALTH_RESPONSE" | grep -q '"status":"ok"'; then \
		echo "   ✅ Health check OK"; \
	else \
		echo "   ❌ Health check falhou"; \
		exit 1; \
	fi
	@echo ""
	@echo "3️⃣ Testando CORS preflight (OPTIONS) para /api/v1/users..."
	@OPTIONS_RESPONSE=$$(curl -s -I -X OPTIONS http://localhost:8080/api/v1/users \
		-H "Origin: http://localhost:8081" \
		-H "Access-Control-Request-Method: POST" \
		-H "Access-Control-Request-Headers: Content-Type,Authorization"); \
	echo "   $$OPTIONS_RESPONSE" | head -5; \
	if echo "$$OPTIONS_RESPONSE" | grep -qi "Access-Control-Allow-Origin"; then \
		echo "   ✅ CORS headers presentes"; \
	else \
		echo "   ❌ CORS headers ausentes"; \
		exit 1; \
	fi
	@echo ""
	@echo "4️⃣ Testando POST /api/v1/users (criar usuário)..."
	@POST_RESPONSE=$$(curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/v1/users \
		-H "Content-Type: application/json" \
		-H "Origin: http://localhost:8081" \
		-d '{"email":"test-'$$(date +%s)'@example.com","name":"Test User"}'); \
	HTTP_CODE=$$(echo "$$POST_RESPONSE" | tail -1); \
	BODY=$$(echo "$$POST_RESPONSE" | head -n -1); \
	echo "   HTTP Status: $$HTTP_CODE"; \
	echo "   Body: $$BODY"; \
	if [ "$$HTTP_CODE" = "201" ]; then \
		echo "   ✅ Usuário criado com sucesso"; \
	else \
		echo "   ❌ Falha ao criar usuário (esperado 201, recebido $$HTTP_CODE)"; \
		exit 1; \
	fi
	@echo ""
	@echo "5️⃣ Testando GET /api/v1/users (listar usuários)..."
	@GET_RESPONSE=$$(curl -s -w "\n%{http_code}" -X GET http://localhost:8080/api/v1/users \
		-H "Origin: http://localhost:8081"); \
	HTTP_CODE=$$(echo "$$GET_RESPONSE" | tail -1); \
	BODY=$$(echo "$$GET_RESPONSE" | head -n -1); \
	echo "   HTTP Status: $$HTTP_CODE"; \
	echo "   Body: $$BODY" | head -c 200; echo "..."; \
	if [ "$$HTTP_CODE" = "200" ]; then \
		echo "   ✅ Lista de usuários OK"; \
	else \
		echo "   ❌ Falha ao listar usuários (esperado 200, recebido $$HTTP_CODE)"; \
		exit 1; \
	fi
	@echo ""
	@echo "✅ Todos os testes de API passaram!"

test-db-up:
	@mkdir -p $(dir $(TEST_DB_SENTINEL))
	@DB_ID=$$(docker compose ps -q db 2>/dev/null || true); \
	if [ -n "$$DB_ID" ] && docker inspect -f '{{.State.Running}}' "$$DB_ID" 2>/dev/null | grep -q true; then \
		echo "🐘 Postgres já está em execução (container $$DB_ID)."; \
		rm -f "$(TEST_DB_SENTINEL)"; \
	else \
		echo "🐘 Iniciando Postgres para testes..."; \
		docker compose up -d db >/dev/null; \
		echo "started" > "$(TEST_DB_SENTINEL)"; \
	fi

test-db-down:
	@if [ -f "$(TEST_DB_SENTINEL)" ]; then \
		echo "🧹 Parando Postgres utilizado nos testes..."; \
		docker compose stop db >/dev/null 2>&1 || true; \
		rm -f "$(TEST_DB_SENTINEL)"; \
	else \
		echo "ℹ️  Mantendo Postgres rodando (não foi iniciado pelo make test)."; \
	fi

# Terraform Local (tflocal) commands for local testing with infra directory
# EC2 is supported in LocalStack free tier
# Cognito resources are excluded as cognito-local is used instead (free alternative)
tflocal-init:
	@echo "🔧 Inicializando Terraform Local..."
	@cd infra && mv cognito.tf cognito.tf.skip 2>/dev/null || true
	@cd infra && tflocal init
	@cd infra && mv cognito.tf.skip cognito.tf 2>/dev/null || true
	@echo "✅ Terraform Local inicializado!"

tflocal-plan:
	@echo "📋 Executando tflocal plan..."
	@cd infra && mv cognito.tf cognito.tf.skip 2>/dev/null || true
	@cd infra && tflocal plan -var="use_localstack=true"
	@cd infra && mv cognito.tf.skip cognito.tf 2>/dev/null || true

tflocal-apply:
	@echo "🚀 Aplicando infraestrutura com tflocal..."
	@cd infra && mv cognito.tf cognito.tf.skip 2>/dev/null || true
	@cd infra && tflocal apply -auto-approve -var="use_localstack=true"
	@cd infra && mv cognito.tf.skip cognito.tf 2>/dev/null || true
	@echo "✅ Infraestrutura aplicada!"

tflocal-destroy:
	@echo "💣 Destruindo infraestrutura com tflocal..."
	@cd infra && mv cognito.tf cognito.tf.skip 2>/dev/null || true
	@cd infra && tflocal destroy -auto-approve -var="use_localstack=true"
	@cd infra && mv cognito.tf.skip cognito.tf 2>/dev/null || true
	@echo "✅ Infraestrutura destruída!"

# Production Terraform commands for infra directory
infra-prod-init:
	@echo "🔧 Inicializando Terraform (produção)..."
	@cd infra && terraform init
	@echo "✅ Terraform inicializado!"

infra-prod-plan:
	@echo "📋 Executando terraform plan (produção)..."
	@cd infra && terraform plan

infra-prod-apply:
	@echo "🚀 Aplicando infraestrutura de produção..."
	@cd infra && terraform apply -auto-approve
	@echo "✅ Infraestrutura aplicada!"
	@echo ""
	@echo "💡 Para ver as senhas geradas, execute: make infra-prod-passwords"

infra-prod-passwords:
	@echo "🔑 Senhas geradas pelo Terraform (produção):"
	@echo ""
	@cd infra && terraform output -raw admin_temp_password 2>/dev/null && echo " -> Senha Admin" || echo "❌ Execute 'make infra-prod-apply' primeiro"
	@cd infra && terraform output -raw reviewer_temp_password 2>/dev/null && echo " -> Senha Reviewer" || true
	@cd infra && terraform output -raw user_temp_password 2>/dev/null && echo " -> Senha User" || true
	@echo ""
	@echo "⚠️  Estas são senhas temporárias. Os usuários devem alterá-las no primeiro login."

infra-prod-destroy:
	@echo "💣 Destruindo infraestrutura de produção..."
	@cd infra && terraform destroy -auto-approve
	@echo "✅ Infraestrutura destruída!"
