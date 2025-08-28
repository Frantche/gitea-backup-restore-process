package storage

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/aws/smithy-go/logging"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"

	appconfig "github.com/Frantche/gitea-backup-restore-process/internal/config"
	"github.com/Frantche/gitea-backup-restore-process/pkg/logger"
)

// S3Backend implements StorageBackend for Amazon S3 compatible storage
type S3Backend struct {
	client *s3.Client
}

// S3Config holds S3-specific configuration
type S3Config struct {
	EndpointURL       string
	AccessKeyID       string
	SecretAccessKey   string
	Bucket            string
	BackupFilename    string
	Prefix            string
	SignatureVersion  string
	Verify            bool
	Region            string
}

// getS3Config reads S3 configuration from environment variables
func getS3Config() (*S3Config, error) {
	s3Config := &S3Config{
		EndpointURL:       os.Getenv("ENDPOINT_URL"),
		AccessKeyID:       os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey:   os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Bucket:            os.Getenv("BUCKET"),
		BackupFilename:    os.Getenv("BACKUP_FILENAME"),
		Prefix:            os.Getenv("PREFIX"),
		SignatureVersion:  "s3v4",
		Verify:            true,
		Region:            os.Getenv("REGION"),
	}
	
	if os.Getenv("VERIFY") == "false" {
		s3Config.Verify = false
	}
	
	if sv := os.Getenv("SIGNATURE_VERSION"); sv != "" {
		s3Config.SignatureVersion = sv
	}
	
	return s3Config, nil
}


// See https://github.com/aws/aws-sdk-go-v2/issues/1816.
func ignoreSigningHeaders(o *s3.Options, headers []string) {
    o.APIOptions = append(o.APIOptions, func(stack *middleware.Stack) error {
        if err := stack.Finalize.Insert(ignoreHeaders(headers), "Signing", middleware.Before); err != nil {
            return err
        }

        if err := stack.Finalize.Insert(restoreIgnored(), "Signing", middleware.After); err != nil {
            return err
        }

        return nil
    })
}

type ignoredHeadersKey struct{}

func ignoreHeaders(headers []string) middleware.FinalizeMiddleware {
    return middleware.FinalizeMiddlewareFunc(
        "IgnoreHeaders",
        func(ctx context.Context, in middleware.FinalizeInput, next middleware.FinalizeHandler) (out middleware.FinalizeOutput, metadata middleware.Metadata, err error) {
            req, ok := in.Request.(*smithyhttp.Request)
            if !ok {
                return out, metadata, &v4.SigningError{Err: fmt.Errorf("(ignoreHeaders) unexpected request middleware type %T", in.Request)}
            }

            ignored := make(map[string]string, len(headers))
            for _, h := range headers {
                ignored[h] = req.Header.Get(h)
                req.Header.Del(h)
            }

            ctx = middleware.WithStackValue(ctx, ignoredHeadersKey{}, ignored)

            return next.HandleFinalize(ctx, in)
        },
    )
}

func restoreIgnored() middleware.FinalizeMiddleware {
    return middleware.FinalizeMiddlewareFunc(
        "RestoreIgnored",
        func(ctx context.Context, in middleware.FinalizeInput, next middleware.FinalizeHandler) (out middleware.FinalizeOutput, metadata middleware.Metadata, err error) {
            req, ok := in.Request.(*smithyhttp.Request)
            if !ok {
                return out, metadata, &v4.SigningError{Err: fmt.Errorf("(restoreIgnored) unexpected request middleware type %T", in.Request)}
            }

            ignored, _ := middleware.GetStackValue(ctx, ignoredHeadersKey{}).(map[string]string)
            for k, v := range ignored {
                req.Header.Set(k, v)
            }

            return next.HandleFinalize(ctx, in)
        },
    )
}

