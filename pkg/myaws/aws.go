package myaws

import (
	"fmt"
	"mime/multipart"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// The session the S3 Uploader will use
func UploadImageFile(fileHeader *multipart.FileHeader, bucket string) (string, error) {
	f, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	var re = regexp.MustCompile(`\W\_`)
	s := fmt.Sprintf("%s%d", re.ReplaceAllString(fileHeader.Filename, ""), time.Now().UnixMilli())

	var sess = session.Must(session.NewSession(&aws.Config{
		Region:     aws.String("eu-central-1"),
		MaxRetries: aws.Int(3),
	}))

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(s),
		Body:   f,
	})
	return s, err
}
