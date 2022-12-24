package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dawkaka/theone/app/middlewares"
	"github.com/dawkaka/theone/app/presentation"
	config "github.com/dawkaka/theone/conf"
	"github.com/dawkaka/theone/entity"
	"github.com/dawkaka/theone/inter"
	"github.com/dawkaka/theone/pkg/myaws"
	"github.com/dawkaka/theone/pkg/password"
	"github.com/dawkaka/theone/pkg/utils"
	"github.com/dawkaka/theone/pkg/validator"
	"github.com/dawkaka/theone/repository"
	"github.com/dawkaka/theone/usecase/couple"
	"github.com/dawkaka/theone/usecase/post"
	"github.com/dawkaka/theone/usecase/user"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func unverifiedSignup(service user.UseCase, verifyRepo repository.VerifyMongo) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var newUser entity.VerifySignup
		lang := utils.GetLang("", ctx.Request.Header)
		err := ctx.ShouldBindJSON(&newUser)
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

		linkID := primitive.NewObjectID().Hex()
		newUser.ID = linkID
		newUser.Date = time.Now().UnixMilli()
		userName, email := newUser.UserName, newUser.Email
		user, err := service.CheckSignup(userName, email)
		if err != nil && err != mongo.ErrNoDocuments {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		if user.UserName == userName || user.Email == email {
			errs := []string{}
			if user.UserName == userName {
				errs = append(errs, inter.Localize(lang, "UserAlreadyExists"))
			}
			if user.Email == email {
				errs = append(errs, inter.Localize(lang, "EmailAlreadyExists"))
			}
			ctx.JSON(http.StatusConflict, gin.H{"type": "ERROR", "errors": errs})
			return
		}

		err = verifyRepo.NewUser(newUser)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		err = myaws.SendEmail(newUser.Email, linkID, "verify-email", lang)
		count := 0
		for err != nil && count < 2 {
			err = myaws.SendEmail(newUser.Email, linkID, "verify-email", lang)
			count++
		}

		ctx.JSON(http.StatusAccepted, gin.H{"link": linkID})
	}
}

func resendEmailVerificationLink(verifyRepo repository.VerifyMongo) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		linkID := ctx.Param("id")
		lang := utils.GetLang("", ctx.Request.Header)
		newUser, err := verifyRepo.GetNewUser(linkID)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(lang, "InvalidOrExpiredLink"))
			return
		}
		err = myaws.SendEmail(newUser.Email, linkID, "verify-email", lang)
		count := 0
		for err != nil && count < 2 {
			err = myaws.SendEmail(newUser.Email, linkID, "verify-email", lang)
			count++
		}
		ctx.JSON(http.StatusNoContent, gin.H{})
	}
}

