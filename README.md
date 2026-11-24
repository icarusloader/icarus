# ![](https://wsrv.nl?url=raw.githubusercontent.com/icarusloader/.github/416d97ada95d88f771b4542b41769d8241d99bad/media/icarus.svg&height=32) icarusloader 

![GitHub License](https://img.shields.io/github/license/icarusloader/icarus?color=8bd5ca) ![Discord](https://img.shields.io/discord/1386295493109481582?color=8bd5ca)

icarusloader is a modular modloader written in Go for Roblox. It is built on the principle of everything in the source code being switchable, with a simple API for implementing your own modules and mods.

## Installation

If you don't have Go installed, you should be able to get a build from the Releases page in GitHub. Running the binary should generate the necessary folders for you where the binary is, so I recommend putting it in its own folder.

If you do have Go installed, you can build it yourself. To build it, you can run the following command from the root:

```sh
go build -o ./bin/ ./src/cmd/main/cli
```
