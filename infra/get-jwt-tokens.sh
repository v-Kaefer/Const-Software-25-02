#!/bin/bash

# Script to get JWT tokens from cognito-local for testing
# These tokens can be used in Swagger UI for authenticated requests

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

ENDPOINT="http://localhost:9229"
REGION="us-east-1"

# Set dummy AWS credentials for cognito-local
export AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID:-local}"
export AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY:-local}"
export AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-us-east-1}"

echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║              JWT Tokens from Cognito Local                ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if config file exists
if [ ! -f "${SCRIPT_DIR}/cognito-local-config/config.json" ]; then
    echo -e "${RED}❌ Configuration file not found!${NC}"
    echo -e "${YELLOW}   Run first: make cognito-local-setup${NC}"
    exit 1
fi

# Load configuration
USER_POOL_ID=$(cat "${SCRIPT_DIR}/cognito-local-config/config.json" | grep -o '"userPoolId": "[^"]*"' | cut -d'"' -f4)
CLIENT_ID=$(cat "${SCRIPT_DIR}/cognito-local-config/config.json" | grep -o '"clientId": "[^"]*"' | cut -d'"' -f4)

echo -e "${YELLOW}📋 Configuration:${NC}"
echo -e "   User Pool ID: ${GREEN}${USER_POOL_ID}${NC}"
echo -e "   Client ID: ${GREEN}${CLIENT_ID}${NC}"
echo ""

# Function to get token for a user
get_token() {
    local username=$1
    local password=$2
    local role=$3
    
    echo -e "${YELLOW}🔐 Getting token for: ${GREEN}${username}${NC} (${role})"
    
    # First, set permanent password to avoid NEW_PASSWORD_REQUIRED challenge
    aws cognito-idp admin-set-user-password \
        --user-pool-id "$USER_POOL_ID" \
        --username "$username" \
        --password "$password" \
        --permanent \
        --endpoint-url "$ENDPOINT" \
        --region "$REGION" 2>/dev/null || true
    
    # Try to authenticate
    AUTH_RESULT=$(aws cognito-idp initiate-auth \
        --client-id "$CLIENT_ID" \
        --auth-flow "USER_PASSWORD_AUTH" \
        --auth-parameters "USERNAME=${username},PASSWORD=${password}" \
        --endpoint-url "$ENDPOINT" \
        --region "$REGION" \
        --output json 2>&1)
    
    if echo "$AUTH_RESULT" | grep -q "IdToken"; then
        ID_TOKEN=$(echo "$AUTH_RESULT" | grep -o '"IdToken": "[^"]*"' | cut -d'"' -f4)
        ACCESS_TOKEN=$(echo "$AUTH_RESULT" | grep -o '"AccessToken": "[^"]*"' | cut -d'"' -f4)
        
        echo -e "${GREEN}✅ Authentication successful!${NC}"
        echo ""
        echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
        echo -e "${CYAN}ID Token (use this in Swagger 'Authorize'):${NC}"
        echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
        echo ""
        echo -e "Bearer ${ID_TOKEN}"
        echo ""
        echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
        echo ""
        return 0
    else
        echo -e "${RED}❌ Authentication failed${NC}"
        echo "$AUTH_RESULT" | head -3
        return 1
    fi
}

# Get tokens for each user
echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}                     ADMIN USER TOKEN                       ${NC}"
echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
get_token "admin@example.com" "AdminTemp123!" "admin-group"

echo ""
echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}                   REVIEWER USER TOKEN                      ${NC}"
echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
get_token "reviewer@example.com" "PassTemp123!" "reviewers-group"

echo ""
echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}                    REGULAR USER TOKEN                      ${NC}"
echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
get_token "user@example.com" "PassTemp123!" "user-group"

echo ""
echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                    How to use in Swagger                  ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}1. Open Swagger UI: ${GREEN}http://localhost:8081${NC}"
echo -e "${YELLOW}2. Click the ${GREEN}'Authorize'${NC} button (🔒)"
echo -e "${YELLOW}3. Paste the token (including 'Bearer ') in the value field${NC}"
echo -e "${YELLOW}4. Click ${GREEN}'Authorize'${NC} and then ${GREEN}'Close'${NC}"
echo -e "${YELLOW}5. Now you can test authenticated endpoints!${NC}"
echo ""
