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

const EIGHTEEN_YEARS = 157680

func signup(service user.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var newUser *entity.Signup
		err := ctx.ShouldBind(newUser)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, entity.ErrNotFound.Error()))
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
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, err.Error()))
			return
		}

		insertedID, err := service.CreateUser(email, hashedPassword, firstName, lastName, userName, dateOfBirth)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(ctx.Request.Header, err.Error()))
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
		ctx.JSON(http.StatusCreated, presentation.Success(ctx.Request.Header, "Signup successfull"))
	}
}

func login(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var login *entity.Login
		err := ctx.ShouldBind(login)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, entity.ErrSomethingWentWrong.Error()))
			return
		}
		user, err := service.GetUser(login.UserName)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, entity.ErrSomethingWentWrong.Error()))
			return
		}
		err = password.Compare(user.Password, login.Password)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "Wrong user name or password"))
			return
		}
		session := sessions.Default(ctx)
		userSession := entity.UserSession{
			ID:                user.ID,
			Name:              user.UserName,
			Email:             user.Email,
			HasPartner:        user.HasPartner,
			PartnerID:         user.PartnerID,
			HasPendingRequest: user.HasPendingRequest,
			FirstName:         user.FirstName,
			LastName:          user.LastName,
			DateOfBirth:       user.DateOfBirth,
		}
		gob.Register(userSession)
		session.Set("user", userSession)
		_ = session.Save()
		ctx.JSON(http.StatusOK, presentation.Success(ctx.Request.Header, "login successfull"))
	}
}

func getUser(service user.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		userName := ctx.Param("userName")
		if !validator.IsUserName(userName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "Invalid user name"))
			return
		}
		user, err := service.GetUser(userName)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, entity.ErrNotFound.Error()))
			return
		}
		pUser := presentation.UserProfile{
			UserName:       user.UserName,
			FirstName:      user.UserName,
			Bio:            user.Bio,
			LastName:       user.LastName,
			ProfilePicture: user.ProfilePicture,
			CoverPicture:   user.CoverPicture,
		}

		ctx.JSON(http.StatusOK, gin.H{"user": pUser})
	}
}

func searchUsers(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		query := ctx.Param("query")
		if !validator.IsUserName(query) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrong"))
			return
		}
		users, err := service.SearchUsers(query)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrong"))
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"users": users})

	}
}

func getFollowing(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userName := sessions.Default(ctx).Get("user").(entity.UserSession).Name
		skip, err := strconv.Atoi(ctx.Param("skip"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethinWentWrong"))
			return
		}
		following, err := service.UserFollowing(userName, skip)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusNotFound, presentation.Error(ctx.Request.Header, "NotFound"))
				return
			}
			ctx.JSON(http.StatusInternalServerError, presentation.Error(ctx.Request.Header, "SomethingWentWrongInternal"))
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
		if !validator.IsUserName(userName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "InvalidUserName"))
			return
		}
		userAge := time.Since(thisUser.DateOfBirth)
		if userAge.Hours() < EIGHTEEN_YEARS {
			ctx.JSON(http.StatusForbidden, presentation.Error(ctx.Request.Header, "UserLessThan18"))
			return
		}
		if thisUser.HasPartner || thisUser.HasPendingRequest {
			ctx.JSON(http.StatusMethodNotAllowed, presentation.Error(ctx.Request.Header, "UserHasPartnerOrPendingRequest"))
			return
		}

		partner, err := service.GetUser(userName)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusNotFound, presentation.Error(ctx.Request.Header, "UserNotFound"))
				return
			}
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrong"))
			return
		}
		partnerAge := time.Since(partner.DateOfBirth)
		if partnerAge.Hours() < EIGHTEEN_YEARS {
			ctx.JSON(http.StatusForbidden, presentation.Error(ctx.Request.Header, "PartnerLessThan18"))
			return
		}
		if partner.HasPartner || partner.HasPendingRequest {
			ctx.JSON(http.StatusMethodNotAllowed, presentation.Error(ctx.Request.Header, "PartnerHasPartnerOrPendingRequest"))
			return
		}
		err = service.CreateRequest(thisUser.ID, partner.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(ctx.Request.Header, "SomethingWentWrongInternal"))
			return
		}
		notification := entity.NotifyRequest{
			Type:     "Couple Request",
			UserName: thisUser.Name,
			Message: inter.LocalizeWithFullName(
				utils.GetLang(ctx.Request.Header),
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

		ctx.JSON(http.StatusCreated, presentation.Success(ctx.Request.Header, "RequestCreated"))
	}
}

func follow(service user.UseCase, coupleService couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		coupleName := ctx.Param("coupleName")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		userName := user.Name
		userID := user.ID

		if !validator.IsCoupleName(coupleName) || !validator.IsUserName(userName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "BadRequest"))
			return
		}
		couple, err := coupleService.GetCouple(coupleName)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusNotFound, presentation.Error(ctx.Request.Header, "CoupleNotFound"))
				return
			}
			ctx.JSON(http.StatusInternalServerError, presentation.Error(ctx.Request.Response.Header, "SomethingWentWrongInternal"))
			return
		}

		err = service.Follow(couple.ID, userID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(ctx.Request.Header, "SomethingWentWrongInternal"))
			return
		}
		_ = coupleService.NewFollower(userID, couple.ID)

	}
}

func updateUser(service user.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {

	}
}

func deleteUser(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		userId := session.Get("userID")

		err := service.DeleteUser((userId.(primitive.ObjectID)))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(ctx.Request.Header, "SomethingWentWrong"))
			return
		}
		ctx.JSON(http.StatusAccepted, presentation.Success(ctx.Request.Header, "Account deleted"))
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
	r.PUT("/user/update", updateUser(service))
	r.DELETE("/user/delete-account", deleteUser(service))
}
