package chatcomplitionstream

import (
	"context"
	"errors"
	"io"
	"strings"

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
	Stream       chan ChatComplitionOutputDTO // conforme recebe pega os dados no canal e joga para uma outra thread
}

func NewChatComplitionUseCase(chatGateway gateway.ChatGateway, openaiClient *openai.Client, stream chan ChatComplitionOutputDTO) *ChatComplitionUseCase {
	return &ChatComplitionUseCase{
		ChatGateway:  chatGateway,
		OpenAiClient: openaiClient,
		Stream:       stream,
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
	userMessage, err := entities.NewMessage("user", input.UserMessage, chat.Config.Model)
	if err != nil {
		return nil, errors.New("error creating new chat" + err.Error())
	}
	err = chat.AddMessage(userMessage)
	if err != nil {
		return nil, errors.New("error adding new message" + err.Error())
	}
	// _ representa "&"
	messages := []openai.ChatCompletionMessage{}
	for _, msg := range chat.Messages {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	resp, err := us.OpenAiClient.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model:            chat.Config.Model.Name,
			Messages:         messages,
			MaxTokens:        chat.Config.MaxTokens,
			Temperature:      chat.Config.Temperature,
			TopP:             chat.Config.TopP,
			PresencePenalty:  chat.Config.PresencePenalty,
			FrequencyPenalty: chat.Config.FrequencyPenalty,
			Stop:             chat.Config.Stop,
			Stream:           true,
		},
	)
	if err != nil {
		return nil, errors.New("error creating chat complition: " + err.Error())
	}

	var fullResponse strings.Builder
	for {
		// Recv() recebendo os dados
		response, err := resp.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, errors.New("error streaming response: " + err.Error())
		}
		// para isso é necessário estudar o package com o tipo de resposta que obtemos do chatComplitionStreamResponse da biblioteca
		fullResponse.WriteString(response.Choices[0].Delta.Content)
		r := ChatComplitionOutputDTO{
			ChatID:  chat.ID,
			UserID:  chat.UserID,
			Content: fullResponse.String(),
		}

		us.Stream <- r
	}

	assistant, err := entities.NewMessage("assistant", fullResponse.String(), chat.Config.Model)
	if err != nil {
		return nil, errors.New("error creating assistant message: " + err.Error())
	}
	err = chat.AddMessage(assistant)
	if err != nil {
		return nil, errors.New("error adding new message: " + err.Error())
	}

	err = us.ChatGateway.SaveChat(ctx, chat)
	if err != nil {
		return nil, errors.New("error saving chat: " + err.Error())
	}

	return &ChatComplitionOutputDTO{
		ChatID:  chat.ID,
		UserID:  input.UserID,
		Content: fullResponse.String(),
	}, nil
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
