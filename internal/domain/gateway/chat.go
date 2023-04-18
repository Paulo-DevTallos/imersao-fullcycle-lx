package gateway

import (
	"context"

	"github.com/Paulo-DevTallos/imersao-fullcycle-lx/internal/domain/entities"
)

type ChatGateway interface {
	CreateChat(ctx context.Context, chat *entities.Chat) error
	FindChatById(ctx context.Context, chaID string) (*entities.Chat, error)
	SaveChat(ctx context.Context, chat *entities.Chat) error
}
