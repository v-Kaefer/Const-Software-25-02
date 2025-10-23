#!/bin/bash

# LocalStack Manager Script
# Alternative to Makefile for managing LocalStack infrastructure

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${GREEN}ℹ️  $1${NC}"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
check_prerequisites() {
    local missing=0
    
    echo "Verificando pré-requisitos..."
    
    if ! command_exists localstack; then
        print_error "LocalStack não encontrado. Instale com: pip install localstack"
        missing=1
    else
        print_success "LocalStack encontrado"
    fi
    
    if ! command_exists terraform; then
        print_error "Terraform não encontrado. Instale de: https://www.terraform.io/downloads"
        missing=1
    else
        print_success "Terraform encontrado"
    fi
    
    if ! command_exists aws; then
        print_error "AWS CLI não encontrado. Instale com: pip install awscli"
        missing=1
    else
        print_success "AWS CLI encontrado"
    fi
    
    if [ $missing -eq 1 ]; then
        exit 1
    fi
}

# Start LocalStack
start_localstack() {
    print_info "Iniciando LocalStack..."
    localstack start -d
    sleep 10
    localstack status
    print_success "LocalStack iniciado!"
}

# Stop LocalStack
stop_localstack() {
    print_info "Parando LocalStack..."
    localstack stop
    print_success "LocalStack parado!"
}

# Check LocalStack status
check_status() {
    print_info "Verificando status do LocalStack..."
    localstack status || print_error "LocalStack não está rodando"
}

# Initialize Terraform
terraform_init() {
    print_info "Inicializando Terraform..."
    cd "$(dirname "$0")"
    terraform init
    print_success "Terraform inicializado!"
}

# Terraform plan
terraform_plan() {
    print_info "Executando terraform plan..."
    cd "$(dirname "$0")"
    terraform plan
}

# Terraform apply
terraform_apply() {
    print_info "Aplicando infraestrutura..."
    cd "$(dirname "$0")"
    terraform apply -auto-approve
    print_success "Infraestrutura aplicada!"
}

# Terraform destroy
terraform_destroy() {
    print_info "Destruindo infraestrutura..."
    cd "$(dirname "$0")"
    terraform destroy -auto-approve
    print_success "Infraestrutura destruída!"
}

# Test infrastructure
test_infrastructure() {
    print_info "Testando infraestrutura..."
    echo ""
    
    echo "1️⃣ Testando S3..."
    aws --endpoint-url=http://localhost:4566 s3 ls || print_error "S3 não disponível"
    echo ""
    
    echo "2️⃣ Testando DynamoDB..."
    aws --endpoint-url=http://localhost:4566 dynamodb list-tables || print_error "DynamoDB não disponível"
    echo ""
    
    echo "3️⃣ Testando Cognito (requer LocalStack Pro)..."
    aws --endpoint-url=http://localhost:4566 cognito-idp list-user-pools --max-results 10 || print_warning "Cognito não disponível no free tier"
    echo ""
    
    print_success "Teste concluído!"
}

# Full setup
full_setup() {
    print_warning "======================================"
    print_warning "ATENÇÃO: Cognito requer LocalStack Pro"
    print_warning "======================================"
    echo ""
    echo "Se você tem LocalStack Pro, defina a variável:"
    echo "  export LOCALSTACK_API_KEY=seu-api-key"
    echo ""
    echo "Para testar sem Cognito (free tier):"
    echo "  mv cognito.tf cognito.tf.disabled"
    echo ""
    read -p "Continuar? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 0
    fi
    
    check_prerequisites
    start_localstack
    terraform_init
    terraform_apply
    test_infrastructure
    
    print_success "Setup completo!"
    print_info "Use '$0 down' para destruir tudo"
}

# Full teardown
full_teardown() {
    terraform_destroy
    stop_localstack
    print_success "Teardown completo!"
}

# Show help
show_help() {
    cat << EOF
LocalStack Manager - Gerenciador de Infraestrutura LocalStack

Uso: $0 [comando]

Comandos:
  check           - Verifica pré-requisitos
  start           - Inicia LocalStack
  stop            - Para LocalStack
  status          - Verifica status do LocalStack
  
  init            - Inicializa Terraform
  plan            - Executa terraform plan
  apply           - Aplica infraestrutura
  destroy         - Destrói infraestrutura
  
  test            - Testa a infraestrutura criada
  
  up              - Setup completo (start + init + apply)
  down            - Teardown completo (destroy + stop)
  
  help            - Mostra esta ajuda

Exemplos:
  $0 up           # Inicia tudo
  $0 test         # Testa recursos
  $0 down         # Destrói tudo

Para mais informações, veja README.md e QUICKSTART.md
EOF
}

# Main script
case "${1:-help}" in
    check)
        check_prerequisites
        ;;
    start)
        start_localstack
        ;;
    stop)
        stop_localstack
        ;;
    status)
        check_status
        ;;
    init)
        terraform_init
        ;;
    plan)
        terraform_plan
        ;;
    apply)
        terraform_apply
        ;;
    destroy)
        terraform_destroy
        ;;
    test)
        test_infrastructure
        ;;
    up)
        full_setup
        ;;
    down)
        full_teardown
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        print_error "Comando desconhecido: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
