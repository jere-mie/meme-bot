# Meme Bot

Discord Bot for making memes

## Getting Set Up

### Building/Running From Source 

- Ensure you have Go v1.22.3+ installed
- Install dependencies with the command `go mod tidy`
- Specify your bot token in a new `.env` file  
- Run the application with `go run .` (don't forget the `.`)

#### Air

You can use [Air](https://github.com/air-verse/air) for live reloading during development. Simply install Air with the following command:

```sh
go install github.com/air-verse/air@latest
```

and then you can type `air` in your terminal to run the application.

<!--

>> keeping this commented out until I actually create some releases

## Downloading Pre-Built Binaries

You can find pre-built memebot binaries for Windows, Linux, and MacOS on the meme-bot repo's [releases page](https://github.com/jere-mie/meme-bot/releases/latest).

If you prefer downloading via the cli, use one of the following commands below:

```sh
# Windows amd64
irm -Uri https://github.com/jere-mie/meme-bot/releases/latest/download/memebot_windows_amd64.exe -O memebot.exe

# Linux amd64
curl -L https://github.com/jere-mie/meme-bot/releases/latest/download/memebot_linux_amd64 -o memebot && chmod +x memebot

# MacOS arm64 (Apple Silicon)
curl -L https://github.com/jere-mie/meme-bot/releases/latest/download/memebot_darwin_arm64 -o memebot && chmod +x memebot

# MacOS amd64 (Intel)
curl -L https://github.com/jere-mie/meme-bot/releases/latest/download/memebot_darwin_amd64 -o memebot && chmod +x memebot
```

-->
