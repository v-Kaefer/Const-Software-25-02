-- Criação da tabela de agendamentos
CREATE TABLE IF NOT EXISTS agendamentos (
  id          SERIAL PRIMARY KEY,
  cliente_id  INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  servico_id  INTEGER NOT NULL REFERENCES servicos(id) ON DELETE RESTRICT,
  data_hora   TIMESTAMPTZ NOT NULL,
  status      VARCHAR(20) NOT NULL DEFAULT 'pendente',
  notas       TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Índices para performance
CREATE INDEX IF NOT EXISTS idx_agendamentos_cliente_id ON agendamentos(cliente_id);
CREATE INDEX IF NOT EXISTS idx_agendamentos_servico_id ON agendamentos(servico_id);
CREATE INDEX IF NOT EXISTS idx_agendamentos_data_hora ON agendamentos(data_hora);
CREATE INDEX IF NOT EXISTS idx_agendamentos_status ON agendamentos(status);

-- Índice único para prevenir agendamentos duplicados no mesmo horário
CREATE UNIQUE INDEX IF NOT EXISTS idx_agendamentos_slot 
ON agendamentos(servico_id, data_hora) 
WHERE status != 'cancelado';
