package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/Games-Gamers/StarBot/config"
)

var (
	BotID string
	goBot *discordgo.Session
)

func Start() {
	var err error
	goBot, err = discordgo.New("Bot " + config.Token)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	u, err := goBot.User("@me")
	if err != nil {
		fmt.Println(err.Error())
	}

	BotID = u.ID

	// goBot.AddHandler(famHandler)
	goBot.AddHandler(starboardHandler)

	err = goBot.Open()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	goBot.ChannelMessageSend(config.LoggingChannel, "I am alive!")
	goBot.UserUpdateStatus(discordgo.StatusOnline)

	fmt.Println("Bot is running!")
}

func Stop() {
	goBot.ChannelMessageSend(config.LoggingChannel, "Attention... I have been murdered.")
	err := goBot.Close()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("\rBot shutting down")
}

func famHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	content := strings.ReplaceAll(m.Content, config.BotPrefix+".", "")

	if strings.HasPrefix(m.Content, config.BotPrefix) {
		if m.Author.ID == BotID {
			return
		}

		if content == "fam" {
			_, _ = s.ChannelMessageSend(m.ChannelID, "<:fam:848761741102153739>")
		}
	}

	if strings.Contains(strings.ToLower(content), "fam") {
		err := s.MessageReactionAdd(m.ChannelID, m.ID, "FAM:848761741102153739")
		if err != nil {
			fmt.Println(err)
		}
	}
}

func starboardHandler(s *discordgo.Session, mr *discordgo.MessageReactionAdd) {

	message, err := s.ChannelMessage(mr.ChannelID, mr.MessageID)
	if err != nil {
		fmt.Println(err)
		return
	}

	channel, err := s.Channel(mr.ChannelID)
	if err != nil {
		fmt.Println(err)
		return
	}

	user, err := s.User(message.Author.ID)
	if err != nil {
		fmt.Println(err)
		return
	}

	guild, err := s.Guild(mr.GuildID)
	if err != nil {
		fmt.Println(err)
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
			// Set the timestamp to OPs message creation ts
			Timestamp: string(message.Timestamp),
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
		lastMessages, err := s.ChannelMessages(config.StarboardChannel, 100, "", "", "")
		if err != nil {
			fmt.Println(err)
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
				Channel: config.StarboardChannel,
			})
		} else if message.Author.ID != mr.UserID { // If the message hasn't been sent to the starboard before AND OP didn't star their own message, make a new post
			goBot.ChannelMessageSendComplex(config.StarboardChannel, &discordgo.MessageSend{Embed: payload})
		}
	}
}
