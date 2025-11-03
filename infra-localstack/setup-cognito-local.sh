#!/bin/bash

# Script to setup cognito-local with the same configuration as Terraform
# This allows testing Cognito infrastructure without LocalStack Pro

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

ENDPOINT="http://localhost:9229"
REGION="us-east-1"

# Set dummy AWS credentials for cognito-local (as per cognito-local documentation)
# cognito-local doesn't validate credentials but AWS CLI requires them
export AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID:-local}"
export AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY:-local}"
export AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-us-east-1}"

echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║     Setup cognito-local (Alternativa ao LocalStack Pro)   ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if AWS CLI is installed
if ! command -v aws &> /dev/null; then
    echo -e "${RED}❌ AWS CLI não encontrado. Instale com: pip install awscli${NC}"
    exit 1
fi

echo -e "${GREEN}✅ AWS CLI encontrado${NC}"
echo -e "${YELLOW}ℹ️  Usando credenciais dummy para cognito-local (AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID})${NC}"
echo ""

# Check if cognito-local is running
echo -e "${YELLOW}🔍 Verificando se cognito-local está rodando...${NC}"
if ! curl -s -f "$ENDPOINT/health" > /dev/null 2>&1; then
    echo -e "${RED}❌ cognito-local não está rodando!${NC}"
    echo -e "${YELLOW}   Inicie com: docker-compose -f docker-compose.cognito-local.yaml up -d${NC}"
    exit 1
fi
echo -e "${GREEN}✅ cognito-local está rodando${NC}"
echo ""

# Load variables from terraform.tfvars if it exists
if [ -f "terraform.tfvars" ]; then
    echo -e "${YELLOW}📂 Carregando variáveis de terraform.tfvars...${NC}"
    # This is a simplified parser - in production you might want to use hcl2json
else
    echo -e "${YELLOW}⚠️  terraform.tfvars não encontrado, usando valores padrão${NC}"
fi
echo ""