func signup(service user.UseCase, verifyRepo repository.VerifyMongo) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		linkID := ctx.Param("id")
		lang := utils.GetLang("", ctx.Request.Header)
		newUser, err := verifyRepo.GetNewUser(linkID)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(lang, "InvalidOrExpiredLink"))
			return
		}

		firstName, lastName, userName, email, dateOfBirth, userPassword, lang, country, state :=
			newUser.FirstName, newUser.LastName, newUser.UserName,
			newUser.Email, newUser.DateOfBirth, newUser.Password, lang, newUser.Country, newUser.State

		user, err := service.CheckSignup(userName, email)
		if err != nil && err != mongo.ErrNoDocuments {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}

		if user.UserName == userName {
			ctx.JSON(http.StatusConflict, presentation.Error(lang, "UserAlreadyExists"))
			return
		}

		if user.Email == email {
			ctx.JSON(http.StatusConflict, presentation.Error(lang, "EmailAlreadyExists"))
			return
		}

		hashedPassword, err := password.Generate(userPassword)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(lang, "SomethingWentWrong"))
			return
		}

		insertedID, err := service.CreateUser(email, hashedPassword, firstName, lastName, userName, dateOfBirth, lang, country, state)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		verifyRepo.Verified(linkID)
		session := sessions.Default(ctx)
		userSession := entity.UserSession{
			ID:             insertedID,
			Name:           userName,
			Email:          email,
			ProfilePicture: "defaultprofile.jpg",
			HasPartner:     false,
			PartnerID:      [12]byte{},
			CoupleID:       [12]byte{},
			PendingRequest: entity.NO_REQUEST,
			FirstName:      firstName,
			LastName:       lastName,
			Lang:           lang,
			DateOfBirth:    dateOfBirth,
			Country:        country,
			LastVisited:    time.Now(),
		}
		session.Set("user", userSession)
		_ = session.Save()
		ctx.SetCookie("user_ID", insertedID.Hex(), 5*365*24*60*60*1000, "/", config.COOKIE_DOMAIN, false, false)
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
		if user.Language != "" {
			lang = user.Language
		}
		session := sessions.Default(ctx)
		userSession := entity.UserSession{
			ID:             user.ID,
			Name:           user.UserName,
			Email:          user.Email,
			ProfilePicture: user.ProfilePicture,
			HasPartner:     user.HasPartner,
			PartnerID:      user.PartnerID,
			CoupleID:       user.CoupleID,
			PendingRequest: user.PendingRequest,
			FirstName:      user.FirstName,
			LastName:       user.LastName,
			Country:        user.Country,
			Lang:           lang,
			DateOfBirth:    user.DateOfBirth,
			LastVisited:    user.LastVisited,
		}
		if user.HasPartner {
			ctx.SetCookie("couple_ID", user.CoupleID.Hex(), 500, "/", config.COOKIE_DOMAIN, false, false)
		}
		ctx.SetCookie("user_ID", user.ID.Hex(), 5*365*24*60*60*1000, "/", config.COOKIE_DOMAIN, false, false)
		session.Set("user", userSession)
		_ = session.Save()
		ctx.JSON(http.StatusOK, presentation.Success(lang, "LoginSuccessfull"))
	}
}

func getUser(service user.UseCase, coupleService couple.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {

		userName := ctx.Param("userName")
		thisUser := utils.GetSession(sessions.Default(ctx))
		lang := utils.GetLang(thisUser.Lang, ctx.Request.Header)
		if !validator.IsUserName(userName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "InvalidUserName"))
			return
		}

		user, err := service.GetUser(userName)

		if err != nil {
			ctx.JSON(http.StatusNotFound, presentation.Error(lang, entity.ErrUserNotFound.Error()))
			return
		}

		coupleList, _ := coupleService.ListCouple([]entity.ID{user.CoupleID}, user.ID)

		var couple presentation.CouplePreview
		if len(coupleList) > 0 {
			couple = coupleList[0]
		}

		pUser := presentation.UserProfile{
			FirstName:      user.FirstName,
			LastName:       user.LastName,
			UserName:       user.UserName,
			ProfilePicture: user.ProfilePicture,
			Bio:            user.Bio,
			FollowingCount: user.FollowingCount,
			ShowPictures:   user.ShowPictures,
			HasPartner:     user.HasPartner,
			CoupleName:     couple.CoupleName,
			IsThisUser:     thisUser.ID == user.ID,
			Website:        user.Website,
		}

		if pUser.IsThisUser {
			pUser.DateOfBirth = user.DateOfBirth
		}

		ctx.JSON(http.StatusOK, pUser)
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
			ctx.JSON(http.StatusOK, []string{})
			return
		}
		ctx.JSON(http.StatusOK, users)
	}
}

