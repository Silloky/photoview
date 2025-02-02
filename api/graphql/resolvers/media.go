package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.64

import (
	"context"
	"fmt"

	"github.com/photoview/photoview/api/dataloader"
	api "github.com/photoview/photoview/api/graphql"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/models/actions"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Thumbnail is the resolver for the thumbnail field.
func (r *mediaResolver) Thumbnail(ctx context.Context, obj *models.Media) (*models.MediaURL, error) {
	return dataloader.For(ctx).MediaThumbnail.Load(obj.ID)
}

// HighRes is the resolver for the highRes field.
func (r *mediaResolver) HighRes(ctx context.Context, obj *models.Media) (*models.MediaURL, error) {
	if obj.Type != models.MediaTypePhoto {
		return nil, nil
	}

	return dataloader.For(ctx).MediaHighres.Load(obj.ID)
}

// VideoWeb is the resolver for the videoWeb field.
func (r *mediaResolver) VideoWeb(ctx context.Context, obj *models.Media) (*models.MediaURL, error) {
	if obj.Type != models.MediaTypeVideo {
		return nil, nil
	}

	return dataloader.For(ctx).MediaVideoWeb.Load(obj.ID)
}

// Album is the resolver for the album field.
func (r *mediaResolver) Album(ctx context.Context, obj *models.Media) (*models.Album, error) {
	var album models.Album
	err := r.DB(ctx).Find(&album, obj.AlbumID).Error
	if err != nil {
		return nil, err
	}
	return &album, nil
}

// Exif is the resolver for the exif field.
func (r *mediaResolver) Exif(ctx context.Context, obj *models.Media) (*models.MediaEXIF, error) {
	if obj.Exif != nil {
		return obj.Exif, nil
	}

	var exif models.MediaEXIF
	if err := r.DB(ctx).Model(obj).Association("Exif").Find(&exif); err != nil {
		return nil, err
	}

	return &exif, nil
}

// Favorite is the resolver for the favorite field.
func (r *mediaResolver) Favorite(ctx context.Context, obj *models.Media) (bool, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return false, auth.ErrUnauthorized
	}

	return dataloader.For(ctx).UserMediaFavorite.Load(&models.UserMediaData{
		UserID:  user.ID,
		MediaID: obj.ID,
	})
}

// Type is the resolver for the type field.
func (r *mediaResolver) Type(ctx context.Context, obj *models.Media) (models.MediaType, error) {
	formattedType := models.MediaType(cases.Title(language.Und).String(string(obj.Type)))
	return formattedType, nil
}

// Shares is the resolver for the shares field.
func (r *mediaResolver) Shares(ctx context.Context, obj *models.Media) ([]*models.ShareToken, error) {
	var shareTokens []*models.ShareToken
	if err := r.DB(ctx).Where("media_id = ?", obj.ID).Find(&shareTokens).Error; err != nil {
		return nil, fmt.Errorf("get shares for media (%s): %w", obj.Path, err)
	}

	return shareTokens, nil
}

// Downloads is the resolver for the downloads field.
func (r *mediaResolver) Downloads(ctx context.Context, obj *models.Media) ([]*models.MediaDownload, error) {
	var mediaUrls []*models.MediaURL
	if err := r.DB(ctx).Where("media_id = ?", obj.ID).Find(&mediaUrls).Error; err != nil {
		return nil, fmt.Errorf("get downloads for media (%s): %w", obj.Path, err)
	}

	downloads := make([]*models.MediaDownload, 0)

	for _, url := range mediaUrls {

		var title string
		switch {
		case url.Purpose == models.MediaOriginal:
			title = "Original"
		case url.Purpose == models.PhotoThumbnail:
			title = "Small"
		case url.Purpose == models.PhotoHighRes:
			title = "Large"
		case url.Purpose == models.VideoThumbnail:
			title = "Video thumbnail"
		case url.Purpose == models.VideoWeb:
			title = "Web optimized video"
		}

		downloads = append(downloads, &models.MediaDownload{
			Title:    title,
			MediaURL: url,
		})
	}

	return downloads, nil
}

// Faces is the resolver for the faces field.
func (r *mediaResolver) Faces(ctx context.Context, obj *models.Media) ([]*models.ImageFace, error) {
	if face_detection.GlobalFaceDetector == nil {
		return []*models.ImageFace{}, nil
	}

	if obj.Faces != nil {
		return obj.Faces, nil
	}

	var faces []*models.ImageFace
	if err := r.DB(ctx).Model(obj).Association("Faces").Find(&faces); err != nil {
		return nil, err
	}

	return faces, nil
}

// FavoriteMedia is the resolver for the favoriteMedia field.
func (r *mutationResolver) FavoriteMedia(ctx context.Context, mediaID int, favorite bool) (*models.Media, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	return user.FavoriteMedia(r.DB(ctx), mediaID, favorite)
}

// MyMedia is the resolver for the myMedia field.
func (r *queryResolver) MyMedia(ctx context.Context, order *models.Ordering, paginate *models.Pagination) ([]*models.Media, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, fmt.Errorf("unauthorized")
	}

	return actions.MyMedia(r.DB(ctx), user, order, paginate)
}

// Media is the resolver for the media field.
func (r *queryResolver) Media(ctx context.Context, id int, tokenCredentials *models.ShareTokenCredentials) (*models.Media, error) {
	db := r.DB(ctx)
	if tokenCredentials != nil {

		shareToken, err := r.ShareToken(ctx, *tokenCredentials)
		if err != nil {
			return nil, err
		}

		if *shareToken.MediaID == id {
			return shareToken.Media, nil
		}
	}

	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	var media models.Media

	err := db.
		Joins("Album").
		Where("media.id = ?", id).
		Where("EXISTS (SELECT * FROM user_albums WHERE user_albums.album_id = media.album_id AND user_albums.user_id = ?)",
			user.ID).
		Where("media.id IN (?)", db.Model(&models.MediaURL{}).Select("media_id").Where("media_urls.media_id = media.id")).
		First(&media).Error

	if err != nil {
		return nil, fmt.Errorf("could not get media by media_id and user_id from database: %w", err)
	}

	return &media, nil
}

// MediaList is the resolver for the mediaList field.
func (r *queryResolver) MediaList(ctx context.Context, ids []int) ([]*models.Media, error) {
	db := r.DB(ctx)
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("no ids provided")
	}

	var media []*models.Media
	err := db.Model(&media).
		Joins("LEFT JOIN user_albums ON user_albums.album_id = media.album_id").
		Where("media.id IN ?", ids).
		Where("user_albums.user_id = ?", user.ID).
		Find(&media).Error

	if err != nil {
		return nil, fmt.Errorf("could not get media list by media_id and user_id from database: %w", err)
	}

	return media, nil
}

// Media returns api.MediaResolver implementation.
func (r *Resolver) Media() api.MediaResolver { return &mediaResolver{r} }

type mediaResolver struct{ *Resolver }
