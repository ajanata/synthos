# SynthOS
A profile play and chat rule enforcement system for Discord.

This system has an orchestration bot that users interact with to create their own user bot,
and will then run a bot for each user.
Users have to create their own Discord developer applications to use this bot.

## Setup

### Orchestration Bot
Create a Discord application for the orchestration bot. Important settings:
* On the Installation tab:
  * Disable User Install
  * Normally, set the Install Link to None, but...
  * If you wish the bot to be discoverable in your community, you can add it to your guild,
    and you'll have to give it "bot" scope in the Default Install Settings,
    but you don't need to give it any permissions.
* On the Bot tab:
  * Disable Public Bot
  * Click Reset Token, and copy the new token into synthos.toml.

And that should be all you have to do for the orchestration bot.

### User Bots
Have your users execute the `/setup start` command in a DM with the orchestration bot.
It will provide instructions on how to set up their bot.



# TODOs

* opening configure for a fresh bot/fresh guild has blanks for display name and an invalid interaction error, but recovers fine
