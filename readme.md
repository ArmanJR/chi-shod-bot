# Chi Shod Bot

Telegram bot for TL;DR usage in a group chat

## What is this

This is the source of a Telegram bot that sits in a group, listens to all the conversations, and generates a summary of recent messages using OpenAI's ChatGPT API.

The messages will be stored on the memory of the server using **Circular Buffer**.

*Chi Shod?* means *What Happened?* in Persian.

## What you should consider

1. Retrieving past messages of a group from Telegram's API is not possible, so the only solution is to store every text message that someone sends to the group. Also, the messages will be sent to OpenAI's servers. Therefore, the privacy of users and the group chats are in danger.

2. Because we are combining all the messages to later summarize, the tokens for each ChatGPT request will be large, and as a result, [expensive](https://openai.com/pricing).

These two main reasons plus the cost of a server to host the bot, make it extremely inefficient to launch a bot just to summarize messages that might be used by a lazy/stalker group member who wants to know what people talked about when he/she was asleep. Still not convinced? continue.

## Getting started

### Bot

1. Create a bot via [BotFather](https://telegram.me/BotFather).
2. Set the Group Privacy to `disabled`.
3. Define a custom command of `/chishod` to use in the group.
4. Add the bot to your group.
5. Set the webhook of the bot to your server which this code is running on:
`https://api.telegram.org/bot[TOKEN]/setWebhook?url=[URL]&drop_pending_updates=True`

### Server

Clone the code, rename `.env.example` to `.env`, and fill the variables:

- `TELEGRAM_BOT_TOKEN`: Token of the bot taken from BotFather
- `TELEGRAM_BOT_USERNAME`: Username that you set for the bot
- `TELEGRAM_BOT_ADMIN_USERNAME`: Your Telegram username
- `TELEGRAM_BOT_ADMIN_CHAT_ID`: Your PV chat id (get from [ChatID bot](https://t.me/chat_id_echo_bot))
- `TELEGRAM_GROUP_CHAT_ID`: Your group chat id (can get directly from the code)
- `HTTP_PORT`: The HTTP port
- `OPENAI_TOKEN`: Your OpenAI token (get from [OpenAI dashbord](https://platform.openai.com/account/api-keys))

Run the server via `go run main.go`.

### Continuous deployment

This source comes with a Github actions workflow `.github/workflows/deploy.yml` which will update the code on your server with every push to the main branch. Before using it remember to add your server's secrets (`SERVER_SSH_KEY`, `SERVER_IP`, `SERVER_USERNAME`, `SERVER_DIR`, `LOGFILE_DIR`) on your repository.

## Contribution

At the moment, this project is far from being safe and efficient to be used. If you have ideas for solving the considerations that were mentioned before, you are more than welcome to contribute.
