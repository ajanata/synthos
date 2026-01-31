package authorizer

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

type Authorizer interface {
	Authorized(ctx context.Context, owner string, u *discordgo.User) (bool, error)
}
