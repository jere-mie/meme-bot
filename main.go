package main

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"log"
	"os"
	"os/signal"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
	"github.com/fogleman/gg"
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

	imagePath := fmt.Sprintf("./assets/img/%s.png", options.Template)
	fontPath := "./assets/fonts/Anton-Regular.ttf"

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return &api.InteractionResponseData{
			Content: option.NewNullableString("Template not found"),
		}
	}

	imgBytes, err := os.ReadFile(imagePath)
	if err != nil {
		return errorResponse(err)
	}

	img, err := png.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return errorResponse(err)
	}

	dc := gg.NewContextForImage(img)
	if err := dc.LoadFontFace(fontPath, float64(options.FontSize)); err != nil {
		return errorResponse(err)
	}

	// Set font color to black
	dc.SetRGB(0, 0, 0)

	maxWidth := float64(dc.Width()) - 20 // Padding from edges
	dc.DrawStringWrapped(options.Text, 10, 10, 0, 0, maxWidth, 1.5, gg.AlignCenter)

	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return errorResponse(err)
	}

	return &api.InteractionResponseData{
		Content: option.NewNullableString("Here's your meme!"),
		Files: []sendpart.File{
			{
				Name:   "meme.png",
				Reader: &buf,
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