func getFollowing(service user.UseCase, coupleService couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		name := ctx.Param("name")
		skip, err := strconv.Atoi(ctx.Param("skip"))
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		following, err := service.UserFollowing(name, skip)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusNotFound, presentation.Error(lang, "UserNotFound"))
				return
			}
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		couples, err := coupleService.ListCouple(following, user.ID)
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
		u, err := service.GetUser(thisUser.Name)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		if u.HasPartner || u.PendingRequest != entity.NO_REQUEST {
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
		notification := entity.Notification{
			Type:   entity.NOTIF.COUPLE_REQUEST,
			UserID: thisUser.ID,
			Date:   time.Now(),
		}

		thisUser.PendingRequest = entity.SENT_REQUEST
		thisUser.PartnerID = partner.ID
		session.Set("user", thisUser)
		session.Save()

		go func() {
			_ = service.NotifyUser(userName, notification)
		}()

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
			ctx.JSON(http.StatusOK, gin.H{"request": nil})
			return
		}
		users, err := service.ListUsers([]entity.ID{res.PartnerID})
		if len(users) != 1 {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(user.Lang, "SomethingWentWrong"))
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
		u, err := service.GetUser(user.Name)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(user.Lang, "SomethingWentWrong"))
			return
		}
		if u.PendingRequest != entity.SENT_REQUEST {
			ctx.JSON(http.StatusForbidden, presentation.Error(user.Lang, "BadRequest"))
			return
		}
		err = service.NullifyRequest([2]entity.ID{user.ID, user.PartnerID})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(user.Lang, "SomethingWentWrong"))
			return
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
		u, err := service.GetUser(user.Name)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(user.Lang, "SomethingWentWrong"))
			return
		}

		if u.PendingRequest != entity.RECIEVED_REQUEST {
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
			ctx.JSON(http.StatusInternalServerError, presentation.Error(user.Lang, "SomethingWentWrong"))
			return
		}
		go func() {
			notif := entity.Notification{
				Type:   entity.NOTIF.REQUEST_REJECTED,
				UserID: user.ID,
				Date:   time.Now(),
			}
			service.NotifyUser(initiator.UserName, notif)
		}()
		user.PendingRequest = entity.NO_REQUEST
		session.Set("user", user)
		session.Save()
		ctx.JSON(http.StatusOK, presentation.Success(user.Lang, "RequestRejected"))
	}
}

func follow(service user.UseCase, coupleService couple.UseCase, postService post.UseCase) gin.HandlerFunc {
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

		couple, err := coupleService.GetCouplePosts(coupleName, 0)
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

		st, ed := len(couple.Posts)-0-entity.LimitP, len(couple.Posts)-0

		if st < 0 {
			st = 0
		}
		if ed < 0 {
			ed = 0
		}

		postIDs := couple.Posts[st:ed]
		posts, _ := postService.GetPosts(couple.ID, user.ID, postIDs)
		postObjectIDs := []entity.ID{}
		for i := 0; i < len(posts); i++ {
			postObjectIDs = append(postObjectIDs, posts[i].ID)
		}
		err = service.Follow(couple.ID, userID, postObjectIDs)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}

		notif := entity.Notification{
			Type:   entity.NOTIF.FOLLOW,
			UserID: user.ID,
			Date:   time.Now(),
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
		couple, err := coupleService.GetCouple(coupleName, userID)
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
		userS := sessions.Default(ctx)
		user := userS.Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		u, err := service.GetUser(user.Name)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "Forbidden"))
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
		user.ProfilePicture = fileName
		userS.Set("user", user)
		userS.Save()

		go func() {
			err = myaws.DeleteFile(u.ProfilePicture, "theone-profile-images")
			count := 0
			for err != nil && count < 2 {
				err = myaws.DeleteFile(u.ProfilePicture, "theone-profile-images")
				count++
			}
		}()

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

		u, err := service.GetUser(user.Name)
		if err != nil {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "Forbidden"))
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

		go func() {
			err = myaws.DeleteFile(u.ShowPictures[index], "theone-profile-images")
			count := 0
			for err != nil && count < 2 {
				err = myaws.DeleteFile(u.ShowPictures[index], "theone-profile-images")
				count++
			}
		}()

		ctx.JSON(http.StatusCreated, presentation.Success(lang, "ShowPictureChanged"))
	}
}

