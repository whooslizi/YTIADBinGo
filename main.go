package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	BOT_TOKEN      = os.Getenv("BOT_TOKEN")
	CHANNEL_ID     = os.Getenv("CHANNEL_ID")
	ROLE_ID        = os.Getenv("ROLE_ID")
	MESSAGE        = "Reacted with 🚽 to get *slightly* access to the server ||if you want full access, message <@763017830215319572>||"
	CARD_COMMAND   = os.Getenv("CARD_COMMAND")
	BG_PATH        = "greet_member/image.webp"
	FONT_PATH      = "greet_member/Roboto-Bold.ttf"
	TOILET_EMOJI   = "🚽"
	sentMessageID  string
)

func generateGreetingCard(name, disc, count string) ([]byte, error) {
	cmd := exec.Command(
		"python3",
		CARD_COMMAND,
		name,
		disc,
		count,
		BG_PATH,
		FONT_PATH,
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.Bytes(), err
}

func main() {
	dg, err := discordgo.New("Bot " + BOT_TOKEN)
	if err != nil {
		log.Fatal(err)
	}

	dg.Identify.Intents =
		discordgo.IntentsGuilds |
			discordgo.IntentsGuildMessages |
			discordgo.IntentsGuildMessageReactions |
			discordgo.IntentsGuildMembers

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		msg, err := s.ChannelMessageSend(CHANNEL_ID, MESSAGE)
		if err != nil {
			return
		}
		sentMessageID = msg.ID
		s.MessageReactionAdd(CHANNEL_ID, msg.ID, TOILET_EMOJI)
	})

	dg.AddHandler(func(s *discordgo.Session, e *discordgo.MessageReactionAdd) {
		if e.ChannelID != CHANNEL_ID {
			return
		}
		if e.Emoji.Name != TOILET_EMOJI {
			return
		}
		if e.UserID == s.State.User.ID {
			return
		}
		s.GuildMemberRoleAdd(e.GuildID, e.UserID, ROLE_ID)
	})

	dg.AddHandler(func(s *discordgo.Session, e *discordgo.MessageReactionRemove) {
		if e.ChannelID != CHANNEL_ID {
			return
		}
		if e.Emoji.Name != TOILET_EMOJI {
			return
		}
		s.GuildMemberRoleRemove(e.GuildID, e.UserID, ROLE_ID)
	})

	dg.AddHandler(func(s *discordgo.Session, e *discordgo.GuildMemberAdd) {
		guild, err := s.State.Guild(e.GuildID)
		if err != nil {
			return
		}

		card, err := generateGreetingCard(
			e.User.Username,
			e.User.Discriminator,
			string(rune(len(guild.Members))),
		)
		if err != nil {
			return
		}

		s.ChannelFileSend(CHANNEL_ID, "welcome.png", bytes.NewReader(card))
	})

	err = dg.Open()
	if err != nil {
		log.Fatal(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	dg.Close()
}