# Check for existing User Pools and clean them up
echo -e "${YELLOW}🔍 Verificando User Pools existentes...${NC}"
EXISTING_POOLS=$(aws cognito-idp list-user-pools \
    --max-results 60 \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" \
    --output json 2>/dev/null || echo '{"UserPools":[]}')

# Delete any existing pools with the same name to avoid conflicts
POOL_IDS=$(echo "$EXISTING_POOLS" | grep -o '"Id": "[^"]*"' | cut -d'"' -f4)
if [ ! -z "$POOL_IDS" ]; then
    echo -e "${YELLOW}⚠️  Encontrados User Pools existentes. Removendo para evitar conflitos...${NC}"
    while IFS= read -r pool_id; do
        if [ ! -z "$pool_id" ]; then
            aws cognito-idp delete-user-pool \
                --user-pool-id "$pool_id" \
                --endpoint-url "$ENDPOINT" \
                --region "$REGION" \
                2>/dev/null || true
            echo -e "${YELLOW}   Removido pool: ${pool_id}${NC}"
        fi
    done <<< "$POOL_IDS"
fi
echo -e "${GREEN}✅ Limpeza concluída${NC}"
echo ""

echo -e "${GREEN}🏗️  Criando User Pool...${NC}"

# Create User Pool with similar configuration to Terraform
USER_POOL_OUTPUT=$(aws cognito-idp create-user-pool \
    --pool-name "CognitoUserPool" \
    --policies "PasswordPolicy={MinimumLength=8,RequireUppercase=true,RequireLowercase=true,RequireNumbers=true,RequireSymbols=false}" \
    --username-configuration "CaseSensitive=false" \
    --auto-verified-attributes "email" \
    --schema \
        "Name=email,AttributeDataType=String,Required=true,Mutable=true" \
        "Name=name,AttributeDataType=String,Required=false,Mutable=true" \
        "Name=role,AttributeDataType=String,Required=false,Mutable=true,DeveloperOnlyAttribute=false" \
    --mfa-configuration "OPTIONAL" \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" \
    --output json 2>&1)

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Erro ao criar User Pool${NC}"
    echo "$USER_POOL_OUTPUT"
    echo ""
    echo -e "${YELLOW}💡 Dica: Se o erro é sobre pool já existente, execute:${NC}"
    echo -e "${YELLOW}   make cognito-local-clean${NC}"
    echo -e "${YELLOW}   make cognito-local-start${NC}"
    echo -e "${YELLOW}   make cognito-local-setup${NC}"
    exit 1
fi

USER_POOL_ID=$(echo "$USER_POOL_OUTPUT" | grep -o '"Id": "[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$USER_POOL_ID" ]; then
    echo -e "${RED}❌ Não foi possível extrair o User Pool ID${NC}"
    echo "$USER_POOL_OUTPUT"
    exit 1
fi
echo -e "${GREEN}✅ User Pool criado: ${USER_POOL_ID}${NC}"
echo ""

# Create App Client
echo -e "${GREEN}🔑 Criando App Client...${NC}"
CLIENT_OUTPUT=$(aws cognito-idp create-user-pool-client \
    --user-pool-id "$USER_POOL_ID" \
    --client-name "my-app-client" \
    --explicit-auth-flows "ALLOW_USER_PASSWORD_AUTH" "ALLOW_REFRESH_TOKEN_AUTH" \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" \
    --output json 2>&1)

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Erro ao criar App Client${NC}"
    echo "$CLIENT_OUTPUT"
    exit 1
fi

CLIENT_ID=$(echo "$CLIENT_OUTPUT" | grep -o '"ClientId": "[^"]*"' | head -1 | cut -d'"' -f4)
echo -e "${GREEN}✅ App Client criado: ${CLIENT_ID}${NC}"
echo ""

# Create User Groups
echo -e "${GREEN}👥 Criando grupos de usuários...${NC}"

# Admin Group
aws cognito-idp create-group \
    --user-pool-id "$USER_POOL_ID" \
    --group-name "admin-group" \
    --description "Admin group managed by cognito-local" \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" > /dev/null 2>&1
echo -e "${GREEN}✅ Grupo admin-group criado${NC}"

# Reviewer Group
aws cognito-idp create-group \
    --user-pool-id "$USER_POOL_ID" \
    --group-name "reviewers-group" \
    --description "Reviewers group managed by cognito-local" \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" > /dev/null 2>&1
echo -e "${GREEN}✅ Grupo reviewers-group criado${NC}"

# User Group
aws cognito-idp create-group \
    --user-pool-id "$USER_POOL_ID" \
    --group-name "user-group" \
    --description "User group managed by cognito-local" \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" > /dev/null 2>&1
echo -e "${GREEN}✅ Grupo user-group criado${NC}"
echo ""

# Create Users
echo -e "${GREEN}👤 Criando usuários de exemplo...${NC}"

# Admin User
aws cognito-idp admin-create-user \
    --user-pool-id "$USER_POOL_ID" \
    --username "admin@example.com" \
    --user-attributes "Name=email,Value=admin@example.com" "Name=name,Value=Admin User" "Name=custom:role,Value=admin" \
    --temporary-password "AdminTemp123!" \
    --message-action "SUPPRESS" \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" > /dev/null 2>&1
echo -e "${GREEN}✅ Usuário admin@example.com criado${NC}"

# Add admin to admin-group
aws cognito-idp admin-add-user-to-group \
    --user-pool-id "$USER_POOL_ID" \
    --username "admin@example.com" \
    --group-name "admin-group" \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" > /dev/null 2>&1

# Reviewer User
aws cognito-idp admin-create-user \
    --user-pool-id "$USER_POOL_ID" \
    --username "reviewer@example.com" \
    --user-attributes "Name=email,Value=reviewer@example.com" "Name=name,Value=Reviewer User" "Name=custom:role,Value=reviewer" \
    --temporary-password "PassTemp123!" \
    --message-action "SUPPRESS" \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" > /dev/null 2>&1
echo -e "${GREEN}✅ Usuário reviewer@example.com criado${NC}"

# Add reviewer to reviewers-group
aws cognito-idp admin-add-user-to-group \
    --user-pool-id "$USER_POOL_ID" \
    --username "reviewer@example.com" \
    --group-name "reviewers-group" \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" > /dev/null 2>&1

# Main User
aws cognito-idp admin-create-user \
    --user-pool-id "$USER_POOL_ID" \
    --username "user@example.com" \
    --user-attributes "Name=email,Value=user@example.com" "Name=name,Value=Main User" \
    --temporary-password "PassTemp123!" \
    --message-action "SUPPRESS" \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" > /dev/null 2>&1
echo -e "${GREEN}✅ Usuário user@example.com criado${NC}"

# Add main user to user-group
aws cognito-idp admin-add-user-to-group \
    --user-pool-id "$USER_POOL_ID" \
    --username "user@example.com" \
    --group-name "user-group" \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" > /dev/null 2>&1

echo ""
echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                  ✅ Setup Completo!                        ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}📋 Informações da Infraestrutura:${NC}"
echo -e "   User Pool ID: ${GREEN}${USER_POOL_ID}${NC}"
echo -e "   App Client ID: ${GREEN}${CLIENT_ID}${NC}"
echo -e "   Endpoint: ${GREEN}${ENDPOINT}${NC}"
echo ""
echo -e "${YELLOW}👥 Usuários criados:${NC}"
echo -e "   ${GREEN}admin@example.com${NC} (senha temporária: AdminTemp123!)"
echo -e "   ${GREEN}reviewer@example.com${NC} (senha temporária: PassTemp123!)"
echo -e "   ${GREEN}user@example.com${NC} (senha temporária: PassTemp123!)"
echo ""
echo -e "${YELLOW}📝 Grupos criados:${NC}"
echo -e "   ${GREEN}admin-group${NC}"
echo -e "   ${GREEN}reviewers-group${NC}"
echo -e "   ${GREEN}user-group${NC}"
echo ""
echo -e "${YELLOW}🧪 Para testar a configuração:${NC}"
echo -e "   ./test-cognito-local.sh"
echo ""
echo -e "${YELLOW}💡 Para usar na aplicação Go:${NC}"
echo -e "   Configure o endpoint: ${GREEN}${ENDPOINT}${NC}"
echo ""

# Save configuration to file
cat > cognito-local-config/config.json << EOF
{
  "userPoolId": "${USER_POOL_ID}",
  "clientId": "${CLIENT_ID}",
  "endpoint": "${ENDPOINT}",
  "region": "${REGION}"
}
EOF

echo -e "${GREEN}✅ Configuração salva em: cognito-local-config/config.json${NC}"
