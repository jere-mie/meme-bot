package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
	"github.com/joho/godotenv"
)

var commands = []api.CreateCommandData{
	{
		Name:        "ping",
		Description: "ping pong!",
	},
	{
		Name:        "echo",
		Description: "echo back the argument",
		Options: []discord.CommandOption{
			&discord.StringOption{
				OptionName:  "argument",
				Description: "what's echoed back",
				Required:    true,
			},
		},
	},
	{
		Name:        "meme",
		Description: "generate a meme",
		Options: []discord.CommandOption{
			&discord.StringOption{
				OptionName:  "template",
				Description: "name of the meme template",
				Required:    true,
			},
			&discord.IntegerOption{
				OptionName:  "fontsize",
				Description: "size of the font",
				Required:    true,
			},
			&discord.StringOption{
				OptionName:  "text",
				Description: "text to place on the meme",
				Required:    true,
			},
		},
	},
	{
		Name:        "memes",
		Description: "list available meme templates",
	},
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	token := os.Getenv("MEME_BOT_TOKEN")
	if token == "" {
		log.Fatalln("No $MEME_BOT_TOKEN given.")
	}

	s := state.New("Bot " + token)
	defer s.Close()
	app, err := s.CurrentApplication()
	if err != nil {
		log.Fatalln("Failed to get application ID:", err)
	}

	guildSnowflake, err := discord.ParseSnowflake(os.Getenv("DEBUG_GUILD_ID"))
	if err == nil {
		if _, err := s.BulkOverwriteGuildCommands(app.ID, discord.GuildID(guildSnowflake), commands); err != nil {
			log.Fatalln("failed to create guild command:", err)
		}
	}

	h := newHandler(state.New("Bot " + token))
	h.s.AddInteractionHandler(h)
	h.s.AddIntents(gateway.IntentGuilds)
	h.s.AddHandler(func(*gateway.ReadyEvent) {
		me, _ := h.s.Me()
		log.Println("connected to the gateway as", me.Tag())
	})

	if err := cmdroute.OverwriteCommands(h.s, commands); err != nil {
		log.Fatalln("cannot update commands:", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := h.s.Connect(ctx); err != nil {
		log.Fatalln("cannot connect:", err)
	}
}

type handler struct {
	*cmdroute.Router
	s *state.State
}

func newHandler(s *state.State) *handler {
	h := &handler{s: s}

	h.Router = cmdroute.NewRouter()
	// Automatically defer handles if they're slow.
	h.Use(cmdroute.Deferrable(s, cmdroute.DeferOpts{}))
	h.AddFunc("ping", h.cmdPing)
	h.AddFunc("echo", h.cmdEcho)
	h.AddFunc("meme", h.cmdMeme)
	h.AddFunc("memes", h.cmdMemes)

	return h
}

func (h *handler) cmdPing(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	return &api.InteractionResponseData{
		Content: option.NewNullableString("Pong!"),
	}
}

func (h *handler) cmdEcho(ctx context.Context, data cmdroute.CommandData) *api.InteractionResponseData {
	var options struct {
		Arg string `discord:"argument"`
	}

	if err := data.Options.Unmarshal(&options); err != nil {
		return errorResponse(err)
	}

	return &api.InteractionResponseData{
		Content:         option.NewNullableString(options.Arg),
		AllowedMentions: &api.AllowedMentions{}, // don't mention anyone
	}
}

func (h *handler) cmdMeme(ctx context.Context, data cmdroute.CommandData) *api.InteractionResponseData {
	var options struct {
		Template string `discord:"template"`
		FontSize int    `discord:"fontsize"`
		Text     string `discord:"text"`
	}

	if err := data.Options.Unmarshal(&options); err != nil {
		return errorResponse(err)
	}

	if options.FontSize > 75 || options.FontSize < 8 {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("Font size must be an integer between 8 and 75"),
		}
	}

	imagePath := fmt.Sprintf("./assets/img/%s.png", options.Template)
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return &api.InteractionResponseData{
			Content: option.NewNullableString(fmt.Sprintf("Meme template `%s` not found!", options.Template)),
		}
	}

	buf, err := drawImage(options.Template, float64(options.FontSize), options.Text)
	if err != nil {
		return errorResponse(err)
	}

	return &api.InteractionResponseData{
		Content: option.NewNullableString("Here's your meme!"),
		Files: []sendpart.File{
			{
				Name:   "meme.png",
				Reader: buf,
			},
		},
	}
}

func (h *handler) cmdMemes(ctx context.Context, cmd cmdroute.CommandData) *api.InteractionResponseData {
	files, err := os.ReadDir("./assets/img")
	if err != nil {
		return errorResponse(err)
	}

	var memes []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".png") {
			memes = append(memes, strings.TrimSuffix(file.Name(), ".png"))
		}
	}

	// Check if memes.png exists
	memeImagePath := "./assets/img/memes.png"
	if _, err := os.Stat(memeImagePath); os.IsNotExist(err) {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("Available meme templates:\n```" + strings.Join(memes, ", ") + "```"),
		}
	}

	// Open the image file
	memeImage, err := os.Open(memeImagePath)
	if err != nil {
		return errorResponse(err)
	}
	defer memeImage.Close()

	// Read the image file into a buffer
	imageData, err := io.ReadAll(memeImage)
	if err != nil {
		return errorResponse(err)
	}
	imageBuffer := strings.NewReader(string(imageData))

	return &api.InteractionResponseData{
		Content: option.NewNullableString("Available meme templates:\n```" + strings.Join(memes, ", ") + "```"),
		Files: []sendpart.File{
			{
				Name:   "memes.png",
				Reader: imageBuffer,
			},
		},
	}
}

func errorResponse(err error) *api.InteractionResponseData {
	return &api.InteractionResponseData{
		Content:         option.NewNullableString("**Error:** " + err.Error()),
		Flags:           discord.EphemeralMessage,
		AllowedMentions: &api.AllowedMentions{ /* none */ },
	}
}
