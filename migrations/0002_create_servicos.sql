-- Criação da tabela de serviços
CREATE TABLE IF NOT EXISTS servicos (
  id          SERIAL PRIMARY KEY,
  nome        VARCHAR(255) NOT NULL UNIQUE,
  descricao   TEXT,
  duracao     INTEGER NOT NULL,
  preco       DECIMAL(10, 2) NOT NULL,
  ativo       BOOLEAN DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Índices para performance
CREATE INDEX IF NOT EXISTS idx_servicos_nome ON servicos(nome);
CREATE INDEX IF NOT EXISTS idx_servicos_ativo ON servicos(ativo);