func changeRequestStatus(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		status := ctx.Param("status")
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		u, err := service.GetUser(user.Name)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		if u.HasPartner || u.PendingRequest != entity.NO_REQUEST {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "CantChangeStatus"))
			return
		}
		if status != "ON" && status != "OFF" {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(lang, "BadRequest"))
			return
		}

		err = service.ChangeUserRequestStatus(user.ID, status)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		ctx.JSON(http.StatusOK, presentation.Success(lang, "RequestStatus"+status))
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

		if !validator.IsNotBadName(newUserName) {
			ctx.JSON(http.StatusConflict, presentation.Error(lang, "UserAlreadyExists"))
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
// func userToACoupleMessages(service user.UseCase, coupleService couple.UseCase, messageService repository.UserCoupleMessage) gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		coupleName := ctx.Param("coupleName")
// 		skip, err := strconv.Atoi(ctx.Param("skip"))
// 		user := sessions.Default(ctx).Get("user").(entity.UserSession)
// 		lang := utils.GetLang(user.Lang, ctx.Request.Header)
// 		if err != nil {
// 			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(lang, "BadRequest"))
// 			return
// 		}
// 		couple, err := coupleService.GetCouple(coupleName, user.ID)
// 		if err != nil {
// 			if err == mongo.ErrNoDocuments {
// 				ctx.JSON(http.StatusNotFound, presentation.Error(lang, "CoupleNotFound"))
// 				return
// 			} else {
// 				ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
// 				return
// 			}
// 		}
// 		messages, err := messageService.GetToCouple(user.ID, couple.ID, skip)
// 		if err != nil {
// 			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
// 			return
// 		}
// 		page := entity.Pagination{
// 			Next: skip + entity.Limit,
// 			End:  len(messages) < entity.Limit,
// 		}
// 		ctx.JSON(http.StatusOK, gin.H{"messages": messages, "page": page})
// 	}
// }

//Get all messages user sent to unique couple
// func userMessages(service user.UseCase, userMessage repository.UserCoupleMessage) gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		user := sessions.Default(ctx).Get("user").(entity.UserSession)
// 		skip, err := strconv.Atoi(ctx.Param("skip"))
// 		lang := utils.GetLang(user.Lang, ctx.Request.Header)
// 		if err != nil {
// 			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(lang, "BadRequest"))
// 			return
// 		}
// 		messages, err := userMessage.Get(user.ID, skip)
// 		if err != nil {
// 			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
// 			return
// 		}
// 		page := entity.Pagination{
// 			Next: skip + entity.Limit,
// 			End:  len(messages) < entity.Limit,
// 		}
// 		ctx.JSON(http.StatusOK, gin.H{"messages": messages, "page": page})
// 	}
// }

func userSession(service user.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		user := session.Get("user").(entity.UserSession)
		u, _ := service.GetUser(user.Name)
		var ss entity.UserSession = entity.UserSession{
			CoupleID:   u.CoupleID,
			Name:       u.UserName,
			FirstName:  u.FirstName,
			ID:         u.ID,
			LastName:   u.LastName,
			HasPartner: u.HasPartner,
		}
		ctx.JSON(http.StatusOK, gin.H{"session": ss})
	}

}

func logout(ctx *gin.Context) {
	session := sessions.Default(ctx)
	user := session.Get("user").(entity.UserSession)
	session.Clear()
	session.Options(sessions.Options{Path: "/", MaxAge: -1})
	session.Save()
	ctx.SetCookie("user_ID", "", -500, "/", config.COOKIE_DOMAIN, false, true)
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
			ctx.SetCookie("NEXT_LOCALE", lang, 1000*60*60*24*12*40, "/", config.COOKIE_DOMAIN, false, false)
		}
		setting = strings.ToUpper(string(setting[0])) + setting[1:]
		ctx.JSON(http.StatusOK, presentation.Success(lang, setting+"Updated"))
	}
}

