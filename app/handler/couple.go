package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/entity"
	"github.com/dawkaka/theone/pkg/myaws"
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

		coupleID, err := service.WhereACouple(user.ID, partner.ID)
		var coupleName string
		if err == nil { //used to have a couple profile
			err = service.MakeUp(coupleID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrong"))
				return
			}

		} else {
			coupleName = fmt.Sprintf("%s.and.%s", strings.ToLower(partner.FirstName), strings.ToLower(user.FirstName))
			_, err = service.GetCouple(coupleName, primitive.NewObjectID())
			if err == nil {
				coupleName += fmt.Sprint(time.Now().Unix())
			} else if err != entity.ErrCoupleNotFound {
				ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
				return
			}
			coupleID, err = service.CreateCouple(userb.ID.Hex(), partnerID.Hex(), coupleName, partner.Country, partner.State)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
				return
			}
		}

		notif := entity.Notification{
			Type:    "Request Accepted",
			Profile: user.ProfilePicture,
			User:    user.UserName,
			Name:    coupleName,
			Date:    time.Now(),
		}

		go func() {
			_ = userService.NotifyUser(partner.UserName, notif)
		}()
		err = userService.NewCouple([2]entity.ID{userb.ID, partnerID}, coupleID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}

		userb.CoupleID = coupleID
		userb.HasPartner = true
		userb.PendingRequest = entity.NO_REQUEST
		session.Set("user", userb)
		session.Save()
		ctx.SetCookie("couple_ID", coupleID.Hex(), 500, "/", "", false, true)
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
		couple, err := service.GetCouple(coupleName, user.ID)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusNotFound, presentation.Error(lang, "CoupleNotFound"))
				return
			}
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}

		pCouple := presentation.CoupleProfile{
			CoupleName:     couple.CoupleName,
			AcceptedAt:     couple.AcceptedAt,
			Bio:            couple.Bio,
			FollowersCount: couple.FollowersCount,
			ProfilePicture: couple.ProfilePicture,
			CoverPicture:   couple.CoverPicture,
			PostCount:      couple.PostCount,
			Married:        couple.Married,
			Verified:       couple.Verified,
			Website:        couple.Website,
			DateCommenced:  couple.DateCommenced,
			IsThisCouple:   user.ID == couple.Initiated || user.ID == couple.Accepted,
			IsFollowing:    couple.IsFollowing,
		}
		ctx.JSON(http.StatusOK, pCouple)
	}
}

func getCouplePosts(service couple.UseCase, postService post.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		coupleName, skip := ctx.Param("coupleName"), ctx.Param("skip")
		skipPosts, err := strconv.Atoi(skip)
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := user.Lang
		if !validator.IsCoupleName(coupleName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		couple, err := service.GetCouplePosts(coupleName, skipPosts)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		st, ed := len(couple.Posts)-skipPosts-entity.LimitP, len(couple.Posts)-skipPosts

		if st < 0 {
			st = 0
		}
		if ed < 0 {
			ed = 0
		}

		postIDs := couple.Posts[st:ed]
		posts, err := postService.GetPosts(couple.ID, user.ID, postIDs)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrong"))
			return
		}

		for i := 0; i < len(posts); i++ {
			posts[i].CoupleName = couple.CoupleName
			posts[i].IsThisCouple = couple.Initiated == user.ID || couple.Accepted == user.ID
			posts[i].Verified = couple.Verified
			posts[i].Married = couple.Married
			posts[i].ProfilePicture = couple.ProfilePicture
		}
		page := entity.Pagination{
			Next: skipPosts + entity.LimitP,
			End:  len(postIDs) < entity.LimitP,
		}
		ctx.JSON(http.StatusOK, gin.H{"posts": posts, "pagination": page})
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
			return
		}
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		videos, err := service.GetCoupleVideos(coupleName, skipVideos)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		page := entity.Pagination{
			Next: skipVideos + entity.Limit,
			End:  len(videos) < entity.Limit,
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"videos": videos, "pagination": page})
	}
}

