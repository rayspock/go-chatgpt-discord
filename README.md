# GPT Discord Bot

Discord bot written in Golang that interacts with the ChatGPT API

## Table of Contents

- [Installation](#installation)
- [Local Development](#local-development)
- [Contributing](#contributing)
- [License](#license)

## Installation

To run this bot, follow these steps:

1. Go to the [releases page](https://github.com/rayspock/go-chatgpt-discord/releases) and download the binary that
   matches your operating system. Extract the binary to a directory of your choice (such as ~/go-chatgpt-discord).
2. In the same directory as the binary, create a copy of .env.example and name it .env. Fill in the appropriate
   credentials as directed.
3. Obtain a new OpenAI API key by following the link [here](https://platform.openai.com/account/api-keys). Then, fill
   in `OPENAI_API_KEY`.
4. Set up your Discord application and add a bot from
   the [Discord Developer Portal](https://discord.com/developers/applications):

    - Fill out `DISCORD_BOT_TOKEN` with your Discord bot token from the Bot settings page.
    - Copy your "Client ID" from the OAuth2 tab and fill in `DISCORD_CLIENT_ID`.

5. Run the bot by executing ./go-chatgpt-discord. A bot invitation URL will appear in the console. Copy and paste this
   URL into your web browser to add the bot to your

## Local Development

```shell
# build a go binary
$ make

# for local development
$ make run 

# run unit tests
$ make test
```

## Contributing

Welcome any kind of contribution to this repository. If you have any suggestions or ideas for improving the code
examples or best practices, please feel free to open an issue or submit a pull request.

## License

Distributed under the MIT License. See `LICENSE.txt` for more information.