func notifications(service user.UseCase, coupleService couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		skip, err := strconv.Atoi(ctx.Param("skip"))
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(user.Lang, "BadRequest"))
			return
		}
		notif, err := service.GetNotifications(user.ID, skip)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(user.Lang, "SomethingWentWrong"))
			return
		}

		usersMap := make(map[entity.ID]presentation.UserPreview)
		couplesMap := make(map[entity.ID]presentation.CouplePreview)

		userIDs := []entity.ID{}
		coupleIDs := []entity.ID{}

		for _, val := range notif.Notifications {
			if !val.UserID.IsZero() {
				usersMap[val.UserID] = presentation.UserPreview{}
			}
			if !val.CoupleID.IsZero() {
				couplesMap[val.CoupleID] = presentation.CouplePreview{}
			}
		}

		for key := range usersMap {
			userIDs = append(userIDs, key)
		}

		for key := range couplesMap {
			coupleIDs = append(coupleIDs, key)
		}

		users, _ := service.ListUsers(userIDs)
		couples, _ := coupleService.ListCouple(coupleIDs, primitive.ObjectID{})

		for _, val := range users {
			usersMap[val.ID] = val
		}

		for _, val := range couples {
			couplesMap[val.ID] = val
		}
		notifs := utils.GetNotifs(usersMap, couplesMap, notif.Notifications)

		page := entity.Pagination{
			Next: skip + entity.Limit,
			End:  len(notif.Notifications) < entity.Limit,
		}
		ctx.JSON(http.StatusOK, gin.H{"notifications": notifs, "new_count": notif.NewCount, "pagination": page})
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

func changeEmail(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		e := struct {
			Email string `json:"email"`
		}{}
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		err := ctx.ShouldBindJSON(&e)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(user.Lang, "BadRequest"))
			return
		}
		if !validator.IsEmail(e.Email) {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(user.Lang, "BadRequest"))
			return
		}

		u, err := service.CheckSignup("justSoOnlyEmailMatchesThisIsNotAValidUserName", e.Email)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(user.Lang, "SomethingWentWrong"))
			return
		}

		if u.Email == e.Email {
			ctx.JSON(http.StatusConflict, presentation.Error(user.Lang, "EmailAlreadyExists"))
			return
		}

		err = service.ChangeSettings(user.ID, "email", e.Email)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(user.Lang, "SomethingWentWrongInternal"))
		}

		ctx.JSON(http.StatusOK, presentation.Success(user.Lang, "EmailChanged"))
	}
}

func startup(service user.UseCase, coupleMessage repository.CoupleMessage) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userSession := sessions.Default(ctx).Get("user").(entity.UserSession)
		user, err := service.GetUser(userSession.Name)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(userSession.Lang, "SomethingWentWrong"))
			return
		}
		startup, err := service.StartupInfo(user.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(userSession.Lang, "SomethingWentWrong"))
			return
		}
		var newMessages int64
		if startup.HasPartner {
			newMessages, _ = coupleMessage.NewMessages(user.ID, user.CoupleID)
		}
		startup.NewMessages = newMessages
		ctx.JSON(http.StatusOK, startup)
	}
}

func clearNewNotifsCount(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userSession := sessions.Default(ctx).Get("user").(entity.UserSession)
		_ = service.ClearNotifsCount(userSession.ID)
		ctx.JSON(http.StatusNoContent, gin.H{})
	}
}

func clearFeedPostsCount(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userSession := sessions.Default(ctx).Get("user").(entity.UserSession)
		_ = service.ClearFeedPostsCount(userSession.ID)
		ctx.JSON(http.StatusNoContent, gin.H{})
	}
}

