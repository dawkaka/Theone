package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/entity"
	"github.com/dawkaka/theone/inter"
	"github.com/dawkaka/theone/pkg/myaws"
	"github.com/dawkaka/theone/pkg/utils"
	"github.com/dawkaka/theone/pkg/validator"
	"github.com/dawkaka/theone/repository"
	"github.com/dawkaka/theone/usecase/couple"
	"github.com/dawkaka/theone/usecase/user"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		if len(users) < 2 {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "InvalidPartnerRequest"))
			return
		}
		var (
			partner presentation.UserPreview
			user    presentation.UserPreview
		)
		for i := 0; i < len(users); i++ {
			if users[i].ID == userId {
				user = users[i]
			} else if users[i].ID == partnerID {
				partner = users[i]
			}
		}

		if partner.PendingRequest != entity.SENT_REQUEST || user.ID != partner.PartnerID {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "Forbidden"))
			return
		}

		coupleName := fmt.Sprintf("%s.and.%s_%d", partner.FirstName, user.FirstName, time.Now().Unix())
		Id, err := service.CreateCouple(userb.ID.Hex(), partnerID.Hex(), coupleName)
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
		go func() {
			_ = userService.NewCouple([2]entity.ID{userb.ID, partnerID}, Id)
			_ = userService.NotifyUser(partner.UserName, notif)
		}()
		ctx.SetCookie("couple_ID", Id.Hex(), 500, "/", "", false, true)
		ctx.JSON(http.StatusCreated, presentation.Success(lang, "CoupleCreated"))
	}

}

func getCouple(service couple.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		coupleName := ctx.Param("coupleName")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if !validator.IsCoupleName(coupleName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "InvalidCoupleName"))
			return
		}

		couple, err := service.GetCouple(coupleName)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
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
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		var update entity.UpdateCouple
		err := ctx.ShouldBindJSON(&update)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		update.Lang = lang
		update.Sanitize()
		errs := update.Validate()
		if len(errs) > 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"type": "error", "errors": errs})
			return
		}
		err = service.UpdateCouple(user.CoupleID, update)
		if err != nil {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		ctx.JSON(http.StatusOK, presentation.Success(lang, "CoupleUpdated"))
	}
}

func changeCoupleName(service couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		newCoupleName := ctx.PostForm("couple_name")
		session := sessions.Default(ctx)
		user := session.Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if !validator.IsCoupleName(newCoupleName) {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, presentation.Error(lang, "InvalidCoupleName"))
		}
		_, err := service.GetCouple(newCoupleName)
		if err == nil {
			ctx.AbortWithStatusJSON(http.StatusConflict, presentation.Error(lang, "CoupleAlreadyExists"))
		} else {
			if err != mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			}
		}
		err = service.ChangeCoupleName(user.CoupleID, newCoupleName)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
		}
		ctx.JSON(http.StatusCreated, presentation.Success(lang, "ChangedCoupleName"))
	}
}

func lastLastEdonCast(service couple.UseCase, userService user.UseCase) gin.HandlerFunc { //Na everybody go chop breakfast
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		user := session.Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if !user.HasPartner {
			ctx.AbortWithStatusJSON(http.StatusForbidden, presentation.Error(lang, "IsSingle"))
		}
		err := service.BreakUp(user.CoupleID)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
		}

		go func() {
			notif := entity.Notification{
				Type:    "breakUp",
				Message: inter.LocalizeWithFullName(lang, user.FirstName, user.LastName, "YourPartnerBrokeUpWithYou"),
			}

			err = userService.NotifyCouple([2]entity.ID{user.PartnerID, primitive.NewObjectID()}, notif)
		}()
		ctx.SetCookie("couple_ID", "", -33, "/", "", false, true)
		ctx.JSON(http.StatusOK, presentation.Success(lang, "BreakedUp"))
	}
}

//Messages between partners
func coupleMessages(coupleMessage repository.CoupleMessage) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		skip, err := strconv.Atoi(ctx.Param("skip"))
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, presentation.Error(lang, "BadRequest"))
		}
		messages, err := coupleMessage.Get(user.CoupleID, skip)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
		}
		page := entity.Pagination{
			Next: skip + entity.Limit,
			End:  len(messages) < entity.Limit,
		}
		ctx.JSON(http.StatusOK, gin.H{"messages": messages, "page": page})
	}
}

//All users couple interected with
func usersCoupleMessages(userMessage repository.UserCoupleMessage) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		skip, err := strconv.Atoi(ctx.Param("skip"))
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, presentation.Error(lang, "BadRequest"))
		}
		messages, err := userMessage.Get(user.CoupleID, skip)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
		}
		page := entity.Pagination{
			Next: skip + entity.Limit,
			End:  len(messages) < entity.Limit,
		}
		ctx.JSON(http.StatusOK, gin.H{"messages": messages, "page": page})
	}
}

//Messages btn couple and specific user
func userCoupleMessages(service couple.UseCase, userService user.UseCase, messageService repository.UserCoupleMessage) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userName := ctx.Param("userName")
		skip, err := strconv.Atoi(ctx.Param("skip"))
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, presentation.Error(lang, "BadRequest"))
		}
		couple, err := userService.GetUser(userName)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, presentation.Error(lang, "CoupleNotFound"))
			} else {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			}
		}
		messages, err := messageService.GetToCouple(user.ID, couple.ID, skip)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
		}
		page := entity.Pagination{
			Next: skip + entity.Limit,
			End:  len(messages) < entity.Limit,
		}
		ctx.JSON(http.StatusOK, gin.H{"messages": messages, "page": page})
	}
}

func MakeCoupleHandlers(r *gin.Engine, service couple.UseCase, userService user.UseCase, coupleMessage repository.CoupleMessage, userMessage repository.UserCoupleMessage) {
	r.GET("/:coupleName", getCouple(service))
	r.GET("/:coupleName/posts/:skip", getCouplePosts(service))
	r.GET("/:coupleName/videos/:skip", getCoupleVideos(service))
	r.GET("/:coupleName/followers/:skip", getFollowers(service))
	r.GET("/couple/p-messages/:skip", coupleMessages(coupleMessage))
	r.GET("/couple/messages/:skip", usersCoupleMessages(userMessage))
	r.GET("/couple/u/messages/:userName/:skip", userCoupleMessages(service, userService, userMessage))
	r.POST("/couple/new/:partnerID", newCouple(service, userService))
	r.POST("/couple/break-up", lastLastEdonCast(service, userService))
	r.PATCH("/couple/profile-picture", updateCoupleProfilePic(service))
	r.PATCH("/couple/cover-picture", updateCoupleCoverPic(service))
	r.PUT("/couple/update", updateCouple(service))
	r.PUT("/couple/change-name", changeCoupleName(service))
}
