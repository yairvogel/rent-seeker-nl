package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

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

	telegramBot := &TelegramBot{
		bot:         bot,
		subscribers: make(map[int64]Subscriber),
	}

	// Load subscribers from disk
	if err := telegramBot.loadSubscribers(); err != nil {
		log.Printf("Warning: Could not load subscribers: %v", err)
	}

	return telegramBot, nil
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
		case "unsubscribe":
			t.handleUnsubscribe(update.Message, &msg)
		default:
			msg.Text = "I don't know that command"
		}

		if _, err := t.bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}

func (t *TelegramBot) HasSubscribers() bool {
	return len(t.subscribers) > 0
}

// handleUnsubscribe handles the /unsubscribe command
func (t *TelegramBot) handleUnsubscribe(message *tgbotapi.Message, reply *tgbotapi.MessageConfig) {
	chatID := message.Chat.ID

	// Check if subscribed
	if _, exists := t.subscribers[chatID]; !exists {
		reply.Text = "You are not currently subscribed to property notifications."
		return
	}

	// Remove subscriber
	delete(t.subscribers, chatID)

	// Save subscribers to disk
	if err := t.saveSubscribers(); err != nil {
		log.Printf("Error saving subscribers: %v", err)
	}

	reply.Text = "You have been unsubscribed from property notifications. Use /subscribe to subscribe again."
	log.Printf("Unsubscribed user: %d", chatID)
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

	// Save subscribers to disk
	if err := t.saveSubscribers(); err != nil {
		log.Printf("Error saving subscribers: %v", err)
	}

	reply.Text = fmt.Sprintf("Thanks for subscribing, %s! You will now receive notifications about new property listings.", message.From.FirstName)
	log.Printf("New subscriber: %s (ID: %d)", message.From.FirstName, chatID)
}

// NotifySubscribers sends a notification to all subscribers
func (t *TelegramBot) NotifySubscribers(message string) {
	for chatID := range t.subscribers {
		msg := tgbotapi.NewMessage(chatID, message)
		msg.ParseMode = "Markdown"
		if _, err := t.bot.Send(msg); err != nil {
			log.Printf("Error sending notification to %d: %v", chatID, err)
		}
	}
}

// saveSubscribers saves the current subscribers to a JSON file
func (t *TelegramBot) saveSubscribers() error {
	// Create subscribers directory if it doesn't exist
	subscribersDir := "subscribers"
	if err := os.MkdirAll(subscribersDir, 0755); err != nil {
		return fmt.Errorf("failed to create subscribers directory: %v", err)
	}

	// Marshal subscribers to JSON
	data, err := json.MarshalIndent(t.subscribers, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal subscribers: %v", err)
	}

	// Write to file
	filePath := filepath.Join(subscribersDir, "subscribers.json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write subscribers file: %v", err)
	}

	log.Printf("Saved %d subscribers to disk", len(t.subscribers))
	return nil
}

// loadSubscribers loads subscribers from a JSON file
func (t *TelegramBot) loadSubscribers() error {
	filePath := filepath.Join("subscribers", "subscribers.json")
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Println("No subscribers file found, starting with empty subscribers list")
		return nil
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read subscribers file: %v", err)
	}

	// Unmarshal JSON
	if err := json.Unmarshal(data, &t.subscribers); err != nil {
		return fmt.Errorf("failed to unmarshal subscribers: %v", err)
	}

	log.Printf("Loaded %d subscribers from disk", len(t.subscribers))
	return nil
}
