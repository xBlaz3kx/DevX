package s3

import (
	"context"
	"io"
	"path/filepath"
	"time"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
	"github.com/xBlaz3kx/DevX/observability"
	"github.com/xBlaz3kx/DevX/tls"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Client interface {
	UploadToPublicBucket(ctx context.Context, body io.Reader, filename string) (*string, error)
	UploadToPrivateBucket(ctx context.Context, body io.Reader, filename string) (*string, error)
	DownloadFromPublicBucket(ctx context.Context, filename string) ([]byte, error)
	DownloadFromPrivateBucket(ctx context.Context, filename string) ([]byte, error)
	DeleteFileFromPublicBucket(ctx context.Context, filename string) error
	DeleteFileFromPrivateBucket(ctx context.Context, filename string) error
	FileExistsInPublicBucket(ctx context.Context, filename string) (bool, error)
	FileExistsInPrivateBucket(ctx context.Context, filename string) (bool, error)
	GetFileURLWithExpiry(ctx context.Context, filename string, duration time.Duration) (string, error)
	ListFilesInPrivateBucket(ctx context.Context, path string) ([]string, error)
}

type ClientImpl struct {
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	s3         *s3.S3
	config     Configuration
	obs        observability.Observability
}

// Configuration storage configuration with TLS
type Configuration struct {
	// Provider is the type of Configuration provider to use (AWS, Minio, etc).
	Provider string `yaml:"provider" json:"provider" mapstructure:"provider"`

	// Endpoint is the endpoint of the Configuration.
	Endpoint string `yaml:"endpoint" json:"endpoint" mapstructure:"endpoint"`

	// PublicURL is the public access url of the Configuration.
	PublicURL string `yaml:"publicUrl" json:"publicUrl" mapstructure:"publicUrl"`

	// AccessKey is access key for the Configuration connection.
	AccessKey string `yaml:"accessKey" json:"accessKey" mapstructure:"accessKey"`

	// Bucket is the name of the bucket to use.
	PrivateBucket string `yaml:"privateBucket" json:"privateBucket" mapstructure:"privateBucket"`

	// Bucket is the name of the bucket to use.
	PublicBucket string `yaml:"publicBucket" json:"publicBucket" mapstructure:"publicBucket"`

	// SecretKey is secret key for the Configuration connection.
	SecretKey string `yaml:"secretKey" json:"secretKey" mapstructure:"secretKey"`

	// Region is the region of the Configuration.
	Region string `yaml:"region" json:"region" mapstructure:"region"`

	AccessTokenId string `yaml:"accessTokenId" json:"accessTokenId" mapstructure:"accessTokenId"`

	// AccessTokenSecret is the access token secret for the Configuration connection.
	AccessTokenSecret string `yaml:"accessTokenSecret" json:"accessTokenSecret" mapstructure:"accessTokenSecret"`

	ForcePathStyle bool `yaml:"forcePathStyle" json:"forcePathStyle" mapstructure:"forcePathStyle"`

	// TLS is the TLS configuration for the Configuration connection.
	TLS tls.TLS `mapstructure:"tls" yaml:"tls" json:"tls"`
}

func (c Configuration) ToAWSConfig() *aws.Config {
	return &aws.Config{
		Endpoint:         &c.Endpoint,
		Region:           &c.Region,
		Credentials:      credentials.NewStaticCredentials(c.AccessTokenId, c.AccessTokenSecret, ""),
		S3ForcePathStyle: &c.ForcePathStyle,
	}
}

func NewClient(config Configuration, obs observability.Observability) Client {
	awsSession := session.Must(session.NewSession(config.ToAWSConfig()))
	uploader := s3manager.NewUploader(awsSession)
	downloader := s3manager.NewDownloader(awsSession)

	return &ClientImpl{
		uploader:   uploader,
		downloader: downloader,
		s3:         s3.New(awsSession),
		config:     config,
		obs:        obs.WithSpanKind(trace.SpanKindClient),
	}
}

func (i *ClientImpl) ListFilesInPrivateBucket(ctx context.Context, path string) ([]string, error) {
	ctx, end, logger := i.obs.LogSpan(ctx, "s3.ListFilesInPrivateBucket", zap.String("path", path))
	defer end()
	logger.Debug("Getting list of files in private bucket")

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(i.config.PrivateBucket),
		Prefix: aws.String(path),
	}

	objects, err := i.s3.ListObjectsV2WithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	logger.Debug("Found files in private bucket", zap.Any("files", objects.String()))

	// todo fetch metadata for each file
	files := []string{}
	for _, obj := range objects.Contents {
		if obj != nil && obj.Key != nil {
			// Hide the path from the user
			files = append(files, stringUtils.RemoveStart(*obj.Key, path))
		}
	}

	return files, nil
}

