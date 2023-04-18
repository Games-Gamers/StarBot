package bot

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var logg *log.Logger = log.New(os.Stdout, "Info: ", log.Ldate|log.Ltime|log.Lshortfile)

type UserFields []*discordgo.MessageEmbedField

func (uf UserFields) Len() int      { return len(uf) }
func (uf UserFields) Swap(i, j int) { uf[i], uf[j] = uf[j], uf[i] }
func (uf UserFields) Less(i, j int) bool {
	l, _ := strconv.Atoi(uf[i].Value)
	r, _ := strconv.Atoi(uf[j].Value)
	return l > r
}

var (
	BotID    string
	goBot    *discordgo.Session
	commands = []*discordgo.ApplicationCommand{
		{
			Name: "stars",
			// All commands and options must have a description
			// Commands/options without description will fail the registration
			// of the command.
			Description: "Who has how many stars?",
		},
	}

	registeredCommands = make([]*discordgo.ApplicationCommand, len(commands))

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"stars": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			fields := UserFields{}
			// Grab all the guild members for looking up usernames and nicknames
			members, err := goBot.GuildMembers(os.Getenv("GuildID"), "", 1000)
			if err != nil {
				logg.Println(err)
				return
			}
			for k, v := range stars {
				username := ""
				for _, m := range members {
					// Grab a name based on the id. Nickname if present, otherwise their username
					if m.User.ID == k {
						if m.Nick != "" {
							username = m.Nick
						} else {
							username = m.User.Username
						}
					}
				}
				if username == "" { // Deleted user or something, skip
					continue
				}
				// Should now have a username as the title and the number of stars as the value
				fields = append(fields, &discordgo.MessageEmbedField{
					Name:   username,
					Value:  strconv.Itoa(v),
					Inline: true,
				})
			}
			// Sort so the largest number of stars are at the front
			sort.Sort(fields)
			payload := &discordgo.MessageEmbed{
				Fields: fields,
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					AllowedMentions: &discordgo.MessageAllowedMentions{},
					Embeds:          []*discordgo.MessageEmbed{payload},
				},
			})
		},
	}
)
var stars = map[string]int{}

