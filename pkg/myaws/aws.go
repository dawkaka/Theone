package myaws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/dawkaka/theone/config"
	"github.com/dawkaka/theone/entity"
	"github.com/dawkaka/theone/pkg/validator"
	"github.com/google/uuid"
	"github.com/h2non/bimg"
)

var sess *session.Session
var acl = "public-read"

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
		Quality:   50,
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
		ACL:         &acl,
	})
	return fileName, err
}

func UploadMultipleFiles(files []*multipart.FileHeader) ([]entity.PostMetadata, *entity.CustomError) {
	ch := make(chan any, len(files))
	filesMeta := make([]entity.PostMetadata, len(files))
	doneCount := 0

	for i := 0; i < len(files); i++ {
		go upload(files[i], ch, i)
	}

	for val := range ch {
		switch v := val.(type) {
		case entity.CustomError:
			return filesMeta, &v
		case entity.PostMetadata:
			ind, err := strconv.Atoi(v.Alt)
			if err != nil {
				return filesMeta, &entity.CustomError{Code: 500, Message: "SomethingWentWrongInternal"}
			}
			filesMeta[ind] = v
			doneCount++
			if doneCount == len(files) {
				return filesMeta, nil
			}
		}
	}
	return filesMeta, nil
}

func upload(file *multipart.FileHeader, ch chan any, i int) {
	f, err := file.Open()
	if err != nil {
		ch <- entity.ErrImageProcessing
	}
	pFile := make([]byte, file.Size)
	if file.Size > 16000000 {
		ch <- entity.CustomError{Code: http.StatusForbidden, Message: entity.ErrFileTooLarge.Error()}
	}
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
		vMeta := vStreamMeta.Streams[0]
		vDuration, err := strconv.ParseFloat(vMeta.Duration, 64)
		if err != nil {
			ch <- entity.CustomError{Code: http.StatusUnprocessableEntity, Message: entity.ErrVideoProcessing.Error()}
			return
		}
		if vDuration > 30.00 {
			ch <- entity.CustomError{Code: http.StatusForbidden, Message: entity.ErrVideoTooLong.Error()}
			return
		}
		fileName := uuid.New().String() + "." + strings.Split(fType, "/")[1]
		videoReader := bytes.NewReader(pFile)
		uploader := s3manager.NewUploader(sess)
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket:      aws.String("theone-profile-images"),
			Key:         aws.String(fileName),
			Body:        videoReader,
			ContentType: aws.String(fType),
			ACL:         &acl,
		})
		if err != nil {
			ch <- entity.CustomError{Code: http.StatusInternalServerError, Message: entity.ErrAWSUpload.Error()}
			return
		}
		ch <- entity.PostMetadata{Name: fileName, Height: int64(vMeta.Height), Width: int64(vMeta.Width), Type: fType, Alt: fmt.Sprint(i)}

	} else if strings.Split(fType, "/")[0] == "image" {
		size, err := bimg.NewImage(pFile).Size()
		if err != nil {
			ch <- entity.CustomError{Code: http.StatusUnprocessableEntity, Message: entity.ErrImageProcessing.Error()}
		}

		height := size.Height
		width := size.Width
		aspR := math.Round(float64(size.Width)/float64(size.Height)*10) / 10 // round to one decimal place
		options := bimg.Options{Quality: 90, Interlace: true}

		if aspR != entity.Thumbnail && aspR != entity.Landscape && aspR != entity.Potrait {
			if width > height {
				width = height
			} else {
				height = width
			}
			options = bimg.Options{Width: height, Height: width, Crop: true, Quality: 90, Interlace: true}
		}

		newImage, err := bimg.NewImage(pFile).Process(options)
		if err != nil {
			ch <- entity.CustomError{Code: http.StatusUnprocessableEntity, Message: entity.ErrImageProcessing.Error()}
			return
		}
		fileName := uuid.New().String() + "." + strings.Split(fType, "/")[1]
		imageReader := bytes.NewReader(newImage)
		uploader := s3manager.NewUploader(sess)
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket:      aws.String("theone-profile-images"),
			Key:         aws.String(fileName),
			Body:        imageReader,
			ContentType: aws.String(fType),
			ACL:         &acl,
		})

		if err != nil {
			ch <- entity.CustomError{Code: http.StatusInternalServerError, Message: entity.ErrAWSUpload.Error()}
			return
		}
		ch <- entity.PostMetadata{Name: fileName, Height: int64(height), Width: int64(width), Type: fType, Alt: fmt.Sprint(i)}
	}
}

const (
	Sender = "mail@toonji.com"

	Subject = "Email Verification"

	TextBody = "This email was sent with Amazon SES using the AWS SDK for Go."

	CharSet = "UTF-8"
)

func SendEmail(Recipient, linkID, eType, lang string) error {
	// Create a new session in the us-west-2 region.
	// Replace us-west-2 with the AWS Region you're using for Amazon SES.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-central-1"), Credentials: credentials.NewStaticCredentials(
			config.AWS_SES_ACCESS_KEY,
			config.AWS_SES_SECRET_KEY,
			""),
	})

	if err != nil {
		return err
	}
	var tmpl *template.Template
	if eType == "verify-email" {
		tmpl = template.Must(template.ParseFiles("../templates/verify_email_" + lang + ".html"))
	} else {
		tmpl = template.Must(template.ParseFiles("../templates/reset_password_" + lang + ".html"))
	}
	body := bytes.Buffer{}
	tmpl.Execute(&body, struct{ LinkID string }{LinkID: linkID})
	svc := ses.New(sess)
	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(Recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(body.String()),
				},
				Text: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(TextBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(Subject),
			},
		},
		Source: aws.String(Sender),
		// Uncomment to use a configuration set
		//ConfigurationSetName: aws.String(ConfigurationSet),
	}

	_, err = svc.SendEmail(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				fmt.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				fmt.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				fmt.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
	}
	return err
}