// getClient creates and returns an S3 client
func (s *S3Backend) getClient() (*s3.Client, error) {
	if s.client != nil {
		return s.client, nil
	}
	
	s3Config, err := getS3Config()
	if err != nil {
		return nil, err
	}
	
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(s3Config.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			s3Config.AccessKeyID,
			s3Config.SecretAccessKey,
			"",
		)),
		config.WithLogger(logging.NewStandardLogger(os.Stderr)), // Add logger
        config.WithClientLogMode(aws.LogRequestWithBody|aws.LogResponseWithBody|aws.LogRetries), // Enable detailed logs
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	
	// S3 options
    opt := s3.Options{
        BaseEndpoint: aws.String(s3Config.EndpointURL),
        UsePathStyle: true,
        Credentials:  cfg.Credentials,
        Region:       s3Config.Region,

        RequestChecksumCalculation:  aws.RequestChecksumCalculationWhenRequired,
        ResponseChecksumValidation:  aws.ResponseChecksumValidationWhenRequired,
    }

    ignoreSigningHeaders(&opt, []string{"Accept-Encoding"})

    s.client = s3.New(opt)
	
	return s.client, nil
}

func (s *S3Backend) ValidateConfig() error {
	s3Config, err := getS3Config()
	if err != nil {
		return err
	}
	
	if s3Config.EndpointURL == "" {
		return fmt.Errorf("ENDPOINT_URL is required for S3 backend")
	}
	if s3Config.AccessKeyID == "" {
		return fmt.Errorf("AWS_ACCESS_KEY_ID is required for S3 backend")
	}
	if s3Config.SecretAccessKey == "" {
		return fmt.Errorf("AWS_SECRET_ACCESS_KEY is required for S3 backend")
	}
	if s3Config.Bucket == "" {
		return fmt.Errorf("BUCKET is required for S3 backend")
	}
	if s3Config.Region == "" {
        return fmt.Errorf("REGION is required for S3 backend")
    }
	
	return nil
}

func (s *S3Backend) Upload(settings *appconfig.Settings) error {
	client, err := s.getClient()
	if err != nil {
		return err
	}
	
	s3Config, err := getS3Config()
	if err != nil {
		return err
	}
	
	// Open the file to upload
	file, err := os.Open(settings.BackupTmpFilename)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()
	
	// Upload to S3
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s3Config.Bucket),
		Key:    aws.String(settings.BackupTmpRemoteFilename),
		Body:   file,
	})
	
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}
	
	logger.Info("Upload to S3 successful")
	return nil
}

func (s *S3Backend) Download(settings *appconfig.Settings) error {
	client, err := s.getClient()
	if err != nil {
		return err
	}
	
	s3Config, err := getS3Config()
	if err != nil {
		return err
	}
	
	// Download from S3
	result, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s3Config.Bucket),
		Key:    aws.String(s3Config.BackupFilename),
	})
	if err != nil {
		return fmt.Errorf("failed to download from S3: %w", err)
	}
	defer result.Body.Close()
	
	// Create the local file
	file, err := os.Create(settings.RestoreTmpFilename)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()
	
	// Copy the downloaded content to the local file
	_, err = file.ReadFrom(result.Body)
	if err != nil {
		return fmt.Errorf("failed to write downloaded content: %w", err)
	}
	
	logger.Info("Download from S3 successful")
	return nil
}

func (s *S3Backend) EnsureMaxRetention(settings *appconfig.Settings) error {
	client, err := s.getClient()
	if err != nil {
		return err
	}
	
	s3Config, err := getS3Config()
	if err != nil {
		return err
	}
	
	// List objects with the backup prefix
	result, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(s3Config.Bucket),
		Prefix: aws.String(settings.BackupPrefix),
	})
	if err != nil {
		return fmt.Errorf("failed to list S3 objects: %w", err)
	}
	
	// Sort objects by last modified date (newest first)
	sort.Slice(result.Contents, func(i, j int) bool {
		return result.Contents[i].LastModified.After(*result.Contents[j].LastModified)
	})
	
	// Delete objects beyond the retention limit
	if len(result.Contents) > settings.BackupMaxRetention {
		objectsToDelete := result.Contents[settings.BackupMaxRetention:]
		
		for _, obj := range objectsToDelete {
			// Don't delete the current backup file if it exists
			if settings.BackupFilename != "" && *obj.Key == settings.BackupFilename {
				continue
			}
			
			_, err := client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
				Bucket: aws.String(s3Config.Bucket),
				Key:    obj.Key,
			})
			if err != nil {
				logger.Errorf("Failed to delete object %s: %v", *obj.Key, err)
				continue
			}
			
			logger.Infof("Deleted old backup from S3: %s", *obj.Key)
		}
	}
	
	return nil
}