# Pararius Property Notification Bot

A Telegram bot that monitors [Pararius](https://www.pararius.nl) for new rental property listings and sends notifications to subscribers.

## Features

- Automatically checks for new rental properties every 10 minutes
- Sends Telegram notifications with property details to subscribers
- Supports user subscription via the `/subscribe` command
- Saves property data as JSON files for deduplication
- Customizable search URL for different locations and price ranges

## Requirements

- Go 1.16 or higher
- A Telegram Bot API token (obtained from [@BotFather](https://t.me/botfather))
- Internet access to fetch property listings

## Installation

1. Clone this repository:
   ```bash
   git clone https://github.com/yourusername/pararius-property-bot.git
   cd pararius-property-bot
   ```

2. Install dependencies:
   ```bash
   go get github.com/PuerkitoBio/goquery
   go get github.com/go-telegram-bot-api/telegram-bot-api/v5
   ```

## Usage

Run the bot with the following command:

```bash
go run *.go -output ./properties -token YOUR_TELEGRAM_BOT_TOKEN
```

### Command-line Arguments

| Flag | Description | Default | Required |
|------|-------------|---------|----------|
| `-output` | Directory to save property JSON files | none | Yes |
| `-token` | Telegram Bot API token | none | Yes |
| `-url` | URL to search for properties | `https://www.pararius.nl/huurwoningen/utrecht/1000-2500/50m2` | No |

### Example URLs

- Utrecht, max €2500, min 50m²: `https://www.pararius.nl/huurwoningen/utrecht/1000-2500/50m2`
- Amsterdam, €1000-€1500: `https://www.pararius.nl/huurwoningen/amsterdam/1000-1500`
- Rotterdam, 2+ bedrooms: `https://www.pararius.nl/huurwoningen/rotterdam/2-slaapkamers`

## Telegram Bot Commands

- `/start` - Display welcome message
- `/subscribe` - Subscribe to property notifications

## How It Works

1. The bot periodically fetches the property listings from Pararius
2. It compares the listings with previously saved properties
3. New properties are saved as JSON files in the specified output directory
4. Subscribers receive Telegram notifications about new properties

## License

MIT
