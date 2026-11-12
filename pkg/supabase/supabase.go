package supabase

import (
	"fmt"
	"io"

	"github.com/mqqff/wasteco-be/internal/domain/env"
	"github.com/mqqff/wasteco-be/pkg/log"
	storageGo "github.com/supabase-community/storage-go"
)

type SupabaseItf interface {
	UploadFile(bucket, path, mimeType string, body io.Reader) (string, error)
}

type SupabaseStruct struct {
	client *storageGo.Client
}

var Supabase = getSupabase()

func getSupabase() SupabaseItf {
	url := fmt.Sprintf("%s/storage/v1", env.AppEnv.SupabaseURL)
	client := storageGo.NewClient(url, env.AppEnv.SupabaseSecret, nil)

	return &SupabaseStruct{client}
}

func (s *SupabaseStruct) UploadFile(bucket, path, mimeType string, body io.Reader) (string, error) {
	if bucket == "" {
		bucket = env.AppEnv.SupabaseBucket
		if bucket == "" {
			bucket = "donations"
		}
	}

	res, err := s.client.UploadFile(bucket, path, body, storageGo.FileOptions{
		ContentType: &mimeType,
	})
	if err != nil {
		log.Error(log.LogInfo{
			"error":  err.Error(),
			"res":    res,
			"bucket": bucket,
			"path":   path,
		}, "[Supabase][UploadFile] failed to upload file")
		return "", err
	}

	publicURL := s.client.GetPublicUrl(bucket, path).SignedURL
	return publicURL, nil
}