func getFollowers(service couple.UseCase, userService user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		coupleName := ctx.Param("coupleName")
		skip, err := strconv.Atoi(ctx.Param("skip")) //pagination
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(lang, "BadRequest"))
			return
		}
		if !validator.IsCoupleName(coupleName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		followers, err := service.GetFollowers(coupleName, skip)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		users, err := userService.ListFollowers(followers)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		page := entity.Pagination{
			Next: skip + entity.Limit,
			End:  len(followers) < entity.Limit,
		}
		ctx.JSON(http.StatusOK, gin.H{"followers": users, "pagination": page})
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
		fileName, err := myaws.UploadImageFile(fileHeader, "theone-profile-images")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "ProfilePicFailed"))
			return
		}
		err = service.UpdateCoupleProfilePic(fileName, user.CoupleID)
		if err != nil {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "Forbidden"))
			return
		}
		ctx.JSON(http.StatusCreated, presentation.Success(lang, "ProfilePicUpdated"))
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
		fileName, err := myaws.UploadImageFile(fileHeader, "theone-profile-images")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "ProfilePicFailed"))
			return
		}
		err = service.UpdateCoupleCoverPic(fileName, user.CoupleID)
		if err != nil {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "Forbidden"))
			return
		}
		ctx.JSON(http.StatusCreated, presentation.Success(lang, "CoverPicUpdated"))
	}
}

func updateCouple(service couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := user.Lang
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
		nameStruct := struct {
			CoupleName string `json:"couple_name"`
		}{}
		session := sessions.Default(ctx)
		user := session.Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		err := ctx.ShouldBindJSON(&nameStruct)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		newCoupleName := nameStruct.CoupleName
		if !validator.IsCoupleName(newCoupleName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "InvalidCoupleName"))
			return
		}
		_, err = service.GetCouple(newCoupleName, primitive.NewObjectID())
		if err == nil {
			ctx.JSON(http.StatusConflict, presentation.Error(lang, "CoupleAlreadyExists"))
			return
		} else {
			if err != mongo.ErrNoDocuments {
				ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
				return
			}
		}
		err = service.ChangeCoupleName(user.CoupleID, newCoupleName)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		ctx.JSON(http.StatusCreated, presentation.Success(lang, "ChangedCoupleName"))
	}
}

func lastLastEdonCast(service couple.UseCase, userService user.UseCase) gin.HandlerFunc { //Na everybody go chop breakfast
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		user := session.Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		u, err := userService.GetUser(user.Name)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		if !u.HasPartner {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "IsSingle"))
			return
		}
		err = service.BreakUp(u.CoupleID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}

		go func() {
			notif := entity.Notification{
				Type:    "Break Up",
				Profile: user.ProfilePicture,
				User:    user.Name,
				Name:    user.FirstName + user.LastName,
				Date:    time.Now(),
			}
			userService.BreakedUp([2]entity.ID{u.ID, u.PartnerID})
			_ = userService.NotifyCouple([2]entity.ID{u.PartnerID, primitive.NewObjectID()}, notif)
		}()
		user.HasPartner = false
		user.PartnerID = primitive.ObjectID{}
		user.CoupleID = primitive.ObjectID{}
		session.Set("user", user)
		session.Save()
		ctx.SetCookie("couple_ID", "", -33, "/", "", false, true)
		ctx.JSON(http.StatusOK, presentation.Success(lang, "BreakedUp"))
	}
}

//Messages between partners
func coupleMessages(coupleMessage repository.CoupleMessage, userService user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		skip, err := strconv.Atoi(ctx.Param("skip"))
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(lang, "BadRequest"))
			return
		}
		u, err := userService.GetUser(user.Name)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrong"))
			return
		}

		messages, err := coupleMessage.Get(u.CoupleID, skip)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		page := entity.Pagination{
			Next: skip + entity.Limit,
			End:  len(messages) < entity.Limit,
		}
		ctx.JSON(http.StatusOK, gin.H{"messages": messages, "pagination": page})
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

