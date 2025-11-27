# 📋 LISTA RÁPIDA - O QUE FALTA FAZER

## CRÍTICO (Sem isso não passa) 🔴

### 1. **3+ Entidades de Negócio**
```
❌ FALTANDO: Seu domínio de negócio

Atualmente tem:
✅ User (entidade técnica)

Precisa ter:
❌ Entidade 1 (ex: Servico)
❌ Entidade 2 (ex: Agendamento)  
❌ Entidade 3+ (ex: Avaliacao)

Impacto: 60 pontos da nota
Prazo: CRÍTICO - Semana 1
```

### 2. **2-3 Fluxos de Negócio Completos**
```
❌ FALTANDO: Fluxos ponta-a-ponta

Exemplos:
- Agendar → Confirmar → Realizar → Cancelar
- Criar Pedido → Pagar → Entregar → Avaliar
- Matricular → Frequentar → Avaliar → Certificar

Impacto: 40 pontos da nota
Prazo: CRÍTICO - Semana 2
```

### 3. **Rotas /api/v1/... (Versionamento)**
```
❌ FALTANDO: Versionamento na URL

Mudar de:
GET /users
GET /users/{id}
POST /users

Para:
GET /api/v1/users
GET /api/v1/users/{id}
POST /api/v1/users
```

---

## IMPORTANTE (Sem isso perde pontos) 🟠

### 4. **Paginação & Filtros em Listagens**
```
❌ FALTANDO: Paginação e filtros

Implementar:
GET /api/v1/usuarios?page=1&limit=10
GET /api/v1/agendamentos?status=pendente
GET /api/v1/produtos?categoria=livros&minPreco=10&maxPreco=100

Response:
{
  "data": [...],
  "total": 150,
  "page": 1,
  "limit": 10,
  "pages": 15
}
```

### 5. **Migrações SQL Corretas**
```
❌ INCONSISTÊNCIA: migrations/0001_init.sql refencia tabelas 
                   (alunos, matriculas) que não existem no código Go

Corrigir:
- Manter migration existente (users)
- Criar 0002_create_servicos.sql
- Criar 0003_create_agendamentos.sql
- Garantir que GORM models correspondem
```

### 6. **OpenAPI Atualizado**
```
❌ FALTANDO: Schemas para novas entidades

Adicionar:
- Servico schema
- Agendamento schema
- Todos os endpoints das 3+ entidades
- Exemplos de erro (400, 401, 403, 404, 500)
- Parâmetros de paginação
```

### 7. **Testes para Fluxos**
```
❌ FALTANDO: Testes de fluxos de negócio

Criar testes para:
- Agendar serviço (happy path)
- Agendar com data passada (erro esperado)
- Agendar com conflito de horário (erro esperado)
- Aprovar agendamento (role: admin)
- Cancelar agendamento (role: admin/owner)
- Listar agendamentos com paginação
```

---

## RECOMENDADO (Melhora a nota) 🟡

### 8. **Coleção Postman ou Scripts curl**
```
Criar arquivo: tests/requests.sh ou postman_collection.json

Exemplos:
#!/bin/bash
ADMIN_TOKEN="..."
USER_TOKEN="..."

# Criar serviço
curl -X POST \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "nome": "Corte de Cabelo",
    "duracao": 30,
    "preco": 50.00
  }' \
  http://localhost:8080/api/v1/servicos

# Agendar serviço
curl -X POST \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "servico_id": 1,
    "data_hora": "2025-12-01T14:00:00Z"
  }' \
  http://localhost:8080/api/v1/agendamentos
```

### 9. **Documentar Fluxos no README**
```
Adicionar seção:

## Fluxos de Negócio

### Fluxo 1: Agendar um Serviço

1. Admin cria serviços:
   POST /api/v1/servicos
   
2. Cliente agenda serviço:
   POST /api/v1/agendamentos
   
3. Admin aprova:
   PATCH /api/v1/agendamentos/{id} {"status": "aprovado"}
   
4. Cliente cancela (se necessário):
   DELETE /api/v1/agendamentos/{id}
```

