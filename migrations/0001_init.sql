-- Extensão UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Tabela de alunos
CREATE TABLE IF NOT EXISTS alunos (
  id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  nome          TEXT        NOT NULL,
  email         TEXT        NOT NULL UNIQUE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alunos_email ON alunos (email);
CREATE INDEX IF NOT EXISTS idx_alunos_created_at ON alunos (created_at);

-- Tabela de matrículas
CREATE TABLE IF NOT EXISTS matriculas (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  aluno_id        UUID        NOT NULL REFERENCES alunos(id) ON DELETE CASCADE,
  curso           TEXT        NOT NULL,
  data_matricula  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  status          TEXT        NOT NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_matriculas_aluno_id ON matriculas (aluno_id);
CREATE INDEX IF NOT EXISTS idx_matriculas_curso ON matriculas (curso);
CREATE INDEX IF NOT EXISTS idx_matriculas_data_matricula ON matriculas (data_matricula);