package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Subscriber represents a Telegram user who has subscribed to property updates
type Subscriber struct {
	ChatID    int64
	FirstName string
	Username  string
}

// TelegramBot handles the Telegram bot functionality
type TelegramBot struct {
	bot         *tgbotapi.BotAPI
	subscribers map[int64]Subscriber
}

// NewTelegramBot creates a new Telegram bot instance
func NewTelegramBot(token string) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	return &TelegramBot{
		bot:         bot,
		subscribers: make(map[int64]Subscriber),
	}, nil
}

// Start starts the Telegram bot and listens for commands
func (t *TelegramBot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := t.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if !update.Message.IsCommand() {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		switch update.Message.Command() {
		case "start":
			msg.Text = "Welcome to the Property Notifier Bot! Use /subscribe to get notifications about new property listings."
		case "subscribe":
			t.handleSubscribe(update.Message, &msg)
		default:
			msg.Text = "I don't know that command"
		}

		if _, err := t.bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}

// handleSubscribe handles the /subscribe command
func (t *TelegramBot) handleSubscribe(message *tgbotapi.Message, reply *tgbotapi.MessageConfig) {
	chatID := message.Chat.ID
	
	// Check if already subscribed
	if _, exists := t.subscribers[chatID]; exists {
		reply.Text = "You are already subscribed to property notifications!"
		return
	}
	
	// Add new subscriber
	t.subscribers[chatID] = Subscriber{
		ChatID:    chatID,
		FirstName: message.From.FirstName,
		Username:  message.From.UserName,
	}
	
	reply.Text = fmt.Sprintf("Thanks for subscribing, %s! You will now receive notifications about new property listings.", message.From.FirstName)
	log.Printf("New subscriber: %s (ID: %d)", message.From.FirstName, chatID)
}

// NotifySubscribers sends a notification to all subscribers
func (t *TelegramBot) NotifySubscribers(message string) {
	for chatID := range t.subscribers {
		msg := tgbotapi.NewMessage(chatID, message)
		if _, err := t.bot.Send(msg); err != nil {
			log.Printf("Error sending notification to %d: %v", chatID, err)
		}
	}
}
