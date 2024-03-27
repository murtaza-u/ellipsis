package fs

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type s3Store struct {
	sess   *s3.S3
	bucket string
	region string
}

func NewS3Store(region, bucket string) (Storage, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:                        aws.String(region),
		CredentialsChainVerboseErrors: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create new s3 session: %w", err)
	}
	return &s3Store{
		sess:   s3.New(sess),
		bucket: bucket,
		region: region,
	}, nil
}

func (s s3Store) Put(k string, v io.ReadSeeker) error {
	_, err := s.sess.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(k),
		Body:   v,
	})
	return err
}

func (s s3Store) Delete(k string) error {
	_, err := s.sess.DeleteObject(&s3.DeleteObjectInput{
		Key:    aws.String(k),
		Bucket: aws.String(s.bucket),
	})
	return err
}

func (s s3Store) GetURL(k string) (string, error) {
	url := fmt.Sprintf(
		"https://%s.s3.%s.amazonaws.com/%s",
		s.bucket, s.region, k,
	)
	return url, nil
}
