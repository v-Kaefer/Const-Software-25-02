# Make Commands for Windows (PowerShell)

Como o projeto usa `Makefile` mas Windows não possui `make` nativo, criamos um wrapper em PowerShell que mapeia todos os comandos.

## Instalação

1. **Permitir execução de scripts PowerShell** (primeira vez apenas):
```powershell
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope CurrentUser
```

2. **Usar o script** em qualquer lugar do projeto:
```powershell
.\make.ps1 <comando>
```

## Comandos Disponíveis

### LocalStack
```powershell
.\make.ps1 localstack-start     # Inicia LocalStack
.\make.ps1 localstack-stop      # Para LocalStack
.\make.ps1 localstack-status    # Verifica status
```

### Cognito-Local
```powershell
.\make.ps1 cognito-local-start  # Inicia cognito-local
.\make.ps1 cognito-local-stop   # Para cognito-local
.\make.ps1 cognito-local-setup  # Configura com Terraform
```

### Terraform Local
```powershell
.\make.ps1 tflocal-init         # Inicializa Terraform
.\make.ps1 tflocal-apply        # Aplica infraestrutura
.\make.ps1 tflocal-destroy      # Destroi infraestrutura
```

### Combinados (Recomendado)
```powershell
.\make.ps1 infra-up             # Inicia tudo (LocalStack + cognito-local + Terraform)
.\make.ps1 infra-down           # Para tudo
```

## Exemplo de Uso

Para subir a infraestrutura completa:

```powershell
# 1. Abra PowerShell no diretório do projeto
cd c:\Users\Administrador\Documents\cs\Const-Software-25-02

# 2. Execute o comando
.\make.ps1 infra-up

# 3. Aguarde a infraestrutura ser provisionada
# Recursos estarão disponíveis em:
# - S3: http://localhost:4566
# - DynamoDB: http://localhost:4566  
# - Cognito: http://localhost:9229
```

Para parar tudo:
```powershell
.\make.ps1 infra-down
```

## Troubleshooting

### "Script cannot be loaded because running scripts is disabled"
Solução: Executar com bypass (já incluído no comando acima)

### LocalStack não inicia
Verifique se Docker está rodando:
```powershell
docker version
```

### Erro em cognito-local-setup
Verifique se os scripts bash têm permissão de execução:
```powershell
ls -la infra/setup-cognito-local.sh
```

Se necessário, em um terminal bash/WSL:
```bash
chmod +x infra/setup-cognito-local.sh
chmod +x infra/test-cognito-local.sh
```

## Alternativa: Usar Makefile Original

Se instalar `make` no Windows via:
- **Chocolatey**: `choco install make`
- **msys2**: `pacman -S make`
- **Git Bash**: Já vem instalado

Então pode usar os comandos originais:
```powershell
make infra-up
make infra-down
```
