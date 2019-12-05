package main

import (
	"flag"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	Token string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {

	discord, err := discordgo.New("Bot " + Token)
	check(err)

	// Register the messageCreate func as a callback for MessageCreate events.
	discord.AddHandler(messageCreate)
	discord.AddHandler(messageReactionAdd)

	// Open a websocket connection to Discord and begin listening.
	err = discord.Open()
	check(err)

	// Cleanly close down the Discord session.
	defer discord.Close()

	log.Println("I'm running! Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages posted by **non** bot
	if m.Author.Bot == false {
		return
	}

	isDM, err := isDM(s, m.Message)
	check(err)

	if m.Author.Bot && !isDM {
		err := s.MessageReactionAdd(m.Message.ChannelID, m.Message.ID, "⭐")
		check(err)
	}
}

func messageReactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	message, err := s.ChannelMessage(m.ChannelID, m.MessageID)
	check(err)

	isDM, err := isDM(s, message)
	check(err)

	user, err := s.User(m.MessageReaction.UserID)
	check(err)

	if m.MessageReaction.Emoji.Name == "⭐" && user.Bot == false && !isDM {

		user, err := s.User(m.MessageReaction.UserID)
		check(err)

		dmChannel, err := s.UserChannelCreate(user.ID)
		check(err)

		message, err := s.ChannelMessageSend(dmChannel.ID, message.Content)
		check(err)

		err = s.MessageReactionAdd(message.ChannelID, message.ID, "❌")
		check(err)

		for _, embed := range message.Embeds {
			// Send message to user with embed content
			message, err = s.ChannelMessageSendEmbed(dmChannel.ID, embed)
			check(err)
			// Add reaction
			err = s.MessageReactionAdd(message.ChannelID, message.ID, "❌")
			check(err)
		}
	}

	if m.MessageReaction.Emoji.Name == "❌" && user.Bot == false && isDM {
		err := s.ChannelMessageDelete(m.ChannelID, m.MessageID)
		check(err)
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// isDM returns true if a message comes from a DM channel
func isDM(s *discordgo.Session, m *discordgo.Message) (bool, error) {
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		if channel, err = s.Channel(m.ChannelID); err != nil {
			return false, err
		}
	}

	return channel.Type == discordgo.ChannelTypeDM, nil
}
