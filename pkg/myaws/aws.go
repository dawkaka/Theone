package myaws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strconv"
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

var sess *session.Session

func init() {
	se, err := session.NewSession(&aws.Config{
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
	sess = se
}

func UploadImageFile(fileHeader *multipart.FileHeader, bucket string) (string, error) {
	f, err := fileHeader.Open()
	var fileName string
	if err != nil {
		return fileName, err
	}
	if fileHeader.Size > 5000000 {
		return fileName, entity.ErrFileTooLarge
	}
	image := make([]byte, fileHeader.Size)
	f.Read(image)
	imgType, isValid := validator.IsSupportedFileType(image)

	if !isValid || strings.Split(imgType, "/")[0] != "image" {
		return fileName, entity.ErrUnsupportedImage
	}
	options := bimg.Options{
		Width:     400,
		Height:    400,
		Crop:      true,
		Quality:   90,
		Interlace: true,
	}
	newImage, err := bimg.NewImage(image).Process(options)
	if err != nil {
		return fileName, entity.ErrImageProcessing
	}

	fileName = uuid.New().String() + "." + strings.Split(imgType, "/")[1]

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(fileName),
		Body:        bytes.NewReader(newImage),
		ContentType: aws.String(imgType),
	})
	return fileName, err
}

func UploadMultipleFiles(files []*multipart.FileHeader) ([]entity.PostMetadata, *entity.CustomError) {
	ch := make(chan any, len(files))
	for _, file := range files {
		go upload(file, ch)
	}
	filesMeta := []entity.PostMetadata{}
	for val := range ch {
		switch v := val.(type) {
		case entity.CustomError:
			return nil, &v
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
	pFile := make([]byte, file.Size)
	// if file.Size > 16000000 {
	// 	ch <- entity.ErrFileTooLarge
	// }
	f.Read(pFile)
	fType, isValid := validator.IsSupportedFileType(pFile)
	if !isValid {
		ch <- entity.ErrUnsupportedFile
	}

	if strings.Split(fType, "/")[0] == "video" {
		file, err := ioutil.TempFile("", "theone-video-")
		if err != nil {
			ch <- entity.CustomError{Code: http.StatusUnprocessableEntity, Message: entity.ErrFileProcessing.Error()}
			return
		}
		defer os.Remove(file.Name())
		file.Write(pFile)
		cmd := exec.Command("ffprobe", "-show_streams", "-select_streams", "v:0", "-print_format", "json",
			file.Name())

		var out bytes.Buffer
		cmd.Stdout = &out
		err = cmd.Run()
		if err != nil {
			ch <- entity.CustomError{Code: http.StatusUnprocessableEntity, Message: entity.ErrVideoProcessing.Error()}
			return
		}
		vStreamMeta := entity.VideoStream{}
		err = json.Unmarshal(out.Bytes(), &vStreamMeta)
		if err != nil {
			ch <- entity.CustomError{Code: http.StatusUnprocessableEntity, Message: entity.ErrVideoProcessing.Error()}
			return
		}
		fmt.Println(vStreamMeta)
		vMeta := vStreamMeta.Streams[0]
		vDuration, err := strconv.ParseFloat(vMeta.Duration, 64)
		if err != nil {
			ch <- entity.CustomError{Code: http.StatusUnprocessableEntity, Message: entity.ErrVideoProcessing.Error()}
			return
		}
		if vDuration > 60.00 {
			ch <- entity.CustomError{Code: http.StatusForbidden, Message: entity.ErrVideoTooLong.Error()}
			return
		}
		fileName := uuid.New().String() + "." + strings.Split(fType, "/")[1]
		videoReader := bytes.NewReader(pFile)
		uploader := s3manager.NewUploader(sess)
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket:      aws.String("theone-postfiles"),
			Key:         aws.String(fileName),
			Body:        videoReader,
			ContentType: aws.String(fType),
		})

		if err != nil {
			ch <- entity.CustomError{Code: http.StatusInternalServerError, Message: entity.ErrAWSUpload.Error()}
			return
		}
		ch <- entity.PostMetadata{Name: fileName, Height: int64(vMeta.Height), Width: int64(vMeta.Width), Type: fType}

	} else if strings.Split(fType, "/")[0] == "image" {
		size, err := bimg.NewImage(pFile).Size()
		if err != nil {
			ch <- entity.CustomError{Code: http.StatusUnprocessableEntity, Message: entity.ErrImageProcessing.Error()}
		}
		height := getPrefDimention(size.Height, "h")
		width := getPrefDimention(size.Width, "w")
		options := bimg.Options{Width: height, Height: width, Crop: true, Quality: 100, Interlace: true}
		newImage, err := bimg.NewImage(pFile).Process(options)
		if err != nil {
			ch <- entity.CustomError{Code: http.StatusUnprocessableEntity, Message: entity.ErrImageProcessing.Error()}
			return
		}
		fileName := uuid.New().String() + "." + strings.Split(fType, "/")[1]
		imageReader := bytes.NewReader(newImage)
		uploader := s3manager.NewUploader(sess)
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket:      aws.String("theone-postfiles"),
			Key:         aws.String(fileName),
			Body:        imageReader,
			ContentType: aws.String(fType),
		})
		fmt.Println(err)

		if err != nil {
			ch <- entity.CustomError{Code: http.StatusInternalServerError, Message: entity.ErrAWSUpload.Error()}
			return
		}
		ch <- entity.PostMetadata{Name: fileName, Height: int64(height), Width: int64(width), Type: fType}
	}
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