// todo add metadata to the upload, such as the user who uploaded it
func (i *ClientImpl) UploadToPublicBucket(ctx context.Context, body io.Reader, filename string) (*string, error) {
	ctx, end, logger := i.obs.LogSpan(ctx, "s3.UploadToPublicBucket", zap.String("fileName", filename))
	defer end()
	logger.Debug("Uploading to public bucket")

	input := s3manager.UploadInput{
		Bucket: aws.String(i.config.PublicBucket),
		Key:    aws.String(uuid.NewString() + filepath.Ext(filename)),
		ACL:    aws.String("public-read"),
		Body:   body,
	}

	res, err := i.uploader.UploadWithContext(ctx, &input)
	if err != nil {
		return nil, err
	}

	return &res.Location, nil
}

// todo add metadata to the upload, such as the user who uploaded it
func (i *ClientImpl) UploadToPrivateBucket(ctx context.Context, body io.Reader, filename string) (*string, error) {
	ctx, end, logger := i.obs.LogSpan(ctx, "s3.UploadToPrivateBucket", zap.String("fileName", filename))
	defer end()

	logger.Debug("Uploading to private bucket")

	input := s3manager.UploadInput{
		Bucket: aws.String(i.config.PrivateBucket),
		Key:    aws.String(filename),
		ACL:    aws.String("private"),
		Body:   body,
	}

	res, err := i.uploader.UploadWithContext(ctx, &input)
	if err != nil {
		logger.Error("Error uploading to private bucket", zap.Error(err))
		return nil, err
	}

	return &res.Location, nil
}

func (i *ClientImpl) DownloadFromPublicBucket(ctx context.Context, path string) ([]byte, error) {
	ctx, end, logger := i.obs.LogSpan(ctx, "s3.DownloadFromPublicBucket", zap.String("path", path))
	defer end()

	logger.Debug("Downloading from public bucket")

	input := &s3.GetObjectInput{
		Bucket: aws.String(i.config.PublicBucket),
		Key:    aws.String(path),
	}

	var buf []byte
	wab := aws.NewWriteAtBuffer(buf)

	_, err := i.downloader.DownloadWithContext(ctx, wab, input)
	if err != nil {
		logger.Error("Error uploading to private bucket", zap.Error(err))
		return nil, err
	}

	return wab.Bytes(), nil
}

func (i *ClientImpl) DownloadFromPrivateBucket(ctx context.Context, path string) ([]byte, error) {
	ctx, end, logger := i.obs.LogSpan(ctx, "s3.DownloadFromPrivateBucket", zap.String("path", path))
	defer end()

	logger.Debug("Downloading from private bucket")

	input := &s3.GetObjectInput{
		Bucket: aws.String(i.config.PrivateBucket),
		Key:    aws.String(path),
	}

	var buf []byte
	wab := aws.NewWriteAtBuffer(buf)

	_, err := i.downloader.DownloadWithContext(ctx, wab, input)
	if err != nil {
		return nil, err
	}

	return wab.Bytes(), nil
}

func (i *ClientImpl) DeleteFileFromPublicBucket(ctx context.Context, filename string) error {
	ctx, end := i.obs.Span(ctx, "s3.DeleteFileFromPrivateBucket", zap.String("fileName", filename))
	defer end()

	_, deleteErr := i.s3.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(i.config.PublicBucket),
		Key:    aws.String(filename),
	})

	if deleteErr != nil {
		return deleteErr
	}

	return nil
}

func (i *ClientImpl) DeleteFileFromPrivateBucket(ctx context.Context, filename string) error {
	ctx, end := i.obs.Span(ctx, "s3.DeleteFileFromPrivateBucket", zap.String("fileName", filename))
	defer end()

	_, deleteErr := i.s3.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(i.config.PrivateBucket),
		Key:    aws.String(filename),
	})
	if deleteErr != nil {
		return deleteErr
	}

	return nil
}

func (i *ClientImpl) FileExistsInPrivateBucket(ctx context.Context, filename string) (bool, error) {
	ctx, end := i.obs.Span(ctx, "s3.FileExistsInPrivateBucket", zap.String("fileName", filename))
	defer end()

	_, err := i.s3.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(i.config.PrivateBucket),
		Key:    aws.String(filename),
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

func (i *ClientImpl) FileExistsInPublicBucket(ctx context.Context, filename string) (bool, error) {
	ctx, end := i.obs.Span(ctx, "s3.FileExistsInPublicBucket", zap.String("fileName", filename))
	defer end()

	_, err := i.s3.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(i.config.PublicBucket),
		Key:    aws.String(filename),
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

func (i *ClientImpl) GetFileURLWithExpiry(ctx context.Context, filename string, duration time.Duration) (string, error) {
	// nolint:all
	ctx, end := i.obs.Span(ctx, "s3.GetFileURLWithExpiry", zap.String("fileName", filename))
	defer end()

	req, _ := i.s3.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(i.config.PrivateBucket),
		Key:    aws.String(filename),
	})
	if req.Error != nil {
		return "", req.Error
	}

	signedURL, err := req.Presign(duration) // URL will be valid for duration
	if err != nil {
		return "", err
	}

	return signedURL, nil
}
