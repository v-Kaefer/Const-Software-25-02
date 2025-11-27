# 🚀 COMANDO PARA CRIAR A BRANCH - Release 4.0

## Padrão de Nomenclatura do Seu Repositório

Baseado nas branches existentes, seu projeto segue este padrão:
```
feat/Sprint0
feat/Sprint1
feat/Autenticacao
feat/Localstack-terraform
```

## ✅ COMANDO RECOMENDADO (Com Padrão Git)

Escolha uma das opções abaixo:

### **Opção 1: Padrão Simples (RECOMENDADO)**
```bash
git checkout -b feat/release-4-0-api-rest
```

### **Opção 2: Padrão com Release**
```bash
git checkout -b release/4.0-sistema-real
```

### **Opção 3: Padrão Feature Completo**
```bash
git checkout -b feat/Release4.0-SistemaReal
```

### **Opção 4: Padrão Feature com Sprint**
```bash
git checkout -b feat/Sprint2-Release4.0
```

---

## 🎯 RECOMENDAÇÃO

Eu recomendo a **Opção 1**:
```bash
git checkout -b feat/release-4-0-api-rest
```

**Por que?**
- ✅ Segue seu padrão de nomenclatura
- ✅ Descritivo e claro
- ✅ Fácil de ler em git log
- ✅ Lowercase (padrão Git)
- ✅ Usa hífens (padrão Git)

---

## 📋 PASSOS COMPLETOS PARA COMEÇAR

### **Passo 1: Garantir que está na develop**
```bash
git checkout develop
git pull origin develop
```

### **Passo 2: Criar a nova branch**
```bash
git checkout -b feat/release-4-0-api-rest
```

### **Passo 3: Verificar que está na branch correta**
```bash
git branch
```
Você deve ver:
```
* feat/release-4-0-api-rest
  develop
  main
```

### **Passo 4: Fazer o push inicial (opcional, mas recomendado)**
```bash
git push -u origin feat/release-4-0-api-rest
```

### **Passo 5: Começar a codificar!**
```bash
# Criar os arquivos
mkdir -p pkg/servico
mkdir -p pkg/agendamento

# ... e começar a desenvolver
```

---

## 🔄 DEPOIS QUANDO TERMINAR (PULL REQUEST)

```bash
# 1. Commit suas mudanças
git add .
git commit -m "feat: implementar entidades servico e agendamento"

# 2. Push para sua branch
git push origin feat/release-4-0-api-rest

# 3. No GitHub, criar Pull Request (PR)
#    - Base: develop
#    - Compare: feat/release-4-0-api-rest
```

---

## 📝 COMANDOS ÚTEIS DURANTE DESENVOLVIMENTO

### Ver status
```bash
git status
```

### Ver mudanças
```bash
git diff
```

### Fazer commit
```bash
git add .
git commit -m "feat: descrição da mudança"
```

### Atualizar com develop
```bash
git pull origin develop
```

### Ver log
```bash
git log --oneline -10
```

---

## ✅ CHECKLIST PRÉ-DESENVOLVIMENTO

```
[ ] git checkout develop
[ ] git pull origin develop
[ ] git checkout -b feat/release-4-0-api-rest
[ ] git branch (verificar)
[ ] Abrir VS Code
[ ] Ler 00_LEIA_PRIMEIRO.md
[ ] Lembrar do domínio escolhido
[ ] Começar com START_NOW.md
```

---

## 🎉 PRONTO!

Agora é só copiar e colar o comando:

```bash
git checkout develop && git pull origin develop && git checkout -b feat/release-4-0-api-rest
```

Depois verifique:
```bash
git branch
```

**E pronto! Você está pronto para começar! 🚀**

---

## 📞 DÚVIDAS FREQUENTES

**P: Preciso fazer push da branch?**
R: Não é obrigatório agora, mas é recomendado para backup:
```bash
git push -u origin feat/release-4-0-api-rest
```

**P: Posso mudar o nome da branch?**
R: Sim, se não fez push ainda:
```bash
git branch -m novo-nome
```

**P: Acidentei e comitei na develop?**
R: Sem problema, recupere com:
```bash
git reset HEAD~1
git checkout -b feat/release-4-0-api-rest
```

**P: Como voltar para develop depois?**
R: Simples:
```bash
git checkout develop
git pull origin develop
```

---

**Está pronto para começar? 🚀**

Execute o comando acima e comece a codificar!
