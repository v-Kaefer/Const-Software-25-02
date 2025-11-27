# Como Usar Make Commands no Windows

Como o projeto usa `Makefile` mas Windows não possui `make` nativo, fornecemos **2 alternativas**:

## Opção 1: PowerShell Script (Recomendado)

### Setup (primeira vez)
```powershell
# Permitir execução de scripts PowerShell
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process -Force
```

### Usar
```powershell
.\make.ps1 help           # Mostra ajuda
.\make.ps1 infra-up       # Inicia infraestrutura
.\make.ps1 infra-down     # Para infraestrutura
```

### Vantagens
- ✅ Colorido e mais visual
- ✅ Melhor mensagens de erro
- ✅ Mais flexível

### Desvantagens
- ⚠️ Precisa permitir scripts PowerShell

---

## Opção 2: Batch Script (Simples)

### Usar
```cmd
.\make.bat help           # Mostra ajuda
.\make.bat infra-up       # Inicia infraestrutura
.\make.bat infra-down     # Para infraestrutura
```

### Vantagens
- ✅ Funciona direto no CMD tradicional
- ✅ Sem necessidade de setup
- ✅ Compatível com qualquer Windows

### Desvantagens
- ⚠️ Menos visual (sem cores)
- ⚠️ Menos mensagens informativos

---

## Opção 3: PowerShell com Bypass (Sem Alterar Política)

Se não quer permitir scripts permanentemente:

```powershell
powershell -ExecutionPolicy Bypass -File .\make.ps1 infra-up
```

---

## Comandos Disponíveis

### LocalStack
```powershell
.\make.ps1 localstack-start      # Inicia LocalStack
.\make.ps1 localstack-stop       # Para LocalStack  
.\make.ps1 localstack-status     # Verifica status
```

### Cognito-Local
```powershell
.\make.ps1 cognito-local-start   # Inicia cognito-local
.\make.ps1 cognito-local-stop    # Para cognito-local
.\make.ps1 cognito-local-setup   # Configura cognito-local
.\make.ps1 cognito-local-clean   # Limpa cognito-local
```

### Terraform Local
```powershell
.\make.ps1 tflocal-init          # Inicializa Terraform
.\make.ps1 tflocal-plan          # Planeja infraestrutura
.\make.ps1 tflocal-apply         # Aplica infraestrutura
.\make.ps1 tflocal-destroy       # Destroi infraestrutura
```

### Combinados (Recomendado)
```powershell
.\make.ps1 infra-up              # Inicia TUDO (LocalStack + cognito-local + Terraform)
.\make.ps1 infra-down            # Para TUDO
```

---

## Pré-requisitos para Funcionar

Você precisa ter instalado:

1. **Docker** (com docker-compose)
2. **LocalStack CLI**: `pip install localstack`
3. **Terraform**: `choco install terraform` ou baixar em terraform.io
4. **tflocal**: `pip install terraform-local`
5. **AWS CLI**: `pip install awscli`

Veja `INSTALACAO_INFRA.md` para detalhes de instalação.

---

## Exemplos de Uso

### Exemplo 1: Iniciar infraestrutura completa
```powershell
# Abrir PowerShell e navegar ao projeto
cd C:\Users\Administrador\Documents\cs\Const-Software-25-02

# Permitir scripts (primeira vez)
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process -Force

# Iniciar infraestrutura
.\make.ps1 infra-up

# Aguardar ~30 segundos
# Recursos estarão em:
# - S3/DynamoDB: http://localhost:4566
# - Cognito: http://localhost:9229
```

### Exemplo 2: Parar infraestrutura
```powershell
.\make.ps1 infra-down
```

### Exemplo 3: Usar batch no CMD tradicional
```cmd
cd C:\Users\Administrador\Documents\cs\Const-Software-25-02
.\make.bat infra-up
.\make.bat infra-down
```

---

## Troubleshooting

### Erro: "não é reconhecido como nome de cmdlet"
**Solução**: Use `.\` antes do comando:
```powershell
.\make.ps1 help          # Correto
make.ps1 help            # Errado
```

### Erro: "scripts is disabled"
**Solução**: Permitir scripts (escolha uma):
```powershell
# Temporário (apenas esta sessão)
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process -Force

# Permanente (todos os usuários)
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope CurrentUser

# Ou use bypass inline
powershell -ExecutionPolicy Bypass -File .\make.ps1 infra-up
```

### LocalStack não inicia
```powershell
# Verificar se Docker está rodando
docker --version

# Se não:
# Windows: Abrir Docker Desktop
# WSL: docker daemon start
```

### Terraform não encontrado
```powershell
# Instalar Terraform
choco install terraform

# Ou baixar em https://www.terraform.io/downloads
```

---

## Alternativa: Instalar Make Real

Se preferir usar `make` real:

### Windows (Chocolatey)
```powershell
choco install make
```

### Windows (scoop)
```powershell
scoop install make
```

Depois pode usar:
```powershell
make infra-up
make infra-down
```

---

## Estrutura

```
Const-Software-25-02/
├── make.ps1              # PowerShell wrapper (recomendado)
├── make.bat              # Batch wrapper (simples)
├── Makefile              # Makefile original (para Linux/Mac)
├── INSTALACAO_INFRA.md   # Guia de instalação de dependências
└── README_MAKE.md        # Este arquivo
```
