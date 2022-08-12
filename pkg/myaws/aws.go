package myaws

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/dawkaka/theone/config"
	"github.com/dawkaka/theone/entity"
	"github.com/dawkaka/theone/pkg/validator"
	"github.com/google/uuid"
	"github.com/h2non/bimg"
)

func UploadImageFile(fileHeader *multipart.FileHeader, bucket string) (string, error) {
	f, err := fileHeader.Open()
	var fileName string
	if err != nil {
		return fileName, err
	}
	fmt.Println(fileHeader.Filename)
	image := make([]byte, fileHeader.Size)
	f.Read(image)
	imgType, isValid := validator.IsSupportedImageType(image)

	if !isValid {
		return fileName, entity.ErrUnsupportedImage
	}

	fileName = uuid.New().String() + "." + strings.Split(imgType, "/")[1]

	session, err := session.NewSession(&aws.Config{
		Region:                        aws.String(endpoints.UsWest1RegionID),
		MaxRetries:                    aws.Int(3),
		CredentialsChainVerboseErrors: aws.Bool(true),
		Credentials: credentials.NewStaticCredentials(
			config.AWS_ACCESS_KEY,
			config.AWS_SECRET_KEY,
			""),
	})
	if err != nil {
		panic(err)
	}

	uploader := s3manager.NewUploader(session)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(fileName),
		Body:        bytes.NewReader(image),
		ContentType: aws.String("image/jpeg"),
	})
	fmt.Println(err)
	return fileName, err
}

func UploadMultipleFiles(files []*multipart.FileHeader, bucket string) ([]entity.PostMetadata, error) {
	ch := make(chan any, len(files))
	for _, file := range files {
		go upload(file, ch)
	}
	filesMeta := []entity.PostMetadata{}
	for val := range ch {
		switch v := val.(type) {
		case error:
			return nil, v
		case entity.PostMetadata:
			filesMeta = append(filesMeta, v)
			if len(filesMeta) == len(files) {
				return filesMeta, nil
			}
		}
	}
	return filesMeta, nil
}

func upload(file *multipart.FileHeader, ch chan any) {
	f, err := file.Open()
	if err != nil {
		ch <- entity.ErrImageProcessing
	}

	image := []byte{}
	f.Read(image)
	imgType, isValid := validator.IsSupportedImageType(image)
	if !isValid {
		ch <- entity.ErrUnsupportedImage
	}
	size, err := bimg.NewImage(image).Size()
	if err != nil {
		ch <- entity.ErrImageProcessing
	}

	height := getPrefDimention(size.Height, "h")
	width := getPrefDimention(size.Width, "w")

	options := bimg.Options{
		Width:     height,
		Height:    width,
		Crop:      false,
		Quality:   100,
		Interlace: true,
	}
	newImage, err := bimg.NewImage(image).Process(options)
	if err != nil {
		ch <- entity.ErrImageProcessing
	}

	fileName := uuid.New().String() + "." + imgType[len(imgType)-3:]
	var sess = session.Must(session.NewSession(&aws.Config{
		Region:     aws.String("eu-central-1"),
		MaxRetries: aws.Int(3),
	}))

	imageReader := bytes.NewReader(newImage)

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String("postimages"),
		Key:         aws.String(fileName),
		Body:        imageReader,
		ContentType: aws.String(imgType),
	})

	if err != nil {
		ch <- entity.ErrImageProcessing
	}
	ch <- entity.PostMetadata{Name: fileName, Height: int64(height), Width: int64(width), Type: imgType}
}

func getPrefDimention(curr int, d string) int {
	var dimen int
	if d == "h" {
		if curr < 566 {
			dimen = 566
		} else if curr > 1350 {
			dimen = 1350
		} else {
			dimen = curr
		}
	}
	if d == "w" {
		if curr < 320 {
			dimen = 320
		} else if curr > 1080 {
			dimen = 1080
		} else {
			dimen = curr
		}
	}

	return dimen
}
