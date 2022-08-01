package handler

import (
	"reflect"
	"testing"

	"github.com/dawkaka/theone/usecase/video"
	"github.com/gin-gonic/gin"
)

func Test_editVideoCaption(t *testing.T) {
	type args struct {
		service video.UseCase
	}
	tests := []struct {
		name string
		args args
		want gin.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := editVideoCaption(tt.args.service); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("editVideoCaption() = %v, want %v", got, tt.want)
			}
		})
	}
}
