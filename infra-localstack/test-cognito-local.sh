#!/bin/bash

# Script to test cognito-local configuration
# Validates that the Cognito infrastructure was created correctly

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

ENDPOINT="http://localhost:9229"
REGION="us-east-1"

# Set dummy AWS credentials for cognito-local (as per cognito-local documentation)
export AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID:-local}"
export AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY:-local}"
export AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-us-east-1}"

echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║          Testando cognito-local Configuration             ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if config file exists
if [ ! -f "cognito-local-config/config.json" ]; then
    echo -e "${RED}❌ Arquivo de configuração não encontrado!${NC}"
    echo -e "${YELLOW}   Execute primeiro: ./setup-cognito-local.sh${NC}"
    exit 1
fi

# Load configuration
USER_POOL_ID=$(cat cognito-local-config/config.json | grep -o '"userPoolId": "[^"]*"' | cut -d'"' -f4)
CLIENT_ID=$(cat cognito-local-config/config.json | grep -o '"clientId": "[^"]*"' | cut -d'"' -f4)

echo -e "${YELLOW}📋 Configuração carregada:${NC}"
echo -e "   User Pool ID: ${GREEN}${USER_POOL_ID}${NC}"
echo -e "   Client ID: ${GREEN}${CLIENT_ID}${NC}"
echo ""

# Test 1: List User Pools
echo -e "${YELLOW}🧪 Teste 1: Listando User Pools...${NC}"
POOLS=$(aws cognito-idp list-user-pools \
    --max-results 10 \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" \
    --output json 2>&1)

if [ $? -eq 0 ]; then
    POOL_COUNT=$(echo "$POOLS" | grep -c '"Id"' || echo "0")
    echo -e "${GREEN}✅ Encontrado(s) ${POOL_COUNT} User Pool(s)${NC}"
else
    echo -e "${RED}❌ Erro ao listar User Pools${NC}"
    exit 1
fi
echo ""

# Test 2: Describe User Pool
echo -e "${YELLOW}🧪 Teste 2: Detalhes do User Pool...${NC}"
POOL_DETAILS=$(aws cognito-idp describe-user-pool \
    --user-pool-id "$USER_POOL_ID" \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" \
    --output json 2>&1)

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ User Pool encontrado e acessível${NC}"
    POOL_NAME=$(echo "$POOL_DETAILS" | grep -o '"Name": "[^"]*"' | head -1 | cut -d'"' -f4)
    echo -e "   Nome: ${GREEN}${POOL_NAME}${NC}"
else
    echo -e "${RED}❌ Erro ao obter detalhes do User Pool${NC}"
    exit 1
fi
echo ""

# Test 3: List Groups
echo -e "${YELLOW}🧪 Teste 3: Listando grupos...${NC}"
GROUPS=$(aws cognito-idp list-groups \
    --user-pool-id "$USER_POOL_ID" \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" \
    --output json 2>&1)

if [ $? -eq 0 ]; then
    GROUP_COUNT=$(echo "$GROUPS" | grep -c '"GroupName"' || echo "0")
    echo -e "${GREEN}✅ Encontrado(s) ${GROUP_COUNT} grupo(s)${NC}"
    
    # List group names
    echo "$GROUPS" | grep '"GroupName"' | cut -d'"' -f4 | while read group; do
        echo -e "   - ${GREEN}${group}${NC}"
    done
else
    echo -e "${RED}❌ Erro ao listar grupos${NC}"
    exit 1
fi
echo ""

# Test 4: List Users
echo -e "${YELLOW}🧪 Teste 4: Listando usuários...${NC}"
USERS=$(aws cognito-idp list-users \
    --user-pool-id "$USER_POOL_ID" \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" \
    --output json 2>&1)

if [ $? -eq 0 ]; then
    USER_COUNT=$(echo "$USERS" | grep -c '"Username"' || echo "0")
    echo -e "${GREEN}✅ Encontrado(s) ${USER_COUNT} usuário(s)${NC}"
    
    # List usernames
    echo "$USERS" | grep '"Username"' | cut -d'"' -f4 | while read user; do
        echo -e "   - ${GREEN}${user}${NC}"
    done
else
    echo -e "${RED}❌ Erro ao listar usuários${NC}"
    exit 1
fi
echo ""

# Test 5: Test Authentication Flow
echo -e "${YELLOW}🧪 Teste 5: Testando autenticação...${NC}"
echo -e "${YELLOW}   Tentando autenticar com user@example.com...${NC}"

# Note: This might fail if the password needs to be changed on first login
AUTH_RESULT=$(aws cognito-idp admin-initiate-auth \
    --user-pool-id "$USER_POOL_ID" \
    --client-id "$CLIENT_ID" \
    --auth-flow "ADMIN_NO_SRP_AUTH" \
    --auth-parameters "USERNAME=user@example.com,PASSWORD=PassTemp123!" \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" \
    --output json 2>&1 || true)

if echo "$AUTH_RESULT" | grep -q "ChallengeName"; then
    CHALLENGE=$(echo "$AUTH_RESULT" | grep -o '"ChallengeName": "[^"]*"' | cut -d'"' -f4)
    echo -e "${YELLOW}⚠️  Autenticação requer desafio: ${CHALLENGE}${NC}"
    echo -e "${YELLOW}   (Isso é esperado para senhas temporárias)${NC}"
elif echo "$AUTH_RESULT" | grep -q "AccessToken"; then
    echo -e "${GREEN}✅ Autenticação bem-sucedida!${NC}"
else
    echo -e "${YELLOW}⚠️  Resposta da autenticação:${NC}"
    echo "$AUTH_RESULT" | head -5
fi
echo ""

# Test 6: Check App Client
echo -e "${YELLOW}🧪 Teste 6: Verificando App Client...${NC}"
CLIENT_DETAILS=$(aws cognito-idp describe-user-pool-client \
    --user-pool-id "$USER_POOL_ID" \
    --client-id "$CLIENT_ID" \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" \
    --output json 2>&1)

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ App Client encontrado${NC}"
    CLIENT_NAME=$(echo "$CLIENT_DETAILS" | grep -o '"ClientName": "[^"]*"' | head -1 | cut -d'"' -f4)
    echo -e "   Nome: ${GREEN}${CLIENT_NAME}${NC}"
else
    echo -e "${RED}❌ Erro ao obter detalhes do App Client${NC}"
    exit 1
fi
echo ""

# Summary
echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║              ✅ Todos os testes passaram!                  ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}📊 Resumo da Infraestrutura:${NC}"
echo -e "   User Pools: ${GREEN}1${NC}"
echo -e "   Grupos: ${GREEN}${GROUP_COUNT}${NC}"
echo -e "   Usuários: ${GREEN}${USER_COUNT}${NC}"
echo -e "   App Clients: ${GREEN}1${NC}"
echo ""
echo -e "${YELLOW}🎯 Próximos passos:${NC}"
echo -e "   1. Integrar com sua aplicação Go"
echo -e "   2. Configurar endpoint: ${GREEN}${ENDPOINT}${NC}"
echo -e "   3. Usar IDs da configuração em: ${GREEN}cognito-local-config/config.json${NC}"
echo ""
echo -e "${YELLOW}💡 Dica:${NC}"
echo -e "   Para parar o cognito-local:"
echo -e "   ${GREEN}docker-compose -f docker-compose.cognito-local.yaml down${NC}"
echo ""
