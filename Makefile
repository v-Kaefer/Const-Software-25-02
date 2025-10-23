.PHONY: help localstack-start localstack-stop localstack-status localstack-logs terraform-init terraform-plan terraform-apply terraform-destroy localstack-clean infra-up infra-down infra-test cognito-local-start cognito-local-stop cognito-local-setup cognito-local-test cognito-local-clean

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
	@echo "Comandos Terraform (infra-localstack):"
	@echo "  make terraform-init      - Inicializa o Terraform"
	@echo "  make terraform-plan      - Executa terraform plan"
	@echo "  make terraform-apply     - Aplica a infraestrutura"
	@echo "  make terraform-destroy   - Destrói a infraestrutura"
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

# Terraform commands
terraform-init:
	@echo "🔧 Inicializando Terraform..."
	@cd infra-localstack && terraform init
	@echo "✅ Terraform inicializado!"

terraform-plan:
	@echo "📋 Executando terraform plan..."
	@cd infra-localstack && terraform plan

terraform-apply:
	@echo "🚀 Aplicando infraestrutura com Terraform..."
	@cd infra-localstack && terraform apply -auto-approve
	@echo "✅ Infraestrutura aplicada!"

terraform-destroy:
	@echo "💣 Destruindo infraestrutura..."
	@cd infra-localstack && terraform destroy -auto-approve
	@echo "✅ Infraestrutura destruída!"

# Combined commands
infra-up: localstack-start terraform-init terraform-apply
	@echo "✅ Infraestrutura completa iniciada!"
	@echo ""
	@echo "📊 Recursos disponíveis:"
	@echo "  - S3: http://localhost:4566"
	@echo "  - DynamoDB: http://localhost:4566"
	@echo "  - Cognito: http://localhost:4566 (requer LocalStack Pro)"
	@echo ""
	@echo "Para testar os recursos:"
	@echo "  make infra-test"

infra-down: terraform-destroy localstack-stop
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
	@cd infra-localstack && ./setup-cognito-local.sh
	@echo "✅ Configuração concluída!"

cognito-local-test:
	@echo "🧪 Testando configuração do cognito-local..."
	@cd infra-localstack && ./test-cognito-local.sh

cognito-local-clean:
	@echo "🧹 Limpando cognito-local..."
	@docker-compose -f docker-compose.cognito-local.yaml down -v
	@rm -rf infra-localstack/cognito-local-config/*.json
	@echo "✅ Limpeza concluída!"
