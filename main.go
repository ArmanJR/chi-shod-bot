package main

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type CircularBuffer struct {
	data  [400]string
	index int
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	botAPI, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	botUsername := os.Getenv("TELEGRAM_BOT_USERNAME")
	if botUsername == "" {
		log.Fatal("TELEGRAM_BOT_USERNAME not set in .env file")
	}

	botAdminUsername := os.Getenv("TELEGRAM_BOT_ADMIN_USERNAME")
	if botAdminUsername == "" {
		log.Fatal("TELEGRAM_BOT_ADMIN_USERNAME not set in .env file")
	}

	botAdminChatIDStr := os.Getenv("TELEGRAM_BOT_ADMIN_CHAT_ID")
	if botAdminChatIDStr == "" {
		log.Fatal("TELEGRAM_BOT_ADMIN_CHAT_ID not set in .env file")
	}
	botAdminChatID, err := strconv.ParseInt(botAdminChatIDStr, 10, 64)
	if err != nil {
		log.Fatal("TELEGRAM_BOT_ADMIN_CHAT_ID is not a valid integer")
	}

	botGroupChatIDStr := os.Getenv("TELEGRAM_BOT_GROUP_CHAT_ID")
	if botGroupChatIDStr == "" {
		log.Fatal("TELEGRAM_BOT_GROUP_CHAT_ID not set in .env file")
	}
	botGroupChatID, err := strconv.ParseInt(botGroupChatIDStr, 10, 64)
	if err != nil {
		log.Fatal("TELEGRAM_BOT_GROUP_CHAT_ID is not a valid integer")
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		log.Fatal("HTTP_PORT not set in .env file")
	}

	openAiToken := os.Getenv("OPENAI_TOKEN")
	if openAiToken == "" {
		log.Fatal("OPENAI_TOKEN not set in .env file")
	}

	botAPI.Debug = true
	log.Printf("Authorized on account %s", botAPI.Self.UserName)

	updates := botAPI.ListenForWebhook("/")
	go http.ListenAndServe(":"+httpPort, nil)

	cb := &CircularBuffer{}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID

		//Uncomment this to know your PV or group chat ID
		//msg := tgbotapi.NewMessage(chatID, strconv.FormatInt(chatID, 10))
		//botAPI.Send(msg)
		//continue

		if chatID != botAdminChatID && chatID != botGroupChatID {
			msg := tgbotapi.NewMessage(chatID, "Sorry, only admin can communicate with this bot.")
			botAPI.Send(msg)
			continue
		}

		if update.Message.Text == "/start" || update.Message.Text == "/start@"+botUsername {
			msg := tgbotapi.NewMessage(chatID, "Hi! I can summarize your group chat in Persian.")
			botAPI.Send(msg)
		} else if update.Message.Text == "/chishod" || update.Message.Text == "/chishod@"+botUsername {
			if update.Message.From.UserName != botAdminUsername {
				msg := tgbotapi.NewMessage(chatID, "Sorry, this command is only available to the bot admin.")
				botAPI.Send(msg)
				continue
			}
			allMessages := cb.ConcatMessages()
			if allMessages == "" {
				msg := tgbotapi.NewMessage(chatID, "No messages yet.")
				botAPI.Send(msg)
			}
			openaiResp := OpenAIRequest(openAiToken, allMessages)
			msg := tgbotapi.NewMessage(chatID, openaiResp)
			botAPI.Send(msg)
			cb.Empty()
		} else if chatID == botGroupChatID {
			cb.AddMessage(update.Message)
		}
	}
}

// AddMessage Adds a new message to the buffer
func (cb *CircularBuffer) AddMessage(message *tgbotapi.Message) {
	if len(message.Text) <= 1 {
		return
	}
	text := message.From.FirstName
	if message.ReplyToMessage != nil {
		text = text + "(در پاسخ به " + message.ReplyToMessage.From.FirstName + ")"
	}
	text += ": " + StringReplace(message.Text)
	cb.data[cb.index] = text
	cb.index = (cb.index + 1) % 400
}

// ConcatMessages Concatenates all messages into a single string with break lines between them
func (cb *CircularBuffer) ConcatMessages() string {
	var messages []string
	messages = append(messages, "این یک مکالمه در گروه است. خلاصه ای از این مکالمه به فارسی و در حداکثر دو خط ارائه بده: ")
	for i := 0; i < 400; i++ {
		idx := (cb.index + i) % 400
		if cb.data[idx] != "" {
			messages = append(messages, cb.data[idx])
		}
	}
	return strings.Join(messages, "\n")
}

func (cb *CircularBuffer) Empty() {
	for i := range cb.data {
		cb.data[i] = ""
	}
	cb.index = 0
}

// StringReplace Replaces all emojis with an empty string
func StringReplace(s string) string {
	//var emojiRx = regexp.MustCompile(`[\x{1F600}-\x{1F6FF}|[\x{2600}-\x{26FF}]`)
	//s = emojiRx.ReplaceAllString(s, ``)
	return strings.Replace(s, "\n", " ", -1)
}

// TrimToMax Trims a string to a maximum length
func TrimToMax(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}

// OpenAIRequest Sends a request to OpenAI API and returns the response
func OpenAIRequest(token string, text string) string {
	client := openai.NewClient(token)
	text = TrimToMax(text, 8000) // Max length of a chatgpt prompt is 4,096 tokens ~ 8,000 characters

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: text,
				},
			},
		},
	)

	if err != nil {
		return fmt.Sprintf("ChatCompletion error: %v\n", err)
	}

	return resp.Choices[0].Message.Content
}
