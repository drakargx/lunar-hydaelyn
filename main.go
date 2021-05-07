package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/drakargx/lunar-hydaelyn/lunar"
)

var (
	inkscape chan bool
	dg       *discordgo.Session
	sc       chan os.Signal
)

func main() {

	// Create a new Discord session using the provided bot token.
	Token := os.Getenv("LUNAR_HYDAELYN_BOT_TOKEN")
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	inkscape = make(chan bool, 1)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGTERM, syscall.SIGTERM, os.Interrupt)
	<-sc
	//go CloseProgram()

	// for {
	// 	func() {}()
	// 	<-inkscape
	// 	writeOutputPng()
	// }
	// go func() {
	// 	for {
	// 		select {
	// 		case <-sc:
	// 			dg.Close()
	// 			os.Exit(0)
	// 		case <-inkscape:
	// 			writeOutputPng()
	// 		}
	// 	}
	// }()
	//<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func CloseProgram() {
	<-sc
	dg.Close()
	os.Exit(0)
}

func writeOutputPng() {

	cmd := exec.Command("cmd", "/c", "inkscape --export-type=\"png\" Output.svg")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		//CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
	err := cmd.Run()

	if err != nil {
		log.Fatal(err)
	}

	inkscape <- true
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if strings.Contains(m.Content, "fflogs.com/reports/") {

		reportID := ""
		segments := strings.Split(m.Content, "/")
		for i, value := range segments {
			if value == "reports" {
				reportID = segments[i+1]
				break
			}
		}

		if reportID != "" {
			client := lunar.NewFFLogsClient()

			fight, _ := client.GetLastFightInfo(reportID)
			if fight == nil {
				panic("FIght is nil")
			}
			resp := client.GrabReportInfo(reportID, *fight)
			lunar.GenerateOutputPng(resp)

			f, err := os.Open("Output.png")
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			s.ChannelFileSend(m.ChannelID, "Output.png", f)

		}

	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}
