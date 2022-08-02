package myaws

import (
	"bytes"
	"mime/multipart"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/dawkaka/theone/entity"
	"github.com/google/uuid"
	"github.com/h2non/bimg"
)

func UploadImageFile(fileHeader *multipart.FileHeader, bucket string) (string, error) {
	f, err := fileHeader.Open()
	var s string
	if err != nil {
		return s, err
	}
	extName := path.Ext(fileHeader.Filename)
	s = uuid.New().String() + extName

	var sess = session.Must(session.NewSession(&aws.Config{
		Region:     aws.String("eu-central-1"),
		MaxRetries: aws.Int(3),
	}))

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(s),
		Body:        f,
		ContentType: aws.String("image/" + extName[1:]),
	})
	return s, err
}

func UploadMultipleFiles(files []*multipart.FileHeader, bucket string) ([]entity.PostMetadata, error) {
	ch := make(chan any, len(files))
	for _, file := range files {
		go upload(file, ch)
	}
	filesMeta := []entity.PostMetadata{}
	for val := range ch {
		if val == entity.ErrImageProcessing {
			return nil, entity.ErrImageProcessing
		}
		filesMeta = append(filesMeta, val.(entity.PostMetadata))
		if len(filesMeta) == len(files) {
			return filesMeta, nil
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
		Quality:   95,
		Interlace: true,
	}
	newImage, err := bimg.NewImage(image).Process(options)
	if err != nil {
		ch <- entity.ErrImageProcessing
	}

	extName := path.Ext(file.Filename)
	s := uuid.New().String() + extName

	var sess = session.Must(session.NewSession(&aws.Config{
		Region:     aws.String("eu-central-1"),
		MaxRetries: aws.Int(3),
	}))

	imageReader := bytes.NewReader(newImage)

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String("postimages"),
		Key:         aws.String(s),
		Body:        imageReader,
		ContentType: aws.String("image/" + extName[1:]),
	})

	if err != nil {
		ch <- entity.ErrImageProcessing
	}
	ch <- entity.PostMetadata{Name: s, Height: int64(height), Width: int64(width), Type: extName[1:]}
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
