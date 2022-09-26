package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dawkaka/theone/app/middlewares"
	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/entity"
	"github.com/dawkaka/theone/inter"
	"github.com/dawkaka/theone/pkg/myaws"
	"github.com/dawkaka/theone/pkg/password"
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

func signup(service user.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var newUser entity.Signup
		err := ctx.ShouldBindJSON(&newUser)
		lang := utils.GetLang("", ctx.Request.Header)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		errs := newUser.Validate()
		if len(errs) > 0 {
			strArr := []string{}
			for _, val := range errs {
				strArr = append(strArr, inter.Localize(lang, val.Error()))
			}
			ctx.JSON(http.StatusBadRequest, gin.H{"type": "ERROR", "errors": strArr})
			return
		}

		firstName, lastName, userName, email, dateOfBirth, userPassword, lang :=
			newUser.FirstName, newUser.LastName, newUser.UserName,
			newUser.Email, newUser.DateOfBirth, newUser.Password, utils.GetLang("", ctx.Request.Header)
		user, err := service.CheckSignup(userName, email)
		if user.UserName == userName {
			ctx.JSON(http.StatusConflict, presentation.Error(lang, "UserAlreadyExists"))
			return
		}

		if user.Email == email {
			ctx.JSON(http.StatusConflict, presentation.Error(lang, "EmailAlreadyExists"))
			return
		}

		if err == entity.ErrSomethingWentWrong {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}

		hashedPassword, err := password.Generate(userPassword)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(lang, "SomethingWentWrong"))
			return
		}

		insertedID, err := service.CreateUser(email, hashedPassword, firstName, lastName, userName, dateOfBirth, lang)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		session := sessions.Default(ctx)
		userSession := entity.UserSession{
			ID:             insertedID,
			Email:          email,
			Name:           userName,
			FirstName:      firstName,
			LastName:       lastName,
			HasPartner:     false,
			PendingRequest: entity.NO_REQUEST,
			DateOfBirth:    dateOfBirth,
			Lang:           lang,
			LastVisited:    time.Now(),
		}
		session.Set("user", userSession)
		_ = session.Save()
		ctx.SetCookie("user_ID", insertedID.Hex(), 500, "/", "", false, true)
		ctx.JSON(http.StatusCreated, presentation.Success(lang, "SignupSuccessfull"))
	}
}

func login(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var login *entity.Login
		err := json.NewDecoder(ctx.Request.Body).Decode(&login)
		lang := utils.GetLang("", ctx.Request.Header)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(lang, entity.ErrSomethingWentWrong.Error()))
			return
		}
		user, err := service.Login(login.UserNameOrEmail)
		if err != nil {
			if err == entity.ErrUserNotFound {
				ctx.JSON(http.StatusUnauthorized, presentation.Error(lang, "LoginFailed"))
				return
			}
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, entity.ErrSomethingWentWrong.Error()))
			return
		}
		err = password.Compare(user.Password, login.Password)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, presentation.Error(lang, "LoginFailed"))
			return
		}
		if user.Lang != "" {
			lang = user.Lang
		}
		session := sessions.Default(ctx)
		userSession := entity.UserSession{
			ID:             user.ID,
			Name:           user.UserName,
			Email:          user.Email,
			HasPartner:     user.HasPartner,
			PartnerID:      user.PartnerID,
			CoupleID:       user.CoupleID,
			PendingRequest: user.PendingRequest,
			FirstName:      user.FirstName,
			LastName:       user.LastName,
			Lang:           lang,
			DateOfBirth:    user.DateOfBirth,
			LastVisited:    user.LastVisited,
		}
		if user.HasPartner {
			ctx.SetCookie("couple_ID", user.CoupleID.Hex(), 500, "/", "", false, true)
		}
		ctx.SetCookie("user_ID", user.ID.Hex(), 500, "/", "", false, true)
		session.Set("user", userSession)
		_ = session.Save()
		ctx.JSON(http.StatusOK, presentation.Success(lang, "LoginSuccessfull"))
	}
}

func getUser(service user.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		userName := ctx.Param("userName")
		thisUser := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(thisUser.Lang, ctx.Request.Header)
		log.Println(lang)
		if !validator.IsUserName(userName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "InvalidUserName"))
			return
		}
		user, err := service.GetUser(userName)

		if err != nil {
			ctx.JSON(http.StatusNotFound, presentation.Error(lang, entity.ErrUserNotFound.Error()))
			return
		}
		pUser := presentation.UserProfile{
			FirstName:      user.FirstName,
			LastName:       user.LastName,
			UserName:       user.UserName,
			ProfilePicture: user.ProfilePicture,
			Bio:            user.Bio,
			FollowingCount: user.FollowingCount,
			ShowPictures:   user.ShowPictures,
		}

		ctx.JSON(http.StatusOK, gin.H{"user": pUser})
	}
}

