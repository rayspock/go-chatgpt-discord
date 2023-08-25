package handler

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/rayspock/go-chatgpt-discord/provider"
	log "github.com/sirupsen/logrus"
	"unicode/utf8"
)

const (
	ApplicationCommandChat string = "chat"
)

const (
	MaxMessageLength int = 2000
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
		switch data.Name {
		case ApplicationCommandChat:
			h.handleChatCommand(s, i)
		}

	}
}

func (h *DiscordHandler) handleChatCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

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
	// reformat user's message and append to response
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

	response += resp.Content
	// split the response into chunks, as Discord has a limit on the length of messages sent.
	chunks := make(chan string)
	go SendMessageByChunk(response, MaxMessageLength, chunks)
	for chunk := range chunks {
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: chunk,
		})
		if err != nil {
			log.Errorf("creates the followup message for an interaction: %v", err)
		}
	}
}

// GetMessageCreateHandler This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
//
// It is called whenever a message is created but only when it's sent through a
// server as we did not request IntentsDirectMessages.
func (h *DiscordHandler) GetMessageCreateHandler() func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore all messages created by the bot itself
		if m.Author.ID == s.State.User.ID {
			return
		}

		// Ignore all messages about interactions as we handle them in GetInteractionCreateHandler
		if m.Interaction != nil && m.Interaction.ID != "" {
			return
		}

		// Ignore all messages sent in a guild
		if m.GuildID != "" {
			return
		}

		h.handleDirectMessage(s, m)

	}
}

func (h *DiscordHandler) handleDirectMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// We create the private channel with the user who sent the message.
	channel, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		// If an error occurred, we failed to create the channel.
		//
		// Some common causes are:
		// 1. We don't share a server with the user (not possible here).
		// 2. We opened enough DM channels quickly enough for Discord to
		//    label us as abusing the endpoint, blocking us from opening
		//    new ones.
		log.Errorf("error creating channel: %v", err)
		s.ChannelMessageSend(
			m.ChannelID,
			"Something went wrong while sending the DM!",
		)
		return
	}

	log.Tracef("received message: %v", m.Content)

	// Send message to chatGPT
	req := provider.ChatCompletionRequest{Messages: []provider.ChatCompletionMessage{
		{
			Role:    provider.ChatMessageRoleUser,
			Content: m.Content,
		},
	}}
	recv := make(chan provider.ChatCompletionStreamResponse)
	go func() {
		err = h.chatGPT.CreateChatCompletionStream(context.Background(), req, recv)
		if err != nil {
			log.Errorf("failed to create chat completion stream: %v", err)
		}
	}()

	// Create a function to send typing indicator
	channelTyping := func() {
		err = s.ChannelTyping(channel.ID)
		if err != nil {
			log.Errorf("error sending typing indicator: %v", err)
		}
	}
	// Send typing indicator
	channelTyping()

	// Create a function to send message to user
	sendMessage := func(messageID, content string) string {
		if messageID == "" {
			_msg, err := s.ChannelMessageSend(channel.ID, content)
			if err != nil {
				log.Errorf("error sending DM message: %v", err)
			}
			return _msg.ID
		}
		_msg, err := s.ChannelMessageEdit(channel.ID, messageID, content)
		if err != nil {
			log.Errorf("error editing message: %v", err)
		}
		return _msg.ID
	}

	// Use to keep track of the last message id sent to the user, to edit the message instead of sending a new one
	var msgID string
	// A buffer string to store the last completion
	var chunkToRead string
	var readChunkLength int
	const (
		// Discord has a limit of 2000 characters per chunk of message
		maxContentLengthPerChunk = 2000
		// The interval of characters to send a message, to keep the user posted on the progress
		intervalOfCharacters = 100
	)

	for resp := range recv {
		if len(resp.Choices) <= 0 {
			continue
		}
		incomingContent := resp.Choices[0].Delta.Content
		if incomingContent == "" {
			continue
		}

		// If the message is too long, split it into multiple messages
		chunkToRead += incomingContent
		for i, w := readChunkLength, 0; i < len(chunkToRead); i += w {
			_, w = utf8.DecodeRuneInString(chunkToRead[i:])
			readChunkLength += w
			if readChunkLength > maxContentLengthPerChunk {
				sendMessage(msgID, chunkToRead[:i])
				// Update the chunkToRead to the rest of the completion that hasn't been sent to the user.
				chunkToRead = chunkToRead[i:]
				// Reset the message id, so that the next message is sent as a new message
				msgID = ""
				// Reset the readChunkLength, so that the next message is sent from the beginning
				readChunkLength = 0
				// Send typing indicator
				channelTyping()
				break
			}
			// Send a message every intervalOfCharacters characters to keep the user posted on the progress
			if readChunkLength%intervalOfCharacters == 0 {
				// Send current message read so far to the user
				msgID = sendMessage(msgID, chunkToRead[:i+w])
				// Send typing indicator
				channelTyping()
			}
		}
	}
	// Make sure the last bit of the message chunk is sent.
	if chunkToRead != "" {
		sendMessage(msgID, chunkToRead)
	}

}
