package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
	tiktoken_go "github.com/j178/tiktoken-go"
)

type Message struct {
	ID        string
	Role      string
	Content   string
	Tokens    int
	Model     *Model
	CreatedAt time.Time
}

// o go permite mais de um retorno na função
/* primeiro ele trata os parametros que são repassados, caso não seja
cumprido nenhum ele retorna uma message vazia ou um erro*/
func newMessage(role, content string, model *Model) (*Message, error) {
	totalTokens := tiktoken_go.CountTokens(model.GetModelName(), content)
	msg := &Message{
		ID:        uuid.New().String(),
		Role:      role,
		Content:   content,
		Tokens:    totalTokens,
		Model:     model,
		CreatedAt: time.Now(),
	}

	if err := msg.Validate(); err != nil {
		return nil, err
	}

	return msg, nil
}

func (m *Message) Validate() error {
	if m.Role != "user" && m.Role != "system" && m.Role != "assistant" {
		return errors.New("Invalid role")
	}

	if m.Content == "" {
		return errors.New("Content is empty")
	}

	if m.CreatedAt.IsZero() {
		return errors.New("Invalid created at")
	}
	return nil
}

func (m *Message) GetQuantityTokens() int {
	return m.Tokens
}
