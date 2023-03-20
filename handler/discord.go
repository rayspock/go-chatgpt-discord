package handler

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/rayspock/go-chatgpt-discord/provider"
	log "github.com/sirupsen/logrus"
)

const (
	ApplicationCommandChat = "chat"
)

type DiscordHandler struct {
	chatGPT provider.ChatGPTService
}

func NewDiscordHandler(chatGPT provider.ChatGPTService) *DiscordHandler {
	return &DiscordHandler{chatGPT: chatGPT}
}

func (h *DiscordHandler) GetInteractionCreateHandler() func(s *discordgo.Session, i *discordgo.InteractionCreate) {

	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		log.Tracef("in GetInteractionCreateHandler, type:%v", i.Type)

		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		data := i.ApplicationCommandData()

		if data.Name != ApplicationCommandChat {
			return
		}

		if len(data.Options) <= 0 {
			log.Warn("empty options")
			return
		}

		content, ok := data.Options[0].Value.(string)
		if !ok {
			_, err := s.ChannelMessageSend(i.ChannelID, fmt.Sprintf("invalid message"))
			if err != nil {
				log.Errorf("couldn't send message: invalid message")
			}
			return
		}

		var author string
		if i.User != nil {
			author = i.User.ID
		}
		if i.Member != nil {
			if i.Member.User != nil {
				author = i.Member.User.ID
			}
		}
		response := getUserMessage(content, author)

		// If we don't send a response in 3 seconds, the error 'The application did not respond' will appear.
		// To avoid this, we send the following type of response
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
		if err != nil {
			log.Errorf("failed to defer message: %v", err)
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "Something went wrong",
			})
			return
		}

		req := provider.ChatCompletionRequest{Messages: []provider.ChatCompletionMessage{
			{
				Role:    provider.ChatMessageRoleUser,
				Content: content,
			},
		}}
		resp, err := h.chatGPT.CreateChatCompletion(context.Background(), req)
		if err != nil {
			log.Errorf("failed to create chat completion: %v", err)
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "Something went wrong",
			})
			return
		}

		//log.Tracef("received message: %v", resp.Content)

		response += resp.Content
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &response,
		})
		if err != nil {
			log.Errorf("failed to edit response: %v", err)
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "Something went wrong",
			})
			return
		}

	}
}

func (h *DiscordHandler) GetMessageCreateHandler() func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		log.Tracef("in GetMessageCreateHandler, type:%v, content:%v", m.Type, m.Content)
	}
}