func reportCouple(reportRepo repository.Reports) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		r := struct {
			Reports []int `json:"reports"`
		}{Reports: []int{}}

		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		coupleID, err := entity.StringToID(ctx.Param("postID"))
		err2 := ctx.ShouldBindJSON(&r)
		lang := user.Lang
		if err2 != nil || err != nil || !validator.IsValidPostReport(r.Reports) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		report := entity.ReportCouple{CoupleID: coupleID, UserID: user.ID, Report: r.Reports, CreatedAt: time.Now(), Type: "couple"}
		err = reportRepo.ReportCouple(report)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		ctx.JSON(http.StatusCreated, presentation.Success(lang, "PostReported"))
	}
}

func searchCouples(service couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		query := ctx.Param("query")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)

		//TODO:
		// if !validator.IsUserName(query) && !validator.IsRealName(query) {
		// 	ctx.JSON(http.StatusBadRequest, presentation.Error(user.Lang, "WrongUserNameFormat"))
		// 	return
		// }
		couples, err := service.SearchCouples(query)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(user.Lang, "SomethingWentWrongInternal"))
			return
		}
		if couples == nil {
			ctx.JSON(http.StatusOK, []string{})
			return
		}
		ctx.JSON(http.StatusOK, couples)
	}
}

func updateRelationshipStatus(service couple.UseCase, userService user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		status := ctx.Param("status")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		if status != "YES" && status != "NO" {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(user.Lang, "BadRequest"))
			return
		}
		u, err := userService.GetUser(user.Name)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(user.Lang, "SomethingWentWrong"))
			return
		}
		var married bool
		if status == "YES" {
			married = true
		}
		err = service.UpdateStatus(u.CoupleID, married)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(user.Lang, "SomethingWentWrong"))
			return
		}
		ctx.JSON(http.StatusOK, presentation.Success(user.Lang, "StatusUpdated"))
	}
}

func getSuggestedAccounts(service couple.UseCase, userService user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userSession := sessions.Default(ctx).Get("user").(entity.UserSession)
		res, err := userService.ExemptedFromSuggestedAccounts(userSession.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(userSession.Lang, "SomethingWentWrongInternal"))
			return
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func MakeCoupleHandlers(r *gin.Engine, service couple.UseCase, userService user.UseCase, postService post.UseCase,
	coupleMessage repository.CoupleMessage, userMessage repository.UserCoupleMessage, reportRepo repository.Reports) {
	r.GET("/:coupleName", getCouple(service)) //tested
	r.GET("/:coupleName/posts/:skip", getCouplePosts(service, postService))
	r.GET("/:coupleName/videos/:skip", getCoupleVideos(service))
	r.GET("/couple/search/:query", searchCouples(service))
	r.GET("/:coupleName/followers/:skip", getFollowers(service, userService)) //tested
	r.GET("/couple/p-messages/:skip", coupleMessages(coupleMessage, userService))
	r.GET("/couple/u/suggested-accounts", getSuggestedAccounts(service, userService))
	r.GET("/couple/messages/:skip", usersCoupleMessages(userMessage))
	r.GET("/couple/u/messages/:userName/:skip", userCoupleMessages(service, userService, userMessage))
	r.POST("/couple/new/:partnerID", newCouple(service, userService))  //tested
	r.POST("/couple/break-up", lastLastEdonCast(service, userService)) //tested
	r.POST("/couple/report", reportCouple(reportRepo))
	r.POST("/couple/profile-picture", updateCoupleProfilePic(service)) //tested
	r.POST("/couple/cover-picture", updateCoupleCoverPic(service))     //tested
	r.POST("/couple/name", changeCoupleName(service))                  //tested
	r.POST("/couple/status/:status", updateRelationshipStatus(service, userService))
	r.PUT("/couple", updateCouple(service)) //tested
}