func getPartner(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userSession := sessions.Default(ctx).Get("user").(entity.UserSession)
		u, _ := service.GetUser(userSession.Name)
		partner, err := service.ListUsers([]primitive.ObjectID{u.PartnerID})
		if err != nil || len(partner) == 0 {
			ctx.JSON(http.StatusNotFound, presentation.Error(userSession.Lang, entity.ErrUserNotFound.Error()))
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"user_name": partner[0].UserName, "profile_picture": partner[0].ProfilePicture})
	}
}

func getFeed(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userSession := sessions.Default(ctx).Get("user").(entity.UserSession)
		skip, err := strconv.Atoi(ctx.Param("skip"))

		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(userSession.Lang, "BadRequest"))
			return
		}

		u, err := service.GetUser(userSession.Name)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(userSession.Lang, "SomethingWentWrong"))
			return
		}

		feed, err := service.GetFeedPosts(u.ID, skip)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(userSession.Lang, "SomethingWentWrong"))
			return
		}

		for i := 0; i < len(feed); i++ {
			if feed[i].CoupleID == u.CoupleID {
				feed[i].IsThisCouple = true
			}
		}

		page := entity.Pagination{
			Next: skip + entity.Limit,
			End:  skip+entity.Limit >= len(u.FeedPosts),
		}
		ctx.JSON(http.StatusOK, gin.H{"feed": feed, "pagination": page})
	}
}

func checkAvailability(service user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userName := ctx.Param("name")
		if !validator.IsUserName(userName) || !validator.IsNotBadName(userName) {
			ctx.JSON(http.StatusOK, gin.H{"available": false})

		}
		res := service.CheckNameAvailability(userName)
		ctx.JSON(http.StatusOK, gin.H{"available": res})
	}
}

func exempt(service user.UseCase, coupleService couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userSession := sessions.Default(ctx).Get("user").(entity.UserSession)
		coupleName := ctx.Param("coupleName")
		if !validator.IsCoupleName(coupleName) {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(userSession.Lang, "BadRequest"))
			return
		}
		couple, err := coupleService.GetCouple(coupleName, primitive.ObjectID{})
		if err != nil {
			ctx.JSON(http.StatusNotFound, presentation.Error(userSession.Lang, "CoupleNotFound"))
			return
		}
		err = service.Exempt(userSession.ID, couple.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(userSession.Lang, "SomethingWentWrongInternal"))
			return
		}
		ctx.JSON(http.StatusNoContent, gin.H{})
	}
}

func passwordResetRequest(service user.UseCase, verifRepo repository.VerifyMongo) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		email := ctx.Param("email")
		lang := utils.GetLang("", ctx.Request.Header)
		if !validator.IsEmail(email) {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(lang, "BadRequest"))
			return
		}
		linkID := primitive.NewObjectID().Hex()
		_, err := service.CheckSignup("", email)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				ctx.JSON(http.StatusOK, gin.H{})
				return
			}
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrong"))
			return
		}

		err = verifRepo.RequestPasswordReset(email, linkID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrong"))
			return
		}

		err = myaws.SendEmail(email, linkID, "password", lang)
		count := 0
		for err != nil && count < 2 {
			err = myaws.SendEmail(email, linkID, "password", lang)
			count++
		}
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrong"))
		}
		ctx.JSON(http.StatusAccepted, gin.H{})
	}
}

func resetPassword(service user.UseCase, verifyRepo repository.VerifyMongo) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		p := struct {
			Password string `json:"password"`
		}{}
		lang := utils.GetLang("", ctx.Request.Header)
		linkID := ctx.Param("linkID")
		err := ctx.ShouldBindJSON(&p)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		if !validator.IsPassword(p.Password) {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(lang, "WrongPasswordFormat"))
			return
		}
		email, err := verifyRepo.GetResetPasswordEmail(linkID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		hash, err := password.Generate(p.Password)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		err = service.ResetPassword(email, hash)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		verifyRepo.Verified(linkID)
		ctx.JSON(http.StatusOK, gin.H{})
	}
}

