# Guia de Instalação - Infraestrutura Local

## Pré-requisitos

### 1. Docker
Necessário para rodar LocalStack e cognito-local.

**Windows:**
- Baixar [Docker Desktop](https://www.docker.com/products/docker-desktop)
- Instalar e reiniciar
- Verificar instalação:
```powershell
docker --version
docker-compose --version
```

### 2. LocalStack CLI
LocalStack fornece um CLI para gerenciar a infraestrutura local.

**Windows (via pip):**
```powershell
# Instalar Python primeiro se não tiver
python --version

# Instalar LocalStack CLI
pip install localstack

# Verificar instalação
localstack --version
```

**Windows (via Chocolatey):**
```powershell
choco install localstack
```

**Windows (via Docker - alternativa, sem CLI instalado):**
Se não quiser instalar o CLI, pode usar Docker Compose diretamente:
```powershell
# Em vez de: localstack start
# Use: docker-compose -f docker-compose.yaml up -d
```

### 3. Terraform
Necessário para provisionar infraestrutura.

**Windows (via Chocolatey):**
```powershell
choco install terraform
```

**Windows (via scoop):**
```powershell
scoop install terraform
```

**Ou baixar manualmente:**
- [terraform.io/downloads](https://www.terraform.io/downloads)

### 4. tflocal (Terraform + LocalStack)
Simplifica integração entre Terraform e LocalStack.

```powershell
pip install terraform-local
```

### 5. AWS CLI
Útil para testar resources criados.

```powershell
pip install awscli
```

## Instalação Rápida (PowerShell com Chocolatey)

Se tem Chocolatey instalado:

```powershell
# Abrir PowerShell como Admin
choco install docker-desktop terraform

# Instalar ferramentas Python
pip install localstack terraform-local awscli

# Verificar instalações
docker --version
localstack --version
terraform --version
tflocal --version
aws --version
```

## Instalação Rápida (WSL + PowerShell)

Se prefere usar WSL (Windows Subsystem for Linux):

```bash
# No terminal WSL/Bash
sudo apt-get install -y awscli docker.io
pip install localstack terraform-local

# Adicionar user ao grupo docker
sudo usermod -aG docker $USER
```

## Testando Instalação

```powershell
# 1. Verificar Docker
docker ps

# 2. Verificar LocalStack CLI
localstack --version

# 3. Verificar Terraform
terraform --version

# 4. Verificar tflocal
tflocal --version
```

## Iniciando a Infraestrutura

Depois de instalar tudo:

```powershell
# Navegar ao diretório do projeto
cd C:\Users\Administrador\Documents\cs\Const-Software-25-02

# Permitir scripts PowerShell (primeira vez)
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope CurrentUser

# Iniciar infraestrutura completa
.\make.ps1 infra-up

# Aguarde ~30 segundos para tudo estar pronto

# Testar se está rodando
.\make.ps1 localstack-status
```

## Resources Disponíveis

Após `.\make.ps1 infra-up`, você terá:

| Serviço | Endpoint | Notas |
|---------|----------|-------|
| S3 | http://localhost:4566 | Buckets, objetos |
| DynamoDB | http://localhost:4566 | Tabelas, scans |
| Cognito | http://localhost:9229 | User pools, tokens |
| LocalStack Admin | http://localhost:4566/_localstack/health | Dashboard |

## Environment Variables

Configure em `~/.bashrc`, `~/.zshrc` ou PowerShell profile:

```powershell
$env:AWS_ACCESS_KEY_ID = "test"
$env:AWS_SECRET_ACCESS_KEY = "test"
$env:AWS_DEFAULT_REGION = "us-east-1"
```

## Troubleshooting

### LocalStack não inicia
```powershell
# Verificar logs
docker logs localstack

# Reiniciar
docker restart localstack

# Limpar completamente
.\make.ps1 localstack-clean
.\make.ps1 localstack-start
```

### Cognito-local não funciona
```powershell
# Verificar container
docker ps | grep cognito

# Visualizar logs
docker logs cognito-local

# Reiniciar
docker-compose -f docker-compose.cognito-local.yaml restart
```

### Portas já em uso
Se as portas 4566 ou 9229 estão em uso:

```powershell
# Windows - encontrar processo usando porta
netstat -ano | findstr :4566

# Matar processo (substitua PID)
taskkill /PID <PID> /F
```

## Próximos Passos

Depois de tudo rodando:

1. **Testar infraestrutura:**
   ```powershell
   .\make.ps1 infra-test
   ```

2. **Desenvolver contra infraestrutura local:**
   - Set environment variables (veja acima)
   - Use `localhost:4566` para AWS services
   - Use `localhost:9229` para Cognito

3. **Parar infraestrutura:**
   ```powershell
   .\make.ps1 infra-down
   ```
