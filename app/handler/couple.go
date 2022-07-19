package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/entity"
	"github.com/dawkaka/theone/inter"
	myaws "github.com/dawkaka/theone/pkg/aws"
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
		lang := utils.GetLang(userb.Lang, ctx.Request.Header)
		partnerID, err := entity.StringToID(ctx.Param("partnerID"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		userId := userb.ID

		users, err := userService.ListUsers([]primitive.ObjectID{userId, partnerID})
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		if len(users) < 2 {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "InvalidParnterRequest"))
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
			ctx.JSON(http.StatusMethodNotAllowed, presentation.Error(lang, "NotAllowed"))
			return
		}

		coupleName := fmt.Sprintf("%s&%s_%d", partner.FirstName, user.FirstName, time.Now())

		Id, err := service.CreateCouple(userb.ID.String(), partnerID.String(), coupleName)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		notif := entity.Notification{
			Type: "Request accepted",
			Message: inter.LocalizeWithFullName(
				lang,
				user.FirstName,
				user.LastName,
				"RequestAccepted",
			),
		}
		_ = userService.NewCouple([2]entity.ID{userb.ID, partnerID}, Id)
		_ = userService.NotifyUser(partner.UserName, notif)
		ctx.JSON(http.StatusCreated, presentation.Success(lang, "CoupleCreated"))
	}

}

func getCouple(service couple.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		coupleName := ctx.Param("coupleName")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if !validator.IsCoupleName(coupleName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "Invalid couple name"))
			return
		}

		couple, err := service.GetCouple(coupleName)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "Something went wrong"))
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
			Married:        couple.Married,
			Verified:       couple.Verified,
		}
		ctx.JSON(http.StatusOK, gin.H{"couple": pCouple})
	}
}

func getCouplePosts(service couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		coupleName, skip := ctx.Param("coupleName"), ctx.Param("skip")
		skipPosts, err := strconv.Atoi(skip)
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if !validator.IsCoupleName(coupleName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
		}
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
		}
		posts, err := service.GetCouplePosts(coupleName, skipPosts)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
		}
		page := entity.Pagination{
			Next: skipPosts + entity.Limit,
			End:  len(posts) < entity.Limit,
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"posts": posts, "pagination": page})
	}
}

func getCoupleVideos(service couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		coupleName, skip := ctx.Param("coupleName"), ctx.Param("skip")
		skipVideos, err := strconv.Atoi(skip)
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if !validator.IsCoupleName(coupleName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
		}
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
		}
		videos, err := service.GetCoupleVideos(coupleName, skipVideos)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
		}
		page := entity.Pagination{
			Next: skipVideos + entity.Limit,
			End:  len(videos) < entity.Limit,
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"videos": videos, "pagination": page})
	}
}

func getFollowers(service couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		coupleName := ctx.Param("coupleName")
		skip, err := strconv.Atoi(ctx.Param("skip")) //pagination
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
		}
		if !validator.IsCoupleName(coupleName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
		}
		followers, err := service.GetFollowers(coupleName, skip)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrongInternal"))
		}
		page := entity.Pagination{
			Next: skip + entity.Limit,
			End:  len(followers) < entity.Limit,
		}
		ctx.JSON(http.StatusOK, gin.H{"followers": followers, "pagination": page})
	}
}

func updateCoupleProfilePic(service couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fileHeader, err := ctx.FormFile("profile-picture")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		fileName, err := myaws.UploadImageFile(fileHeader, "toonjimages")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "ProfilePicFailed"))
			return
		}
		err = service.UpdateCoupleProfilePic(fileName, user.CoupleID)
		if err != nil {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "Forbidden"))
			return
		}
		ctx.JSON(http.StatusCreated, presentation.Success(lang, "ProfilePicSuccess"))
	}
}

func updateCoupleCoverPic(service couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fileHeader, err := ctx.FormFile("cover-picture")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		fileName, err := myaws.UploadImageFile(fileHeader, "toonjimages")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "ProfilePicFailed"))
			return
		}
		err = service.UpdateCoupleCoverPic(fileName, user.CoupleID)
		if err != nil {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "Forbidden"))
			return
		}
		ctx.JSON(http.StatusCreated, presentation.Success(lang, "ProfilePicUpdated"))
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
	r.PATCH("/couple/profile-picture", updateCoupleProfilePic(service))
	r.PATCH("/couple/cover-picture", updateCoupleCoverPic(service))
	r.PUT("/couple/update", updateCouple(service))
}
