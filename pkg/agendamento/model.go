package agendamento

import (
	"time"

	"github.com/v-Kaefer/Const-Software-25-02/pkg/servico"
	"github.com/v-Kaefer/Const-Software-25-02/pkg/user"
)

// Status representa os estados possíveis de um agendamento
type Status string

const (
	Pendente  Status = "pendente"
	Aprovado  Status = "aprovado"
	Concluido Status = "concluido"
	Cancelado Status = "cancelado"
)

// Agendamento representa uma reserva de cliente para um serviço
type Agendamento struct {
	ID        uint      `gorm:"primaryKey"`
	ClienteID uint      `gorm:"not null;index"`
	ServicoID uint      `gorm:"not null;index"`
	DataHora  time.Time `gorm:"not null;index"`
	Status    Status    `gorm:"type:varchar(20);default:'pendente';index"`
	Notas     string    `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// Relacionamentos GORM
	Cliente *user.User        `gorm:"foreignKey:ClienteID;constraint:OnDelete:CASCADE"`
	Servico *servico.Servico `gorm:"foreignKey:ServicoID;constraint:OnDelete:RESTRICT"`
}

// TableName especifica o nome da tabela
func (Agendamento) TableName() string {
	return "agendamentos"
}
