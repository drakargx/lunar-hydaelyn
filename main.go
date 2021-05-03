// package main

// import (
// 	"flag"
// 	"fmt"
// 	"os"
// 	"os/signal"
// 	"strings"
// 	"syscall"

// 	"github.com/bwmarrin/discordgo"
// 	"github.com/drakargx/lunar-hydaelyn/lunar"
// )

// // Variables used for command line parameters
// var (
// 	Token string
// )

// func init() {

// 	flag.StringVar(&Token, "t", "", "Bot Token")
// 	flag.Parse()
// }

// func main() {
//
// 	// Create a new Discord session using the provided bot token.
// 	dg, err := discordgo.New("Bot " + Token)
// 	if err != nil {
// 		fmt.Println("error creating Discord session,", err)
// 		return
// 	}

// 	// Register the messageCreate func as a callback for MessageCreate events.
// 	dg.AddHandler(messageCreate)

// 	// In this example, we only care about receiving message events.
// 	dg.Identify.Intents = discordgo.IntentsGuildMessages

// 	// Open a websocket connection to Discord and begin listening.
// 	err = dg.Open()
// 	if err != nil {
// 		fmt.Println("error opening connection,", err)
// 		return
// 	}

// 	// Wait here until CTRL-C or other term signal is received.
// 	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
// 	sc := make(chan os.Signal, 1)
// 	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
// 	<-sc

// 	// Cleanly close down the Discord session.
// 	dg.Close()
// }

// // This function will be called (due to AddHandler above) every time a new
// // message is created on any channel that the authenticated bot has access to.
// func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

// 	// Ignore all messages created by the bot itself
// 	// This isn't required in this specific example but it's a good practice.
// 	if m.Author.ID == s.State.User.ID {
// 		return
// 	}
// 	// If the message is "ping" reply with "Pong!"
// 	if strings.Contains(m.Content, "fflogs.com/reports/") {

// 		reportID := ""
// 		segments := strings.Split(m.Content, "/")
// 		for i, value := range segments {
// 			if value == "reports" {
// 				reportID = segments[i+1]
// 				break
// 			}
// 		}

// 		if reportID != "" {
// 			client := lunar.NewFFLogsClient()

// 			fight, _ := client.GetLastFightInfo(reportID)
// 			if fight != nil {
// 				s.ChannelMessageSend(m.ChannelID, fight.BossName)
// 			} else {
// 				s.ChannelMessageSend(m.ChannelID, "nil fight")
// 			}

// 		}

// 	}

// 	// If the message is "pong" reply with "Ping!"
// 	if m.Content == "pong" {
// 		s.ChannelMessageSend(m.ChannelID, "Ping!")
// 	}
// }

package main

import (
	//"fmt"

	"github.com/drakargx/lunar-hydaelyn/lunar"
)

func main() {

	client := lunar.NewFFLogsClient()

	report := "7TZLVgR29PGDfrMH"
	fight, _ := client.GetLastFightInfo(report)

	lunar.GenerateOutputPng("", lunar.NewFFLogsClient().GrabReportInfo(report, *fight))

}
