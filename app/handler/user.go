package handler

import (
	"net/http"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/entity"
	"github.com/dawkaka/theone/pkg/password"
	"github.com/dawkaka/theone/pkg/password/validator"
	"github.com/dawkaka/theone/usecase/user"
	"github.com/gin-gonic/gin"
)

func signup(service user.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var newUser *entity.Signup
		err := ctx.ShouldBind(newUser)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(entity.ErrNotFound.Error()))
			return
		}

		errs := newUser.Validate()
		if len(errs) > 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"type": "ERROR", "errors": errs})
			return
		}

		firstName, lastName, userName, email, dateOfBith, userPassword :=
			newUser.FirstName, newUser.LastName, newUser.UserName,
			newUser.Email, newUser.DateOfBirth, newUser.Password

		hashedPassword, err := password.Generate(userPassword)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(err.Error()))
			return
		}

		err = service.CreateUser(email, hashedPassword, firstName, lastName, userName, dateOfBith)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(err.Error()))
			return
		}

		ctx.JSON(http.StatusCreated, presentation.Success("Signup successfull"))
	}
}

func login(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var login *entity.Login
		err := ctx.ShouldBind(login)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error((entity.ErrSomethingWentWrong.Error())))
			return
		}
		user, err := service.GetUser(login.UserName)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error((entity.ErrSomethingWentWrong.Error())))
			return
		}
		err = password.Compare(user.Password, login.Password)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error("Wrong user name or password"))
			return
		}
		ctx.JSON(http.StatusOK, presentation.Success("login successfull"))
	}
}

func getUser(service user.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		userName := ctx.Param("userName")
		if !validator.IsUserName(userName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error("Invalid user name"))
			return
		}
		user, err := service.GetUser(userName)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(entity.ErrNotFound.Error()))
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

func updateUser(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func deleteUser(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func searchUsers(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func MakeUserHandlers(r *gin.Engine, service user.UseCase) {

	r.GET("/user/:userName", getUser(service))
	r.GET("/user/search/:query", searchUsers(service))
	r.POST("/user/signup", signup(service))
	r.POST("/user/login", login(service))
	r.PUT("/user/update", updateUser(service))
	r.DELETE("/user/delete-account", deleteUser(service))
}
