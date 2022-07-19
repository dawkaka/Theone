package handler

import (
	"encoding/gob"
	"net/http"
	"strconv"
	"time"

	"github.com/dawkaka/theone/app/middlewares"
	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/entity"
	"github.com/dawkaka/theone/inter"
	"github.com/dawkaka/theone/pkg/myaws"
	"github.com/dawkaka/theone/pkg/password"
	"github.com/dawkaka/theone/pkg/utils"
	"github.com/dawkaka/theone/pkg/validator"
	"github.com/dawkaka/theone/usecase/couple"
	"github.com/dawkaka/theone/usecase/user"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const EIGHTEEN_YEARS = 157680 //number of hours in 18 years

func signup(service user.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var newUser *entity.Signup
		err := ctx.ShouldBind(newUser)
		lang := utils.GetLang("", ctx.Request.Header)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, entity.ErrNotFound.Error()))
			return
		}

		errs := newUser.Validate()
		if len(errs) > 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"type": "ERROR", "errors": errs})
			return
		}

		firstName, lastName, userName, email, dateOfBirth, userPassword :=
			newUser.FirstName, newUser.LastName, newUser.UserName,
			newUser.Email, newUser.DateOfBirth, newUser.Password

		hashedPassword, err := password.Generate(userPassword)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, err.Error()))
			return
		}

		insertedID, err := service.CreateUser(email, hashedPassword, firstName, lastName, userName, dateOfBirth)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, err.Error()))
			return
		}
		session := sessions.Default(ctx)
		userSession := entity.UserSession{
			ID:                insertedID,
			Email:             email,
			Name:              userName,
			FirstName:         firstName,
			LastName:          lastName,
			HasPartner:        false,
			HasPendingRequest: false,
			DateOfBirth:       dateOfBirth,
		}
		gob.Register(userSession)
		session.Set("user", userSession)
		_ = session.Save()
		ctx.JSON(http.StatusCreated, presentation.Success(lang, "Signup successfull"))
	}
}

func login(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var login *entity.Login
		err := ctx.ShouldBind(login)
		lang := utils.GetLang("", ctx.Request.Header)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, entity.ErrSomethingWentWrong.Error()))
			return
		}
		user, err := service.GetUser(login.UserName)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, entity.ErrSomethingWentWrong.Error()))
			return
		}
		err = password.Compare(user.Password, login.Password)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "Wrong user name or password"))
			return
		}
		if user.Lang != "" {
			lang = user.Lang
		}
		session := sessions.Default(ctx)
		userSession := entity.UserSession{
			ID:                user.ID,
			Name:              user.UserName,
			Email:             user.Email,
			HasPartner:        user.HasPartner,
			PartnerID:         user.PartnerID,
			CoupleID:          user.CoupleID,
			HasPendingRequest: user.HasPendingRequest,
			FirstName:         user.FirstName,
			LastName:          user.LastName,
			Lang:              lang,
			DateOfBirth:       user.DateOfBirth,
		}
		gob.Register(userSession)
		session.Set("user", userSession)
		_ = session.Save()
		ctx.JSON(http.StatusOK, presentation.Success(lang, "login successfull"))
	}
}

func getUser(service user.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		userName := ctx.Param("userName")
		thisUser := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(thisUser.Lang, ctx.Request.Header)
		if !validator.IsUserName(userName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "Invalid user name"))
			return
		}
		user, err := service.GetUser(userName)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, entity.ErrNotFound.Error()))
			return
		}
		pUser := presentation.UserProfile{
			FirstName:      user.UserName,
			LastName:       user.LastName,
			UserName:       user.UserName,
			ProfilePicture: user.ProfilePicture,
			Bio:            user.Bio,
			FollowingCount: user.FollowingCount,
		}

		ctx.JSON(http.StatusOK, gin.H{"user": pUser})
	}
}

func searchUsers(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		query := ctx.Param("query")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if !validator.IsUserName(query) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		users, err := service.SearchUsers(query)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"users": users})

	}
}

func getFollowing(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		userName := user.Name
		skip, err := strconv.Atoi(ctx.Param("skip"))
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethinWentWrong"))
			return
		}
		following, err := service.UserFollowing(userName, skip)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusNotFound, presentation.Error(lang, "NotFound"))
				return
			}
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}

		page := entity.Pagination{
			Next: skip + entity.Limit,
			End:  len(following) < entity.Limit,
		}
		ctx.JSON(http.StatusOK, gin.H{"following": following, "pagination": page})
	}
}

