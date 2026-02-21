package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Synth represents a Synth bot owned by a particular Discord user in the database.
type Synth struct {
	ID            uint64 `gorm:"primary_key;auto_increment"`
	DiscordUserID string `gorm:"unique;not null"`
	ApplicationID string `gorm:"not null"`
	Token         string `gorm:"not null"`
	Enabled       bool   `gorm:"not null"`
	AllowLogging  bool   `gorm:"not null;default:false"`

	CreatedAt time.Time
	UpdatedAt time.Time

	db *DB
}

func (db *DB) InsertSynth(ctx context.Context, userID, appID, token string) error {
	err := gorm.G[Synth](db.g).Create(ctx, &Synth{
		DiscordUserID: userID,
		ApplicationID: appID,
		Token:         token,
		Enabled:       true,
		AllowLogging:  false,
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return fmt.Errorf("%w: %v", ErrAlreadyExists, err)
	}
	return err
}

func (db *DB) GetSynth(ctx context.Context, userID string) (*Synth, error) {
	t, err := gorm.G[Synth](db.g).Where("discord_user_id = ?", userID).Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading Synth: %w", err)
	} else if len(t) == 0 {
		return nil, ErrNotFound
	} else if len(t) != 1 {
		// database unique constraint should prevent this from ever happening
		log.Ctx(ctx).Panic().Str("user_id", userID).Msg("Found more than one Synth for user")
	}

	s := &t[0]
	s.db = db
	return s, nil
}

// GetEnabledSynths gets all enabled Synths. TODO pagination
func (db *DB) GetEnabledSynths(ctx context.Context) ([]*Synth, error) {
	synths, err := gorm.G[Synth](db.g).Find(ctx)
	if err != nil {
		return nil, err
	}

	ret := make([]*Synth, 0, len(synths))
	for _, synth := range synths {
		s := &synth
		s.db = db
		ret = append(ret, s)
	}
	return ret, nil
}

func (s *Synth) Save(ctx context.Context) error {
	_, err := gorm.G[Synth](s.db.g).
		Where("id = ?", s.ID).
		Select("*").
		Updates(ctx, *s)
	return err
}
