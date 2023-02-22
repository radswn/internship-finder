# Internship Finder

## Description

This is an AWS-deployed bot created to retrieve and send job (internship) offers from career sites of various companies. 
The architecture of the bot consists of two Lambda functions and an SQS queue between them.

The reasons for this application to be created:
- getting more familiar with AWS on my own
- writing some code in Go as a challenge
- looking for actual job opportunities in top companies
- it seemed like a cool resume project

## Functionality

1. Scrape data of internships from a career site
2. Pack the mentioned data into a universal array of `Offer` structs
3. Send the array to the SQS
4. Retrieve the offers by the second Lambda
5. Forward them to a Telegram webhook
6. User receives offers as separate messages with links in a Telegram chat

Handled companies:

* Apple
* Amazon

## Deploy on AWS

1. Rename the `.env.example` file to `.env`
2. Create a Telegram bot, e.g. [like that](https://medium.com/swlh/build-a-telegram-bot-in-go-in-9-minutes-e06ad38acef1)
3. Fill out the secrets
   - `BotToken` - Telegram API token of the bot you created
   - `ChatID` - ID of your chat with the bot ([how to obtain it](https://stackoverflow.com/a/50736131))
4. Create an S3 bucket in AWS manually and substitute its name in `deploy.ps1` line 8
5. Validate your AWS identity in terminal
6. Run `deploy.ps1` (goes without saying - that's a PowerShell script, so Windows-only)
