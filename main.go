package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/rayspock/go-chatgpt-discord/handler"
	"github.com/rayspock/go-chatgpt-discord/provider"
	"github.com/rayspock/go-chatgpt-discord/ref"
	"github.com/rayspock/go-chatgpt-discord/setup"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

type config struct {
	openaiAPIKey    string
	openaiModel     string
	discordClientID string
	botToken        string
	logConfig       setup.LogConfig
}

func readConfig() config {
	return config{
		openaiModel:     os.Getenv("OPENAI_MODEL"),
		openaiAPIKey:    os.Getenv("OPENAI_API_KEY"),
		discordClientID: os.Getenv("DISCORD_CLIENT_ID"),
		botToken:        os.Getenv("DISCORD_BOT_TOKEN"),
		logConfig: setup.LogConfig{
			LogLevel: os.Getenv("LOG_LEVEL"),
		},
	}
}

func init() {
	err := godotenv.Load()
	if err != nil {
		panic(errors.Wrapf(err, "couldn't load .env file"))
	}
}

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:         handler.ApplicationCommandChat,
			Description:  "Create a new thread for conversation.",
			DMPermission: ref.Of(true),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "messages",
					Description: "Message to send",
					Required:    true,
				},
			},
		},
	}
)

func main() {
	// load configuration
	cfg := readConfig()
	setup.ConfigureLogger(cfg.logConfig)

	// create a new Discord session using the provided bot token.
	discord, err := discordgo.New("Bot " + cfg.botToken)
	if err != nil {
		log.Fatalf("error creating discord session: %v", err)
		return
	}

	// get app id
	app, err := discord.Application("@me")
	if err != nil {
		log.Fatalf("couldn't get app id: %v", err)
	}

	log.Println("adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := discord.ApplicationCommandCreate(app.ID, "", v)
		if err != nil {
			log.Panicf("cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	// configure discord handler
	chatGPTService := provider.NewChatGPTService(cfg.openaiAPIKey, cfg.openaiModel)
	discordHandler := handler.NewDiscordHandler(chatGPTService)
	discord.AddHandler(discordHandler.GetInteractionCreateHandler())
	discord.AddHandler(discordHandler.GetMessageCreateHandler())

	// open a websocket connection to discord and begin listening.
	err = discord.Open()
	if err != nil {
		log.Fatalf("error opening connection: %v", err)
		return
	}
	defer discord.Close()

	// show bot invite url
	botInviteURL := fmt.Sprintf("https://discord.com/api/oauth2/authorize?client_id=%s&permissions=%s&scope=%s",
		cfg.discordClientID, "328565073920", "bot")
	log.Infof("invite bot to your server: %s", botInviteURL)

	// wait here until ctrl-c or other term signal is received.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	log.Info("bot is now running. press ctrl-c to exit.")
	<-stop

	// remove commands
	log.Println("removing commands...")
	for _, v := range registeredCommands {
		err = discord.ApplicationCommandDelete(app.ID, "", v.ID)
		if err != nil {
			log.Panicf("cannot delete '%v' command: %v", v.Name, err)
		}
	}

	log.Println("bot is now exiting. bye!")
}
