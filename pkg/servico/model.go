package servico

import "time"

// Servico representa um serviço oferecido pelo sistema
// Ex: Corte de Cabelo, Massagem, Consulta Médica, etc.
type Servico struct {
	ID        uint      `gorm:"primaryKey"`
	Nome      string    `gorm:"size:255;not null;uniqueIndex:idx_servicos_nome"`
	Descricao string    `gorm:"type:text"`
	Duracao   int       `gorm:"not null"` // Minutos
	Preco     float64   `gorm:"type:decimal(10,2);not null"`
	Ativo     bool      `gorm:"default:true;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName especifica o nome da tabela
func (Servico) TableName() string {
	return "servicos"
}