func initiateRequest(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		thisUser := session.Get("user").(entity.UserSession)
		userName := ctx.Param("userName")
		lang := utils.GetLang(thisUser.Lang, ctx.Request.Header)
		if !validator.IsUserName(userName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "InvalidUserName"))
			return
		}
		userAge := time.Since(thisUser.DateOfBirth)
		if userAge.Hours() < EIGHTEEN_YEARS {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "UserLessThan18"))
			return
		}
		if thisUser.HasPartner || thisUser.HasPendingRequest {
			ctx.JSON(http.StatusMethodNotAllowed, presentation.Error(lang, "UserHasPartnerOrPendingRequest"))
			return
		}

		partner, err := service.GetUser(userName)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusNotFound, presentation.Error(lang, "UserNotFound"))
				return
			}
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		partnerAge := time.Since(partner.DateOfBirth)
		if partnerAge.Hours() < EIGHTEEN_YEARS {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "PartnerLessThan18"))
			return
		}
		if partner.HasPartner || partner.HasPendingRequest {
			ctx.JSON(http.StatusMethodNotAllowed, presentation.Error(lang, "PartnerHasPartnerOrPendingRequest"))
			return
		}
		err = service.CreateRequest(thisUser.ID, partner.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		notification := entity.NotifyRequest{
			Type:     "Couple Request",
			UserName: thisUser.Name,
			Message: inter.LocalizeWithFullName(
				lang,
				thisUser.FirstName,
				thisUser.LastName,
				"NewCoupleRequest",
			),
		}
		_ = service.NotifyUser(userName, notification)
		thisUser.HasPendingRequest = true
		thisUser.PartnerID = partner.ID

		session.Set("user", thisUser)
		session.Save()

		ctx.JSON(http.StatusCreated, presentation.Success(lang, "RequestCreated"))
	}
}

func follow(service user.UseCase, coupleService couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		coupleName := ctx.Param("coupleName")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		userName := user.Name
		userID := user.ID
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if !validator.IsCoupleName(coupleName) || !validator.IsUserName(userName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		couple, err := coupleService.GetCouple(coupleName)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusNotFound, presentation.Error(lang, "CoupleNotFound"))
				return
			}
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}

		err = service.Follow(couple.ID, userID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}

		_ = coupleService.NewFollower(userID, couple.ID)
		notif := entity.Notification{
			Type:    "follow",
			Message: inter.LocalizeWithUserName(lang, userName, "NewFollower"),
		}
		_ = service.NotifyCouple([2]primitive.ObjectID{couple.Accepted, couple.Initiated}, notif)
		ctx.JSON(http.StatusNoContent, presentation.Success(lang, "Followed"))
	}
}

func unfollow(service user.UseCase, coupleService couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		coupleName := ctx.Param("coupleName")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		userName := user.Name
		userID := user.ID
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if !validator.IsCoupleName(coupleName) || !validator.IsUserName(userName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		couple, err := coupleService.GetCouple(coupleName)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusNotFound, presentation.Error(lang, "CoupleNotFound"))
				return
			}
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		err = service.Unfollow(couple.ID, userID)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		_ = coupleService.RemoveFollower(userID, couple.ID)
		ctx.JSON(http.StatusNoContent, presentation.Success(lang, "Unfollowed"))
	}
}

func updateUser(service user.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		var update entity.UpdateUser
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
		err = service.UpdateUser(user.ID, update)
		if err != nil {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		ctx.JSON(http.StatusNoContent, presentation.Success(lang, "UserUpdated"))
	}
}

func deleteUser(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		user := session.Get("user").(entity.UserSession)
		userId := user.ID
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		err := service.DeleteUser((userId))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		ctx.JSON(http.StatusAccepted, presentation.Success(lang, "AccountDeleted"))
	}
}

func updateUserProfilePic(service user.UseCase) gin.HandlerFunc {
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
		err = service.UpdateUserProfilePic(fileName, user.ID)
		if err != nil {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "Forbidden"))
			return
		}
		ctx.JSON(http.StatusCreated, presentation.Success(lang, "ProfilePicUpdated"))
	}
}

func updateShowPicture(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		index, err := strconv.Atoi(ctx.Param("index"))
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		fileHeader, err := ctx.FormFile("show_picture")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}

		fileName, err := myaws.UploadImageFile(fileHeader, "toonjimages")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		err = service.UpdateShowPicture(user.ID, index, fileName)

		if err != nil {
			if err == entity.ErrNoMatch {
				ctx.JSON(http.StatusForbidden, presentation.Error(lang, "Forbidden"))
				return
			}
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		ctx.JSON(http.StatusCreated, presentation.Success(lang, "ShowPictureChanged"))
	}
}

func MakeUserHandlers(r *gin.Engine, service user.UseCase, coupleService couple.UseCase) {
	r.GET("/user/:userName", middlewares.Authenticate(), getUser(service))
	r.GET("/user/search/:query", searchUsers(service))
	r.GET("/user/following/:skip", getFollowing(service))
	r.POST("/user/signup", signup(service))
	r.POST("/user/login", login(service))
	r.PATCH("/user/follow/:coupleName", follow(service, coupleService))
	r.PATCH("/user/couple-request/:userName", initiateRequest(service))
	r.PATCH("/user/unfollow/:coupleName", unfollow(service, coupleService))
	r.PATCH("/user/update/profile-pic", updateUserProfilePic(service))
	r.PUT("/user/update", updateUser(service))
	r.PUT("/user/show-pictures/:index", updateShowPicture(service))
	r.DELETE("/user/delete-account", deleteUser(service))
}
