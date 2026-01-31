package authorizer

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// Self is a simple Authorizer that allows the operation if and only if the user requestion the operation is the owner
// of the synth.
type Self struct{}

var _ Authorizer = (*Self)(nil)

// Authorized returns if the given user is the user that owns the synth.
func (Self) Authorized(_ context.Context, owner string, u *discordgo.User) (bool, error) {
	return owner == u.ID, nil
}
