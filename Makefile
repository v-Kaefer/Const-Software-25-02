.PHONY: help localstack-start localstack-stop localstack-status localstack-logs localstack-clean infra-up infra-down infra-test cognito-local-start cognito-local-stop cognito-local-setup cognito-local-test cognito-local-clean tflocal-init tflocal-plan tflocal-apply tflocal-destroy infra-prod-init infra-prod-plan infra-prod-apply infra-prod-destroy

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
	@echo "  make cognito-local-start - Inicia cognito-local"
	@echo "  make cognito-local-setup - Configura cognito-local com Terraform"
	@echo "  make cognito-local-test  - Testa configuração do cognito-local"
	@echo "  make cognito-local-stop  - Para cognito-local"
	@echo "  make cognito-local-clean - Remove cognito-local e dados"
	@echo ""
	@echo "Comandos Terraform Local (infra com tflocal para testes):"
	@echo "  make tflocal-init        - Inicializa o Terraform Local"
	@echo "  make tflocal-plan        - Executa tflocal plan"
	@echo "  make tflocal-apply       - Aplica a infraestrutura com tflocal"
	@echo "  make tflocal-destroy     - Destrói a infraestrutura com tflocal"
	@echo ""
	@echo "Comandos Terraform Produção (infra):"
	@echo "  make infra-prod-init     - Inicializa o Terraform (produção)"
	@echo "  make infra-prod-plan     - Executa terraform plan (produção)"
	@echo "  make infra-prod-apply    - Aplica a infraestrutura (produção)"
	@echo "  make infra-prod-destroy  - Destrói a infraestrutura (produção)"
	@echo ""
	@echo "Comandos combinados:"
	@echo "  make infra-up           - Inicia LocalStack + Terraform apply"
	@echo "  make infra-down         - Terraform destroy + Para LocalStack"
	@echo "  make infra-test         - Testa a infraestrutura criada"
	@echo ""
	@echo "==================================================================="
	@echo "IMPORTANTE: Cognito requer LocalStack Pro!"
	@echo "==================================================================="
	@echo "O LocalStack free tier NÃO suporta Cognito."
	@echo ""
	@echo "✅ SOLUÇÃO IMPLEMENTADA: cognito-local"
	@echo "Para testar Cognito GRATUITAMENTE com cognito-local:"
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
infra-up: localstack-start tflocal-init tflocal-apply
	@echo "✅ Infraestrutura completa iniciada!"
	@echo ""
	@echo "📊 Recursos disponíveis:"
	@echo "  - S3: http://localhost:4566"
	@echo "  - DynamoDB: http://localhost:4566"
	@echo "  - Cognito: http://localhost:4566 (requer LocalStack Pro)"
	@echo ""
	@echo "Para testar os recursos:"
	@echo "  make infra-test"

infra-down: tflocal-destroy localstack-stop
	@echo "✅ Infraestrutura completa parada!"

infra-test:
	@echo "🧪 Testando infraestrutura LocalStack..."
	@echo ""
	@echo "1️⃣ Testando S3..."
	@aws --endpoint-url=http://localhost:4566 s3 ls || echo "❌ S3 não disponível"
	@echo ""
	@echo "2️⃣ Testando DynamoDB..."
	@aws --endpoint-url=http://localhost:4566 dynamodb list-tables || echo "❌ DynamoDB não disponível"
	@echo ""
	@echo "3️⃣ Testando Cognito (requer LocalStack Pro)..."
	@aws --endpoint-url=http://localhost:4566 cognito-idp list-user-pools --max-results 10 || echo "❌ Cognito não disponível no free tier"
	@echo ""
	@echo "✅ Teste concluído!"

# cognito-local commands
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

cognito-local-clean:
	@echo "🧹 Limpando cognito-local..."
	@docker-compose -f docker-compose.cognito-local.yaml down -v
	@rm -rf infra/cognito-local-config/*.json
	@echo "✅ Limpeza concluída!"

# Terraform Local (tflocal) commands for local testing with infra directory
# EC2 is supported in LocalStack free tier
tflocal-init:
	@echo "🔧 Inicializando Terraform Local..."
	@cd infra && tflocal init
	@echo "✅ Terraform Local inicializado!"

tflocal-plan:
	@echo "📋 Executando tflocal plan..."
	@cd infra && tflocal plan

tflocal-apply:
	@echo "🚀 Aplicando infraestrutura com tflocal..."
	@cd infra && tflocal apply -auto-approve
	@echo "✅ Infraestrutura aplicada!"

tflocal-destroy:
	@echo "💣 Destruindo infraestrutura com tflocal..."
	@cd infra && tflocal destroy -auto-approve
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

infra-prod-destroy:
	@echo "💣 Destruindo infraestrutura de produção..."
	@cd infra && terraform destroy -auto-approve
	@echo "✅ Infraestrutura destruída!"
