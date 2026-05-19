package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

func cosClient() *cos.Client {
	u, _ := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.COSBucket, cfg.COSRegion))
	b := &cos.BaseURL{BucketURL: u}
	return cos.NewClient(b, &http.Client{
		Timeout: 30 * time.Second,
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.COSSecretID,
			SecretKey: cfg.COSSecretKey,
		},
	})
}

// cosUpload uploads data to COS and returns the public URL.
func cosUpload(data []byte, key, contentType string) (string, error) {
	client := cosClient()
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	}
	_, err := client.Object.Put(context.Background(), key, bytes.NewReader(data), opt)
	if err != nil {
		return "", fmt.Errorf("COS上传失败: %w", err)
	}
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", cfg.COSBucket, cfg.COSRegion, key), nil
}

// cosUploadReader uploads from an io.Reader, returns public URL.
func cosUploadReader(reader io.Reader, key, contentType string) (string, error) {
	client := cosClient()
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	}
	_, err := client.Object.Put(context.Background(), key, reader, opt)
	if err != nil {
		return "", fmt.Errorf("COS上传失败: %w", err)
	}
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", cfg.COSBucket, cfg.COSRegion, key), nil
}

// cosDownload downloads a file from COS by key.
func cosDownload(key string) ([]byte, error) {
	client := cosClient()
	resp, err := client.Object.Get(context.Background(), key, nil)
	if err != nil {
		return nil, fmt.Errorf("COS下载失败: %w", err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// cosPublicURL returns the public URL for a COS key.
func cosPublicURL(key string) string {
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", cfg.COSBucket, cfg.COSRegion, key)
}

// cosListObjects lists object keys under a given prefix.
func cosListObjects(prefix string, maxKeys int) ([]string, error) {
	client := cosClient()
	opt := &cos.BucketGetOptions{
		Prefix:  prefix,
		MaxKeys: maxKeys,
	}
	result, _, err := client.Bucket.Get(context.Background(), opt)
	if err != nil {
		return nil, fmt.Errorf("COS列表失败: %w", err)
	}
	var keys []string
	for _, obj := range result.Contents {
		if obj.Key != prefix {
			keys = append(keys, obj.Key)
		}
	}
	return keys, nil
}

// generateCOSKey creates a COS key for a given folder and filename.
func generateCOSKey(folder, filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".jpg"
	}
	timestamp := time.Now().Unix()
	randomID := randomString(8)
	return fmt.Sprintf("id-photo/%s/%d_%s%s", folder, timestamp, randomID, ext)
}
