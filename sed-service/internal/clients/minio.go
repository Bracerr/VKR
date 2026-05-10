package clients

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/industrial-sed/sed-service/internal/config"
)

// Minio хранилище файлов.
type Minio struct {
	Client *minio.Client
	Bucket string
}

// NewMinio клиент S3-совместимый.
func NewMinio(cfg *config.Config) (*Minio, error) {
	c, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, err
	}
	return &Minio{Client: c, Bucket: cfg.MinioBucket}, nil
}

// EnsureBucket создаёт bucket при отсутствии.
func (m *Minio) EnsureBucket(ctx context.Context) error {
	exists, err := m.Client.BucketExists(ctx, m.Bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return m.Client.MakeBucket(ctx, m.Bucket, minio.MakeBucketOptions{})
}

// Put читает из reader до size байт.
func (m *Minio) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := m.Client.PutObject(ctx, m.Bucket, key, r, size, minio.PutObjectOptions{ContentType: contentType})
	return err
}

// Get стрим объекта.
func (m *Minio) Get(ctx context.Context, key string) (*minio.Object, error) {
	o, err := m.Client.GetObject(ctx, m.Bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return o, nil
}

// Remove удаляет объект.
func (m *Minio) Remove(ctx context.Context, key string) error {
	return m.Client.RemoveObject(ctx, m.Bucket, key, minio.RemoveObjectOptions{})
}

// ObjectKey формирует ключ.
func ObjectKey(tenant string, docID string, fileID string, origName string) string {
	return fmt.Sprintf("%s/%s/%s_%s", tenant, docID, fileID, origName)
}
