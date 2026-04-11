package domain

import (
	"errors"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Link struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ShortCode   string             `bson:"short_code" json:"short_code"`
	OriginalURL string             `bson:"original_url" json:"original_url"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	Visits      int64              `bson:"visits" json:"visits"`
}

func NewLink(shortCode, originalURL string) (*Link, error) {
	if err := validateURL(originalURL); err != nil {
		return nil, err
	}

	if err := validateShortCode(shortCode); err != nil {
		return nil, err
	}

	return &Link{
		ID:          primitive.NewObjectID(),
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		CreatedAt:   time.Now().UTC(),
		Visits:      0,
	}, nil
}

func validateURL(rawURL string) error {
	if rawURL == "" {
		return errors.New("url is required")
	}

	if len(rawURL) > 2048 {
		return errors.New("url too long")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return errors.New("invalid url format")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("only http and https schemes are allowed")
	}

	return nil
}

func validateShortCode(code string) error {
	if code == "" {
		return errors.New("short code is required")
	}

	if len(code) > 11 {
		return errors.New("short code too long")
	}

	for _, r := range code {
		if !((r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')) {
			return errors.New("short code contains invalid characters")
		}
	}

	return nil
}