func searchUsers(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		query := ctx.Param("query")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)

		//TODO:
		// if !validator.IsUserName(query) && !validator.IsRealName(query) {
		// 	ctx.JSON(http.StatusBadRequest, presentation.Error(user.Lang, "WrongUserNameFormat"))
		// 	return
		// }
		users, err := service.SearchUsers(query)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(user.Lang, "SomethingWentWrongInternal"))
			return
		}
		if users == nil {
			ctx.JSON(http.StatusOK, gin.H{"users": []string{}})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"users": users})
	}
}

func getFollowing(service user.UseCase, coupleService couple.UseCase) gin.HandlerFunc {
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
				ctx.JSON(http.StatusNotFound, presentation.Error(lang, "UserNotFound"))
				return
			}
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		couples, err := coupleService.ListCouple(following)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrong"))
		}

		page := entity.Pagination{
			Next: skip + entity.Limit,
			End:  len(following) < entity.Limit,
		}
		ctx.JSON(http.StatusOK, gin.H{"following": couples, "pagination": page})
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
		if thisUser.Name == userName {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "BadRequest"))
			return
		}
		if !validator.Is18Plus(thisUser.DateOfBirth) {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "UserLessThan18"))
			return
		}
		if thisUser.HasPartner || thisUser.PendingRequest != entity.NO_REQUEST {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "UserHasPartnerOrPendingRequest"))
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
		if !validator.Is18Plus(partner.DateOfBirth) {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "PartnerLessThan18"))
			return
		}
		if partner.HasPartner || partner.PendingRequest != entity.NO_REQUEST {
			ctx.JSON(http.StatusMethodNotAllowed, presentation.Error(lang, "PartnerHasPartnerOrPendingRequest"))
			return
		}
		if !partner.OpenToRequests {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "PartnerNotOpen"))
			return
		}
		err = service.SendRequest(thisUser.ID, partner.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		err = service.RecieveRequest(partner.ID, thisUser.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "RequestPartiallyCompleted"))
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
		thisUser.PendingRequest = entity.SENT_REQUEST
		thisUser.PartnerID = partner.ID

		session.Set("user", thisUser)
		session.Save()

		ctx.JSON(http.StatusCreated, presentation.Success(lang, "RequestCreated"))
	}
}

func getPendingRequest(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		res, err := service.GetUser(user.Name)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(user.Lang, "BadRequest"))
			return
		}
		if res.PendingRequest == entity.NO_REQUEST {
			ctx.JSON(http.StatusOK, gin.H{"request": []string{}})
			return
		}
		users, err := service.ListUsers([]entity.ID{user.PartnerID})
		if len(users) != 1 {
			ctx.JSON(http.StatusUnprocessableEntity, "SomethingWentWrong")
			return
		}
		partner := users[0]
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(user.Lang, "SomethingWentWrong"))
			return
		}
		request := presentation.UserPreview{
			ID:             partner.ID,
			FirstName:      partner.FirstName,
			LastName:       partner.LastName,
			UserName:       partner.UserName,
			PendingRequest: partner.PendingRequest,
			ProfilePicture: partner.ProfilePicture,
		}
		ctx.JSON(http.StatusOK, gin.H{"request": request})
	}
}

func cancelRequest(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		user := session.Get("user").(entity.UserSession)
		if user.PendingRequest != entity.SENT_REQUEST {
			ctx.JSON(http.StatusForbidden, presentation.Error(user.Lang, "BadRequest"))
			return
		}
		err := service.NullifyRequest([2]entity.ID{user.ID, user.PartnerID})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(user.Lang, "SomethingWentWrong"))
		}
		user.PendingRequest = entity.NO_REQUEST
		session.Set("user", user)
		session.Save()
		ctx.JSON(http.StatusOK, presentation.Success(user.Lang, "RequestCancelled"))
	}
}

