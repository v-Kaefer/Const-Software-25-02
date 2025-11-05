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
	@echo "  make infra-up           - Inicia LocalStack + cognito-local + tflocal"
	@echo "  make infra-down         - Para tudo (tflocal + cognito-local + LocalStack)"
	@echo "  make infra-test         - Testa a infraestrutura criada"
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
infra-up: localstack-start cognito-local-start tflocal-init cognito-local-setup tflocal-apply
	@echo "✅ Infraestrutura completa iniciada!"
	@echo ""
	@echo "📊 Recursos disponíveis:"
	@echo "  - S3: http://localhost:4566"
	@echo "  - DynamoDB: http://localhost:4566"
	@echo "  - Cognito: http://localhost:9229 (cognito-local)"
	@echo ""
	@echo "Para testar os recursos:"
	@echo "  make infra-test"

infra-down: tflocal-destroy cognito-local-stop localstack-stop
	@echo "✅ Infraestrutura completa parada!"

infra-test:
	@echo "🧪 Testando infraestrutura LocalStack + cognito-local..."
	@echo ""
	@echo "1️⃣ Testando S3..."
	@aws --endpoint-url=http://localhost:4566 s3 ls s3://grupo-l-terraform 2>/dev/null && echo "✅ Bucket S3 'grupo-l-terraform' existe" || echo "❌ Bucket S3 não encontrado"
	@echo ""
	@echo "2️⃣ Testando DynamoDB..."
	@aws --endpoint-url=http://localhost:4566 dynamodb describe-table --table-name GrupoLConstSoftSprint1DynamoDB 2>/dev/null | grep -q "TableName" && echo "✅ Tabela DynamoDB 'GrupoLConstSoftSprint1DynamoDB' existe" || echo "❌ Tabela DynamoDB não encontrada"
	@echo ""
	@echo "3️⃣ Testando IAM Roles..."
	@aws --endpoint-url=http://localhost:4566 iam get-role --role-name ec2_role 2>/dev/null | grep -q "ec2_role" && echo "✅ IAM Role 'ec2_role' existe" || echo "❌ IAM Role não encontrada"
	@echo ""
	@echo "4️⃣ Testando VPC Security Groups..."
	@aws --endpoint-url=http://localhost:4566 ec2 describe-security-groups --filters "Name=group-name,Values=allow-http" 2>/dev/null | grep -q "allow-http" && echo "✅ Security Group 'allow-http' existe" || echo "❌ Security Group não encontrado"
	@echo ""
	@echo "5️⃣ Testando EC2 Key Pair..."
	@aws --endpoint-url=http://localhost:4566 ec2 describe-key-pairs --key-names grupo-l-key 2>/dev/null | grep -q "grupo-l-key" && echo "✅ Key Pair 'grupo-l-key' existe" || echo "❌ Key Pair não encontrado"
	@echo ""
	@echo "6️⃣ Testando EC2 Instance..."
	@aws --endpoint-url=http://localhost:4566 ec2 describe-instances --filters "Name=tag:Name,Values=grupo-l-sprint1" 2>/dev/null | grep -q "grupo-l-sprint1" && echo "✅ EC2 Instance 'grupo-l-sprint1' existe" || echo "❌ EC2 Instance não encontrada"
	@echo ""
	@echo "7️⃣ Testando Cognito (cognito-local)..."
	@aws --endpoint-url=http://localhost:9229 cognito-idp list-user-pools --max-results 10 2>/dev/null | grep -q "CognitoUserPool" && echo "✅ Cognito User Pool existe (cognito-local)" || echo "❌ Cognito não disponível"
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

infra-prod-destroy:
	@echo "💣 Destruindo infraestrutura de produção..."
	@cd infra && terraform destroy -auto-approve
	@echo "✅ Infraestrutura destruída!"
