package chatcomplitionstream

import (
	"context"
	"errors"

	"github.com/Paulo-DevTallos/imersao-fullcycle-lx/internal/domain/entities"
	"github.com/Paulo-DevTallos/imersao-fullcycle-lx/internal/domain/gateway"
	openai "github.com/sashabaranov/go-openai"
)

// definindo dados de entrada
type ChatComplitionConfigInputDTO struct {
	Model                string
	ModelMaxTokens       int
	Temperature          float32
	TopP                 float32
	N                    int
	Stop                 []string
	MaxTokens            int
	PresencePenalty      float32
	FrequencyPenalty     float32
	InitialSystemMessage string
}

// dados de complitions que o usuário envia
type ChatComplitionInputDTO struct {
	ChatID      string
	UserID      string
	UserMessage string
	Config      ChatComplitionConfigInputDTO
}

// dados de saída
type ChatComplitionOutputDTO struct {
	ChatID  string
	UserID  string
	Content string
}

type ChatComplitionUseCase struct {
	ChatGateway  gateway.ChatGateway
	OpenAiClient *openai.Client
}

func NewChatComplitionUseCase(chatGateway gateway.ChatGateway, openaiClient *openai.Client) *ChatComplitionUseCase {
	return &ChatComplitionUseCase{
		ChatGateway:  chatGateway,
		OpenAiClient: openaiClient,
	}
}

func (us *ChatComplitionUseCase) Execute(ctx context.Context, input ChatComplitionInputDTO) (*ChatComplitionOutputDTO, error) {
	chat, err := us.ChatGateway.FindChatById(ctx, input.ChatID)

	if err != nil {
		if err.Error() == "chat not found" {
			// cria um nobvo chat (entities)
			chat, err = createNewChat(input)
			if err != nil {
				return nil, errors.New("error creating new chat" + err.Error())
			}
			// salvar chat no banco de dados***
			err = us.ChatGateway.CreateChat(ctx, chat)
			if err != nil {
				return nil, errors.New("error persisting new chat" + err.Error())
			}
		} else {
			return nil, errors.New("error fatching existing chat" + err.Error())
		}
	}
	return nil, err
}

func createNewChat(input ChatComplitionInputDTO) (*entities.Chat, error) {
	model := entities.NewModel(input.Config.Model, input.Config.ModelMaxTokens)
	chatConfig := &entities.ChatConfig{
		Temperature:      input.Config.Temperature,
		TopP:             input.Config.TopP,
		N:                input.Config.N,
		Stop:             input.Config.Stop,
		MaxTokens:        input.Config.MaxTokens,
		PresencePenalty:  input.Config.PresencePenalty,
		FrequencyPenalty: input.Config.FrequencyPenalty,
		Model:            model,
	}

	initialMessage, err := entities.NewMessage("system", input.Config.InitialSystemMessage, model)
	if err != nil {
		return nil, errors.New("error creating initial message" + err.Error())
	}
	chat, err := entities.NewChat(input.UserID, initialMessage, chatConfig)
	if err != nil {
		return nil, errors.New("error creating new chat" + err.Error())
	}
	return chat, nil
}
