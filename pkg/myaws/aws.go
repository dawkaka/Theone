package myaws

import (
	"bytes"
	"mime/multipart"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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
	image := make([]byte, 512)
	f.Read(image)
	imgType, isValid := validator.IsSupportedImageType(image)
	if !isValid {
		return fileName, entity.ErrUnsupportedImage
	}

	fileName = uuid.New().String() + "." + strings.Split(imgType, "/")[1]

	var sess = session.Must(session.NewSession(&aws.Config{
		Region:                        aws.String(endpoints.UsWest1RegionID),
		MaxRetries:                    aws.Int(3),
		CredentialsChainVerboseErrors: aws.Bool(true),
	}))

	//creds := stscreds.NewCredentials(sess, "dawkaka")

	// uploader := s3manager.NewUploader(sess)
	// _, err = uploader.Upload(&s3manager.UploadInput{
	// 	Bucket:      aws.String(bucket),
	// 	Key:         aws.String(fileName),
	// 	Body:        f,
	// 	ContentType: aws.String(imgType),
	// })

	s3C := s3.New(sess)
	_, err = s3C.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(fileName),
		Body:        f,
		ContentType: aws.String(imgType),
	})
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
