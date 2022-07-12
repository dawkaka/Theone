package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/entity"
	"github.com/dawkaka/theone/inter"
	"github.com/dawkaka/theone/pkg/utils"
	"github.com/dawkaka/theone/pkg/validator"
	"github.com/dawkaka/theone/usecase/couple"
	"github.com/dawkaka/theone/usecase/user"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func newCouple(service couple.UseCase, userService user.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		userb := session.Get("user").(entity.UserSession)
		partnerID, err := entity.StringToID(ctx.Param("partnerID"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrong"))
			return
		}
		userId := userb.ID

		users, err := userService.ListUsers([]primitive.ObjectID{userId, partnerID})
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrong"))
			return
		}
		if len(users) < 2 {
			ctx.JSON(http.StatusForbidden, presentation.Error(ctx.Request.Header, "InvalidParnterRequest"))
			return
		}
		var (
			partner entity.User
			user    entity.User
		)
		for i := 0; i < len(users); i++ {
			if users[i].ID == userId {
				user = users[i]
			} else if users[i].ID == partnerID {
				partner = users[i]
			}
		}

		if !partner.HasPendingRequest || user.ID != partner.PartnerID {
			ctx.JSON(http.StatusMethodNotAllowed, presentation.Error(ctx.Request.Header, "NotAllowed"))
			return
		}

		coupleName := fmt.Sprintf("%s&%s_%d", partner.FirstName, user.FirstName, time.Now())

		err = service.CreateCouple(userb.ID.String(), partnerID.String(), coupleName)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrong"))
			return
		}
		notif := entity.Notification{
			Type: "Request accepted",
			Message: inter.LocalizeWithFullName(
				utils.GetLang(ctx.Request.Header),
				user.FirstName,
				user.LastName,
				"RequestAccepted",
			),
		}
		_ = userService.NotifyUser(partner.UserName, notif)
		ctx.JSON(http.StatusCreated, presentation.Success(ctx.Request.Header, "CoupleCreated"))
	}

}

func getCouple(service couple.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		coupleName := ctx.Param("coupleName")
		if !validator.IsCoupleName(coupleName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "Invalid couple name"))
			return
		}

		couple, err := service.GetCouple(coupleName)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "Something went wrong"))
			return
		}
		pCouple := presentation.CoupleProfile{
			CoupleName:     couple.CoupleName,
			AcceptedAt:     couple.AcceptedAt,
			Bio:            couple.Bio,
			Status:         couple.Status,
			FollowersCount: couple.FollowersCount,
			ProfilePicture: couple.ProfilePicture,
			CoverPicture:   couple.CoverPicture,
			PostCount:      couple.PostCount,
		}
		ctx.JSON(http.StatusOK, gin.H{"couple": pCouple})
	}
}

func getCouplePosts(service couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		coupleName, skip := ctx.Param("coupleName"), ctx.Param("skip")
		skipPosts, err := strconv.Atoi(skip)
		if !validator.IsCoupleName(coupleName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "BadRequest"))
		}
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrong"))
		}
		posts, err := service.GetCouplePosts(coupleName, skipPosts)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrong"))
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"posts": posts})
	}
}

func getCoupleVideos(service couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		coupleName, skip := ctx.Param("coupleName"), ctx.Param("skip")
		skipPosts, err := strconv.Atoi(skip)
		if !validator.IsCoupleName(coupleName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "BadRequest"))
		}
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrong"))
		}
		posts, err := service.GetCoupleVideos(coupleName, skipPosts)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrong"))
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"videos": posts})
	}
}

func getFollowers(service couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		coupleName := ctx.Param("coupleName")
		skip, err := strconv.Atoi(ctx.Param("skip")) //pagination
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "BadRequest"))
		}
		if !validator.IsCoupleName(coupleName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "BadRequest"))
		}
		followers, err := service.GetFollowers(coupleName, skip)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrongInternal"))
		}
		ctx.JSON(http.StatusOK, gin.H{"followers": followers})
	}
}

func updateCouple(service couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func MakeCoupleHandlers(r *gin.Engine, service couple.UseCase, userService user.UseCase) {
	r.GET("/:coupleName", getCouple(service))
	r.GET("/:coupleName/posts/:skip", getCouplePosts(service))
	r.GET("/:coupleName/videos/:skip", getCoupleVideos(service))
	r.GET("/:coupleName/followers/:skip", getFollowers(service))
	r.POST("/couple/new/:partnerID", newCouple(service, userService))
	r.PUT("/couple/update", updateCouple(service))

}