func rejectRequest(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		user := session.Get("user").(entity.UserSession)
		if user.PendingRequest != entity.RECIEVED_REQUEST {
			ctx.JSON(http.StatusForbidden, presentation.Error(user.Lang, "BadRequest"))
			return
		}

		users, err := service.ListUsers([]entity.ID{user.PartnerID})
		if err != nil {
			ctx.JSON(http.StatusForbidden, presentation.Error(user.Lang, "SomethingWentWrong"))
		}
		initiator := users[0]
		err = service.NullifyRequest([2]entity.ID{user.ID, user.PartnerID})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(user.Lang, "SomethinWentWrong"))
			return
		}
		go func() {
			notif := entity.Notification{
				Type:    "Request Rejected",
				Message: inter.LocalizeWithFullName(initiator.Lang, user.FirstName, user.LastName, "RequestRejected"),
			}
			service.NotifyUser(initiator.UserName, notif)
		}()
		user.PendingRequest = entity.NO_REQUEST
		session.Set("user", user)
		session.Save()
		ctx.JSON(http.StatusOK, presentation.Success(user.Lang, "RequestRejected"))
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
		err = coupleService.NewFollower(userID, couple.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrong"))
			return
		}

		err = service.Follow(couple.ID, userID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		notif := entity.Notification{
			Type:    "follow",
			Message: inter.LocalizeWithUserName(lang, userName, "NewFollower"),
		}
		_ = service.NotifyCouple([2]primitive.ObjectID{couple.Accepted, couple.Initiated}, notif)
		ctx.JSON(http.StatusOK, presentation.Success(lang, "Followed"))
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
		err = coupleService.RemoveFollower(couple.ID, userID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		err = service.Unfollow(couple.ID, userID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		ctx.JSON(http.StatusOK, presentation.Success(lang, "Unfollowed"))
	}
}

func updateUser(service user.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		user := session.Get("user").(entity.UserSession)
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
		user.FirstName = update.FirstName
		user.LastName = update.LastName
		user.DateOfBirth = update.DateOfBirth
		session.Set("user", user)
		session.Save()
		ctx.JSON(http.StatusOK, presentation.Success(lang, "UserUpdated"))
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
		fileName, err := myaws.UploadImageFile(fileHeader, "theone-profile-images")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "ProfilePicFailed"))
			return
		}
		err = service.UpdateUserProfilePic(fileName, user.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "Forbidden"))
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
		if index < 0 || index > 5 {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		fileHeader, err := ctx.FormFile("show_picture")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}

		fileName, err := myaws.UploadImageFile(fileHeader, "theone-profile-images")
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

func changeRequestStatus(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		status := ctx.Param("status")
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if status != "ON" && status != "OFF" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		err := service.ChangeUserRequestStatus(user.ID, status)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		ctx.JSON(http.StatusAccepted, presentation.Success(lang, "RequestStatus"+status))
	}
}

func changeUserName(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		session := sessions.Default(ctx)
		user := session.Get("user").(entity.UserSession)
		lang := user.Lang
		var newName struct {
			UserName string `json:"user_name"`
		}
		err := ctx.ShouldBindJSON(&newName)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		newUserName := newName.UserName
		if !validator.IsUserName(newUserName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "InvalidUserName"))
			return
		}
		if newUserName == user.Name {
			ctx.JSON(http.StatusAccepted, presentation.Success(lang, "ChangedUserName"))
			return
		}
		_, err = service.GetUser(newUserName)
		if err == nil {
			ctx.JSON(http.StatusConflict, presentation.Error(lang, "UserAlreadyExists"))
			return
		} else {
			if err != entity.ErrUserNotFound {
				ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
				return
			}
		}
		err = service.ChangeUserName(user.ID, newUserName)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		user.Name = newUserName
		session.Set("user", user)
		session.Save()
		ctx.JSON(http.StatusCreated, presentation.Success(lang, "ChangedUserName"))
	}
}

//Get messages a user send to a particular couple for main inbox
func userToACoupleMessages(service user.UseCase, coupleService couple.UseCase, messageService repository.UserCoupleMessage) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		coupleName := ctx.Param("coupleName")
		skip, err := strconv.Atoi(ctx.Param("skip"))
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(lang, "BadRequest"))
			return
		}
		couple, err := coupleService.GetCouple(coupleName)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusNotFound, presentation.Error(lang, "CoupleNotFound"))
				return
			} else {
				ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
				return
			}
		}
		messages, err := messageService.GetToCouple(user.ID, couple.ID, skip)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		page := entity.Pagination{
			Next: skip + entity.Limit,
			End:  len(messages) < entity.Limit,
		}
		ctx.JSON(http.StatusOK, gin.H{"messages": messages, "page": page})
	}
}

//Get all messages user sent to unique couple
func userMessages(service user.UseCase, userMessage repository.UserCoupleMessage) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		skip, err := strconv.Atoi(ctx.Param("skip"))
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(lang, "BadRequest"))
			return
		}
		messages, err := userMessage.Get(user.ID, skip)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		page := entity.Pagination{
			Next: skip + entity.Limit,
			End:  len(messages) < entity.Limit,
		}
		ctx.JSON(http.StatusOK, gin.H{"messages": messages, "page": page})
	}
}

