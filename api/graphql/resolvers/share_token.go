package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.64

import (
	"context"
	"errors"
	"fmt"
	"time"

	api "github.com/photoview/photoview/api/graphql"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/models/actions"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ShareAlbum is the resolver for the shareAlbum field.
func (r *mutationResolver) ShareAlbum(ctx context.Context, albumID int, expire *time.Time, password *string) (*models.ShareToken, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	return actions.AddAlbumShare(r.DB(ctx), user, albumID, expire, password)
}

// ShareMedia is the resolver for the shareMedia field.
func (r *mutationResolver) ShareMedia(ctx context.Context, mediaID int, expire *time.Time, password *string) (*models.ShareToken, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	return actions.AddMediaShare(r.DB(ctx), user, mediaID, expire, password)
}

// DeleteShareToken is the resolver for the deleteShareToken field.
func (r *mutationResolver) DeleteShareToken(ctx context.Context, token string) (*models.ShareToken, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	return actions.DeleteShareToken(r.DB(ctx), user.ID, token)
}

// ProtectShareToken is the resolver for the protectShareToken field.
func (r *mutationResolver) ProtectShareToken(ctx context.Context, token string, password *string) (*models.ShareToken, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	return actions.ProtectShareToken(r.DB(ctx), user.ID, token, password)
}

// ShareToken is the resolver for the shareToken field.
func (r *queryResolver) ShareToken(ctx context.Context, credentials models.ShareTokenCredentials) (*models.ShareToken, error) {
	var token models.ShareToken
	if err := r.DB(ctx).Preload(clause.Associations).Where("value = ?", credentials.Token).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("share not found")
		} else {
			return nil, fmt.Errorf("failed to get share token from database: %w", err)
		}
	}

	if token.Password != nil {
		if err := bcrypt.CompareHashAndPassword([]byte(*token.Password), []byte(*credentials.Password)); err != nil {
			if err == bcrypt.ErrMismatchedHashAndPassword {
				return nil, errors.New("unauthorized")
			} else {
				return nil, fmt.Errorf("failed to compare token password hashes: %w", err)
			}
		}
	}

	return &token, nil
}

// ShareTokenValidatePassword is the resolver for the shareTokenValidatePassword field.
func (r *queryResolver) ShareTokenValidatePassword(ctx context.Context, credentials models.ShareTokenCredentials) (bool, error) {
	var token models.ShareToken
	if err := r.DB(ctx).Where("value = ?", credentials.Token).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("share not found")
		} else {
			return false, fmt.Errorf("failed to get share token from database: %w", err)
		}
	}

	if token.Password == nil {
		return true, nil
	}

	if credentials.Password == nil {
		return false, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*token.Password), []byte(*credentials.Password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		} else {
			return false, fmt.Errorf("could not compare token password hashes: %w", err)
		}
	}

	return true, nil
}

// HasPassword is the resolver for the hasPassword field.
func (r *shareTokenResolver) HasPassword(ctx context.Context, obj *models.ShareToken) (bool, error) {
	hasPassword := obj.Password != nil
	return hasPassword, nil
}

// ShareToken returns api.ShareTokenResolver implementation.
func (r *Resolver) ShareToken() api.ShareTokenResolver { return &shareTokenResolver{r} }

type shareTokenResolver struct{ *Resolver }