func Start() {
	var err error
	goBot, err = discordgo.New("Bot " + os.Getenv("Token"))
	if err != nil {
		logg.Println(err.Error())
		return
	}

	u, err := goBot.User("@me")
	if err != nil {
		logg.Println(err.Error())
	}

	BotID = u.ID

	goBot.AddHandler(starboardHandler)

	err = goBot.Open()
	if err != nil {
		logg.Println(err.Error())
		return
	}

	// Get all the historical stars and store them locally
	logg.Println("Populating historical stars")
	lastMessages, err := goBot.ChannelMessages(os.Getenv("StarboardChannel"), 100, "", "", "") // last 100 messages
	if err != nil {
		logg.Println(err)
		return
	}
	earliestID := "999999999999999999"
	earliestTS := time.Unix(1<<63-62135596801, 999999999)
	for len(lastMessages) > 0 {
		for _, msg := range lastMessages {
			// Get all the historical starbot messages
			if !(msg.Author.ID == "398591330806398989" || // Spud bot
				msg.Author.ID == "903055942218821682" || // Starbot
				msg.Author.ID == goBot.State.User.ID) { // Current bot
				continue
			}

			// Figure out if this message is newer than the previous for future fetching
			if msg.Timestamp.Before(earliestTS) { 
				earliestID = msg.ID
				earliestTS = msg.Timestamp
			}

			name := ""
			starCnt := 0
			// First, extract the name from the embedded fields
			emb := msg.Embeds[0]
			for _, field := range emb.Fields {
				if field.Name == "Author" {
					re := regexp.MustCompile("<@([0-9]+)>")
					name = re.FindStringSubmatch(field.Value)[1]
					break
				}
			}
			// Next, get the number of stars from the embedded Footer
			re := regexp.MustCompile("⭐ ([0-9]+)")
			starCnt, err := strconv.Atoi(re.FindStringSubmatch(emb.Footer.Text)[1])
			if err != nil {
				logg.Println(err)
				continue
			}

			// Finally, load the stars into the map
			stars[name] += starCnt
		}
		// Get the next 100 messages from before the earliest id in this batch
		lastMessages, err = goBot.ChannelMessages(os.Getenv("StarboardChannel"), 100, earliestID, "", "")
		if err != nil {
			logg.Println(err)
			break
		}
	}

	// Register the slash commands
	goBot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
	for i, v := range commands {
		cmd, err := goBot.ApplicationCommandCreate(goBot.State.User.ID, os.Getenv("GuildID"), v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	goBot.ChannelMessageSend(os.Getenv("LoggingChannel"), "I am alive!")

	logg.Println("Bot is running!")
}

func Stop() {
	for _, v := range registeredCommands {
		goBot.ApplicationCommandDelete(goBot.State.User.ID, os.Getenv("GuildID"), v.ID)
	}

	goBot.ChannelMessageSend(os.Getenv("LoggingChannel"), "Attention... I have been murdered.")
	err := goBot.Close()
	if err != nil {
		logg.Println(err.Error())
		return
	}
	logg.Println("\rBot shutting down")
}

func starboardHandler(s *discordgo.Session, mr *discordgo.MessageReactionAdd) {

	message, err := s.ChannelMessage(mr.ChannelID, mr.MessageID)
	if err != nil {
		logg.Println(err)
		return
	}

	channel, err := s.Channel(mr.ChannelID)
	if err != nil {
		logg.Println(err)
		return
	}

	user, err := s.User(message.Author.ID)
	if err != nil {
		logg.Println(err)
		return
	}

	guild, err := s.Guild(mr.GuildID)
	if err != nil {
		logg.Println(err)
		return
	}

	// Only handle messages that get the star reaction and are not by the bot itself
	if mr.MessageReaction.Emoji.Name == "⭐" && message.Author.ID != s.State.User.ID {

		// Get the number of stars on the post
		reactionCnt := 0
		for _, m := range message.Reactions {
			if m.Emoji.Name == "⭐" {
				reactionCnt = m.Count
				break
			}
		}

		payload := &discordgo.MessageEmbed{
			// Tag the server as the author
			Author: &discordgo.MessageEmbedAuthor{
				Name:    guild.Name,
				IconURL: guild.IconURL(),
			},
			// Set the thumbnail to the OPs profile picture
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: user.AvatarURL("128"),
			},
			// Set the footer to have the star emoji, the emoji count, and message ID. This allows us to search the channel to avoid duplicate posts as well
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("⭐ %d | %s", reactionCnt, mr.MessageID),
			},
			// Populate the author, channel, and link to the starred message
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Author",
					Value:  user.Mention(),
					Inline: true,
				},
				{
					Name:   "Channel",
					Value:  channel.Mention(),
					Inline: true,
				},
				{
					Name:   "Jump",
					Value:  fmt.Sprintf("[Link](https://discord.com/channels/%s/%s/%s/)", guild.ID, channel.ID, message.ID),
					Inline: true,
				},
			},
		}

		// If the starred message has text, add the text into the post
		if len(message.Content) > 0 {
			payload.Fields = append(payload.Fields, &discordgo.MessageEmbedField{
				Name:   "Message",
				Value:  message.Content,
				Inline: false,
			})
		}

		// If the starred message has an attachment (image, file, etc) then add the first attachment to the post
		if len(message.Attachments) > 0 {
			payload.Image = &discordgo.MessageEmbedImage{
				URL: message.Attachments[0].URL,
			}
		}

		// Search the starboard channel to see if this message has been starred before
		lastMessages, err := s.ChannelMessages(os.Getenv("StarboardChannel"), 100, "", "", "")
		if err != nil {
			logg.Println(err)
			return
		}
		existingMessage := ""
		for _, msg := range lastMessages {
			if msg.Author.ID == s.State.User.ID {
				if len(msg.Embeds) >= 1 {
					emb := msg.Embeds[0]
					if emb.Footer != nil && strings.Contains(emb.Footer.Text, "⭐") && strings.Contains(emb.Footer.Text, message.ID) {
						existingMessage = msg.ID
						break
					}
				}
			}
		}

		if existingMessage != "" { // If the message has had a starboard post before, just edit that post
			goBot.ChannelMessageEditComplex(&discordgo.MessageEdit{
				Embed:   payload,
				ID:      existingMessage,
				Channel: os.Getenv("StarboardChannel"),
			})
		} else if message.Author.ID != mr.UserID { // If the message hasn't been sent to the starboard before AND OP didn't star their own message, make a new post
			goBot.ChannelMessageSendComplex(os.Getenv("StarboardChannel"), &discordgo.MessageSend{Embed: payload})
		}
		// Increment their star count in our local count cache
		stars[message.Author.ID]++
	}
}