func MakeUserHandlers(r *gin.Engine, service user.UseCase, coupleService couple.UseCase, postService post.UseCase, coupleMessage repository.CoupleMessage, verifyRepo repository.VerifyMongo) {
	r.POST("/user/u/signup", unverifiedSignup(service, verifyRepo)) //tested
	r.POST("/user/resend-verification-link/:id", resendEmailVerificationLink(verifyRepo))
	r.POST("/user/verify-signup/:id", signup(service, verifyRepo))
	r.POST("/user/u/login", login(service)) //tested
	r.POST("/user/request-password-reset/:email", passwordResetRequest(service, verifyRepo))
	r.POST("/user/reset-password/:linkID", resetPassword(service, verifyRepo))
	r.GET("/user/availability/:name", checkAvailability(service))
	r.GET("/user/:userName", getUser(service, coupleService)) //tested

	r.GET("/user/u/session", middlewares.Authenticate(), userSession(service))                             //tested
	r.GET("/user/search/:query", middlewares.Authenticate(), searchUsers(service))                         //tested
	r.GET("/user/following/:name/:skip", middlewares.Authenticate(), getFollowing(service, coupleService)) //tested
	r.GET("/user/u/pending-request", middlewares.Authenticate(), getPendingRequest(service))               //tested
	// r.GET("/user/messages/:skip", userMessages(service, userMessage))
	// r.GET("/user/c/messages/:coupleName/:skip", userToACoupleMessages(service, coupleService, userMessage))
	r.GET("/user/u/startup", middlewares.Authenticate(), startup(service, coupleMessage))                       //tested
	r.GET("/user/notifications/:skip", middlewares.Authenticate(), notifications(service, coupleService))       //tested
	r.GET("/user/u/partner", middlewares.Authenticate(), getPartner(service))                                   //tested
	r.GET("/user/feed/:skip", middlewares.Authenticate(), getFeed(service))                                     //tested
	r.POST("/user/logout", middlewares.Authenticate(), logout)                                                  //tested
	r.POST("/user/u/cancel-request", middlewares.Authenticate(), cancelRequest(service))                        //tested
	r.POST("/user/u/reject-request", middlewares.Authenticate(), rejectRequest(service))                        //tested
	r.POST("/user/couple-request/:userName", middlewares.Authenticate(), initiateRequest(service))              //tested
	r.POST("/user/follow/:coupleName", middlewares.Authenticate(), follow(service, coupleService, postService)) //tested
	r.POST("/user/unfollow/:coupleName", middlewares.Authenticate(), unfollow(service, coupleService))          //tested
	r.POST("/user/exempt/:coupleName", middlewares.Authenticate(), exempt(service, coupleService))
	r.PUT("/user/new-notifications", middlewares.Authenticate(), clearNewNotifsCount(service))        //tested
	r.PUT("/user/new-posts", middlewares.Authenticate(), clearFeedPostsCount(service))                //tested
	r.PUT("/user/name", middlewares.Authenticate(), changeUserName(service))                          //tested
	r.PUT("/user/password", middlewares.Authenticate(), changePassword(service))                      //tested
	r.PUT("/user/email", middlewares.Authenticate(), changeEmail(service))                            //tested
	r.PUT("/user/request-status/:status", middlewares.Authenticate(), changeRequestStatus(service))   //tested
	r.PUT("/user", middlewares.Authenticate(), updateUser(service))                                   //tested
	r.POST("/user/show-pictures/:index", middlewares.Authenticate(), updateShowPicture(service))      //tested
	r.PATCH("/user/settings/:setting/:newValue", middlewares.Authenticate(), changeSettings(service)) //tested
	r.POST("/user/profile-pic", middlewares.Authenticate(), updateUserProfilePic(service))            //tested
	r.DELETE("/user", middlewares.Authenticate(), deleteUser(service))
}