func userSession(ctx *gin.Context) {
	session := sessions.Default(ctx)
	user := session.Get("user").(entity.UserSession)
	ctx.JSON(http.StatusOK, gin.H{"session": user})
}

func logout(ctx *gin.Context) {
	session := sessions.Default(ctx)
	user := session.Get("user").(entity.UserSession)
	session.Clear()
	ctx.JSON(http.StatusOK, presentation.Success(user.Lang, "LogedOut"))
}

func changeSettings(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		setting, value := ctx.Param("setting"), ctx.Param("newValue")
		session := sessions.Default(ctx)
		user := session.Get("user").(entity.UserSession)
		lang := user.Lang
		if !validator.IsValidSetting(setting, value) {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "InvalidSetting"))
			return
		}
		err := service.ChangeSettings(user.ID, setting, value)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		if setting == "language" {
			lang = value
			user.Lang = lang
			session.Set("user", user)
			session.Save()
		}
		setting = strings.ToUpper(string(setting[0])) + setting[1:]
		ctx.JSON(http.StatusOK, presentation.Success(lang, setting+"Updated"))
	}
}

func notifications(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		skip, err := strconv.Atoi(ctx.Param("skip"))
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(user.Lang, "BadRequest"))
			return
		}
		notifs, err := service.GetNotifications(user.Name, skip)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(user.Lang, "SomethingWentWrong"))
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"notifications": notifs})
	}
}

func changePassword(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		p := entity.ChangePassword{}
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		err := ctx.ShouldBindJSON(&p)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(user.Lang, "BadRequest"))
		}
		msg := p.Validate()
		if msg != "" {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(user.Lang, msg))
			return
		}
		u, err := service.GetUser(user.Name)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(user.Lang, "SomethingWentWrongInternal"))
			return
		}

		err = password.Compare(u.Password, p.Current)
		if err != nil {
			ctx.JSON(http.StatusForbidden, presentation.Error(user.Lang, "Forbidden"))
			return
		}

		hash, err := password.Generate(p.New)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(user.Lang, "SomethingWentWrongInternal"))
			return
		}

		err = service.ChangeSettings(user.ID, "password", hash)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(user.Lang, "SomethingWentWrongInternal"))
		}

		ctx.JSON(http.StatusOK, presentation.Success(user.Lang, "PasswordChanged"))
	}
}

func MakeUserHandlers(r *gin.Engine, service user.UseCase, coupleService couple.UseCase, userMessage repository.UserCoupleMessage) {
	r.POST("/user/u/signup", signup(service))                            //tested
	r.POST("/user/u/login", login(service))                              //tested
	r.Use(middlewares.Authenticate())                                    //tested
	r.GET("/user/u/session", userSession)                                //tested
	r.GET("/user/:userName", getUser(service))                           //tested
	r.GET("/user/search/:query", searchUsers(service))                   //tested
	r.GET("/user/following/:skip", getFollowing(service, coupleService)) //tested
	r.GET("/user/u/pending-request", getPendingRequest(service))         //tested
	r.GET("/user/messages/:skip", userMessages(service, userMessage))
	r.GET("/user/c/messages/:coupleName/:skip", userToACoupleMessages(service, coupleService, userMessage))
	r.GET("/user/notifications/:skip", notifications(service))             //tested
	r.POST("/user/logout", logout)                                         //tested
	r.POST("/user/u/cancel-request", cancelRequest(service))               //tested
	r.POST("/user/u/reject-request", rejectRequest(service))               //tested
	r.POST("/user/couple-request/:userName", initiateRequest(service))     //tested
	r.POST("/user/follow/:coupleName", follow(service, coupleService))     //tested
	r.POST("/user/unfollow/:coupleName", unfollow(service, coupleService)) //tested
	r.PUT("/user/name", changeUserName(service))                           //tested
	r.PUT("/user/password", changePassword(service))
	r.PUT("/user/request-status/:status", changeRequestStatus(service))   //tested
	r.PUT("/user", updateUser(service))                                   //tested
	r.POST("/user/show-pictures/:index", updateShowPicture(service))      //tested
	r.PATCH("/user/settings/:setting/:newValue", changeSettings(service)) //tested
	r.POST("/user/profile-pic", updateUserProfilePic(service))            //tested
	r.DELETE("/user", deleteUser(service))
}