### 10. **Validações de Negócio**
```
Implementar regras como:
- Data no futuro
- Não agendar 2x mesmo horário
- Aprovar apenas agendamentos pendentes
- Cancelar apenas agendamentos não concluídos
- Email único
- Campos obrigatórios
```

---

## RESUMO DE PRIORIDADES

| # | Tarefa | Status | Prazo | Impacto |
|---|--------|--------|-------|---------|
| 1 | 3+ Entidades | ❌ | URGENTE | 🔴 CRÍTICO |
| 2 | 2-3 Fluxos | ❌ | URGENTE | 🔴 CRÍTICO |
| 3 | Versionamento /api/v1 | ❌ | Até Semana 2 | 🟠 IMPORTANTE |
| 4 | Paginação & Filtros | ❌ | Até Semana 2 | 🟠 IMPORTANTE |
| 5 | Migrações Corretas | ⚠️ | Até Semana 1 | 🟠 IMPORTANTE |
| 6 | OpenAPI Completo | ⚠️ | Até Semana 2 | 🟠 IMPORTANTE |
| 7 | Testes Fluxos | ❌ | Até Semana 3 | 🟠 IMPORTANTE |
| 8 | Postman/curl | ❌ | Até Semana 3 | 🟡 RECOMENDADO |
| 9 | README Fluxos | ❌ | Até Semana 3 | 🟡 RECOMENDADO |
| 10 | Validações | ⚠️ | Até Semana 2 | 🟡 RECOMENDADO |

---

## COMO COMEÇAR (HOJE)

### Passo 1: Reunir o Grupo (1h)
- [ ] Decidir domínio (Agendamento? E-commerce? Cursos?)
- [ ] Listar 3+ entidades
- [ ] Desenhar relacionamentos

### Passo 2: Criar Models (2h)
```
pkg/servico/model.go          (criar)
pkg/agendamento/model.go      (criar)
```

### Passo 3: Criar Migrations (2h)
```
migrations/0002_servicos.sql
migrations/0003_agendamentos.sql
```

### Passo 4: Criar Repositórios (3h)
```
pkg/servico/repo.go
pkg/agendamento/repo.go
```

### Passo 5: Criar Serviços (3h)
```
pkg/servico/service.go
pkg/agendamento/service.go
```

### Passo 6: Criar Handlers (4h)
```
internal/http/servico_handler.go
internal/http/agendamento_handler.go
```

**Total Estimado: 15h = 2 dias de trabalho**

---

## COMANDOS ÚTEIS

### Gerar testes boilerplate
```bash
cd c:\Users\Administrador\Documents\cs\Const-Software-25-02
go test ./... -v
```

### Verificar estrutura
```bash
tree pkg/
tree internal/http/
```

### Formatar código
```bash
go fmt ./...
go vet ./...
```

---

## FAQ

**P: Por onde começo?**
R: Escolha o domínio → Crie 3 models → Crie migrations → Crie repo → Crie service → Crie handler

**P: Posso fazer tudo hoje?**
R: Não. Domínio + models + migrations + repos = ~8h. Deixe handlers para amanhã.

**P: E se mudar de ideia sobre o domínio?**
R: Simples. Delete os arquivos criados e comece novo. Por isso decide rápido!

**P: Qual domínio é mais fácil?**
R: Agendamento. Já tem migration de "alunos", é intuitivo, e tem 3 entidades naturais.

**P: Preciso de testes agora?**
R: Não. Faça código funcionar primeiro. Testes depois.

**P: Como testo localmente?**
R: `docker compose up` → Usa curl/Postman → Verifica em http://localhost:8081 (Swagger)

---

**Status:** Pronto para começar!
**Última atualização:** 27 de novembro de 2025
