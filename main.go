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

	discord, err := discordgo.New("Bot " + "NjIzOTgyOTc2OTU4NDY0MDA2.Xei_TA.27ZAAHIpRbInxPCnfnZ36Wva5YA")

	if err != nil {
		log.Fatal("Unable to start discord bot ", err)
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	discord.AddHandler(messageCreate)
	discord.AddHandler(messageReactionAdd)

	// Open a websocket connection to Discord and begin listening.
	err = discord.Open()
	if err != nil {
		log.Fatal("Unable to open websocket connection ", err)
	}

	// Cleanly close down the Discord session.
	defer discord.Close()

	log.Println("I'm running!")

	// Wait here until CTRL-C or other term signal is received.
	log.Println("Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages posted by **non** bot
	//if m.Author.Bot == false {
	//	return
	//}

	isDM, err := isDM(s, m.Message)
	if err != nil {
		log.Fatal(err)
	}

	if m.Author.Bot && !isDM {
		err := s.MessageReactionAdd(m.Message.ChannelID, m.Message.ID, "⭐")
		if err != nil {
			log.Fatal("Unable to add reaction ", err)
		}
	}

	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Pong!")
		if err != nil {
			log.Fatal("Unable to reply ", err)
		}
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Ping!")
		if err != nil {
			log.Fatal("Unable to reply ", err)
		}
	}
}

func messageReactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	log.Println(m.MessageReaction.Emoji.Name)

	message, err := s.ChannelMessage(m.ChannelID, m.MessageID)
	if err != nil {
		log.Fatal(err)
	}

	isDM, err := isDM(s, message)
	if err != nil {
		log.Fatal(err)
	}

	user, err := s.User(m.MessageReaction.UserID)
	if err != nil {
		log.Fatal(err)
	}

	if m.MessageReaction.Emoji.Name == "⭐" && user.Bot == false && !isDM {

		user, err := s.User(m.MessageReaction.UserID)
		if err != nil {
			log.Fatal(err)
		}

		dmChannel, err := s.UserChannelCreate(user.ID)
		if err != nil {
			log.Fatal(err)
		}

		message, err := s.ChannelMessageSend(dmChannel.ID, message.Content)
		if err != nil {
			log.Fatal(err)
		}

		err = s.MessageReactionAdd(message.ChannelID, message.ID, "❌")
		if err != nil {
			log.Fatal(err)
		}

		for _, embed := range message.Embeds {
			// Send message to user with embed content
			message, err = s.ChannelMessageSendEmbed(dmChannel.ID, embed)
			if err != nil {
				log.Fatal(err)
			}
			// Add reaction
			err = s.MessageReactionAdd(message.ChannelID, message.ID, "❌")
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	if m.MessageReaction.Emoji.Name == "❌" && user.Bot == false && isDM {
		err := s.ChannelMessageDelete(m.ChannelID, m.MessageID)
		if err != nil {
			log.Fatal(err)
		}
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
