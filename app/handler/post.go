package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dawkaka/theone/app/middlewares"
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

func newPost(service post.UseCase, coupleService couple.UseCase, userService user.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := user.Lang
		form, err := ctx.MultipartForm()
		files := form.File["files"]
		caption := strings.TrimSpace(ctx.PostForm("caption"))
		location := strings.TrimSpace(ctx.PostForm("location"))
		alts := [10]string{}

		if !validator.IsCaption(caption) || err != nil || len(location) > 50 || len(files) == 0 || len(files) > 10 {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}

		_ = json.Unmarshal([]byte(ctx.PostForm("alts")), &alts)
		u, err := userService.GetUser(user.Name)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		if !u.HasPartner {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "OnlyCoupleCanPost"))
			return
		}
		filesMetadata, cErr := myaws.UploadMultipleFiles(files)
		if cErr != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
		}
		for i := 0; i < len(filesMetadata); i++ {
			filesMetadata[i].Alt = alts[i]
		}

		if cErr != nil {

			ctx.JSON(cErr.Code, presentation.Error(lang, cErr.Error()))
			return
		}

		mentions := utils.ExtracMentions(caption)
		postID := utils.GenerateID()
		post := entity.Post{
			PostID:      postID,
			CoupleID:    u.CoupleID,
			InitiatedID: user.ID,
			AcceptedID:  u.PartnerID,
			PostedBy:    u.ID,
			Files:       filesMetadata,
			Caption:     caption,
			Location:    location,
			Mentioned:   mentions,
			CreatedAt:   time.Now(),
			Likes:       []entity.ID{},
			Comments:    []entity.Comment{},
		}
		pID, err := service.CreatePost(&post)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}

		err = coupleService.AddPost(u.CoupleID, postID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}

		//Post created, whether notifications are successful or not user does't need to know
		go func() {

			notif := entity.Notification{
				Type:     "Mentioned",
				Message:  caption,
				PostID:   post.PostID,
				CoupleID: u.CoupleID,
				Date:     time.Now(),
			}
			partnerNotif := entity.Notification{
				Type:     "Partner Posted",
				Message:  caption,
				PostID:   post.PostID,
				CoupleID: u.CoupleID,
				UserID:   user.ID,
				Date:     time.Now(),
			}
			userService.NotifyCouple([2]entity.ID{u.PartnerID, primitive.NewObjectID()}, partnerNotif)
			if len(mentions) > 0 {
				userService.NotifyMultipleUsers(mentions, notif)
			}

		}()

		go func() {
			skip := 0
			for {
				followers, err := coupleService.FollowersToNotify(post.CoupleID, skip)
				if err != nil {
					break
				}
				err = userService.NewFeedPost(pID, followers)
				if err != nil {
					break
				}
				if len(followers) < 1000 {
					break
				}
				skip += 1000
			}
		}()

		ctx.JSON(http.StatusCreated, presentation.Success(lang, "NewPostAdded"))
	}
}

func getPost(service post.UseCase, coupleService couple.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		coupleName, postID := ctx.Param("coupleName"), ctx.Param("postID")
		user := utils.GetSession(sessions.Default(ctx))
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if !validator.IsCoupleName(coupleName) || strings.TrimSpace(postID) == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		couple, err := coupleService.GetCouple(coupleName, user.ID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "UserNotFound"))
				return
			}
		}
		post, err := service.GetPost(couple.ID.Hex(), user.ID.Hex(), postID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
			return
		}

		p := presentation.Post{
			CoupleName:     couple.CoupleName,
			Married:        couple.Married,
			Verified:       couple.Verified,
			ProfilePicture: couple.ProfilePicture,
			ID:             post.ID,
			CreatedAt:      post.CreatedAt,
			Caption:        post.Caption,
			LikesCount:     post.LikesCount,
			HasLiked:       post.HasLiked,
			CommentsCount:  post.CommentsCount,
			Files:          post.Files,
			IsThisCouple:   user.ID == couple.Initiated || user.ID == couple.Accepted,
			Location:       post.Location,
			CommentsClosed: post.CommentsClosed,
		}
		ctx.JSON(http.StatusOK, p)
	}
}

func newComment(service post.UseCase, userService user.UseCase, coupleService couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		postID := ctx.Param("postID")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)

		if strings.TrimSpace(postID) == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		post, err := service.GetPostByID(postID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusNotFound, presentation.Error(lang, "NotFoundComment"))
				return
			}
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		var comment entity.Comment
		err = ctx.ShouldBindJSON(&comment)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		if !validator.IsCaption(comment.Comment) {
			ctx.JSON(http.StatusUnprocessableEntity, "InvalidComment")
			return
		}
		comment.UserID = user.ID
		comment.ID = primitive.NewObjectID()
		comment.CreatedAt = time.Now()
		comment.Likes = []entity.ID{}
		err = service.NewComment(postID, comment)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		u, _ := userService.GetUser(user.Name)
		go func() {
			notif := entity.Notification{
				Type:     "comment",
				Message:  comment.Comment,
				PostID:   post.PostID,
				CoupleID: post.CoupleID,
				UserID:   user.ID,
				Date:     time.Now(),
			}
			_ = userService.NotifyCouple([2]entity.ID{post.InitiatedID, post.AcceptedID}, notif)
		}()

		ctx.JSON(http.StatusCreated, gin.H{
			"comment": presentation.Comment{
				Comment:        comment,
				HasPartner:     u.HasPartner,
				ProfilePicture: u.ProfilePicture,
				UserName:       u.UserName,
				LikesCount:     0,
				HasLiked:       false,
			},
			"notif": presentation.Success(lang, "CommentAdded"),
		})
	}
}

func like(service post.UseCase, userService user.UseCase, coupleService couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		postID := ctx.Param("postID")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if strings.TrimSpace(postID) == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		post, err := service.GetPostByID(postID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusNotFound, presentation.Error(lang, "NotFoundComment"))
				return
			}
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}

		err = service.LikePost(postID, user.ID.Hex())
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		go func() {
			notif := entity.Notification{
				Type:     "like",
				PostID:   post.PostID,
				UserID:   user.ID,
				CoupleID: post.CoupleID,
				Date:     time.Now(),
			}
			_ = userService.NotifyCouple([2]entity.ID{post.InitiatedID, post.AcceptedID}, notif)

		}()
		ctx.JSON(http.StatusCreated, presentation.Success(lang, "PostLiked"))
	}
}

func postComments(service post.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		skip, err := strconv.Atoi(ctx.Param("skip"))
		user := utils.GetSession(sessions.Default(ctx))
		lang := utils.GetLang(user.Lang, ctx.Request.Header)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		postID := ctx.Param("postID")
		comments, err := service.GetComments(postID, user.ID.Hex(), skip)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		if len(comments) == 0 {
			comments = []presentation.Comment{}
		}
		page := entity.Pagination{
			Next: skip + entity.Limit,
			End:  len(comments) < entity.Limit,
		}
		ctx.JSON(http.StatusOK, gin.H{"comments": comments, "pagination": page})
	}
}

func deletePostComment(service post.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		commentID, postID := ctx.Param("commentID"), ctx.Param("postID")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := user.Lang
		if strings.TrimSpace(commentID) == "" || strings.TrimSpace(postID) == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		userID := user.ID
		err := service.DeleteComment(postID, commentID, userID)
		if err != nil {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "Forbidden"))
			return
		}
		ctx.JSON(http.StatusOK, presentation.Success(lang, "CommentDeleted"))
	}
}

func unLikePost(service post.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		postID := ctx.Param("postID")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if strings.TrimSpace(postID) == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		userID := user.ID
		err := service.UnLikePost(postID, userID)
		if err != nil {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "Forbidden"))
			return
		}
		ctx.JSON(http.StatusOK, presentation.Success(lang, "UnlikePost"))
	}
}

func editPost(service post.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		postID := ctx.Param("postID")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		edit := entity.EditPost{}
		err := ctx.ShouldBindJSON(&edit)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		err = service.EditPost(postID, user.CoupleID, edit)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "Forbidden"))
			return
		}
		ctx.JSON(http.StatusOK, presentation.Success(lang, "PostEdited"))
	}
}

func likeComment(service post.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		commentID, postID := ctx.Param("commentID"), ctx.Param("postID")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := user.Lang
		if strings.TrimSpace(commentID) == "" || strings.TrimSpace(postID) == "" {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(lang, "BadRequest"))
			return
		}
		pID, err := entity.StringToID(postID)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(lang, "BadRequest"))
			return
		}

		cID, err := entity.StringToID(commentID)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(lang, "BadRequest"))
			return
		}

		err = service.LikeComment(pID, cID, user.ID)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, "SomethingWentWrong")
			return
		}
		ctx.JSON(http.StatusCreated, presentation.Success(lang, "PostLiked"))
	}
}

func unlikeComment(service post.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		commentID, postID := ctx.Param("commentID"), ctx.Param("postID")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := user.Lang
		if strings.TrimSpace(commentID) == "" || strings.TrimSpace(postID) == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		pID, err := entity.StringToID(postID)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(lang, "BadRequest"))
			return
		}

		cID, err := entity.StringToID(commentID)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(lang, "BadRequest"))
			return
		}
		err = service.UnLikeComment(pID, cID, user.ID)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, "SomethingWentWrong")
			return
		}
		ctx.JSON(http.StatusCreated, presentation.Success(lang, "UnlikePost"))
	}
}

func deletePost(service post.UseCase, coupleService couple.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		postID := ctx.Param("postID")
		lang := user.Lang
		if postID == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		post, err := service.GetPostByID(postID)

		if err != nil {
			ctx.JSON(http.StatusForbidden, "Forbidden")
			return
		}
		err = service.DeletePost(user.CoupleID, postID)

		if err != nil {
			if err == entity.ErrNotFound {
				ctx.JSON(http.StatusForbidden, "Forbidden")
				return
			}
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		err = coupleService.RemovePost(user.CoupleID, postID)
		count := 0
		for err != nil && count < 2 {
			err = coupleService.RemovePost(user.CoupleID, postID)
			count++
		}

		for i := 0; i < len(post.Files); i++ {

			go func(key string) {
				count = 0
				err = myaws.DeleteFile(key, "theone-profile-images")
				for err != nil && count < 2 {
					err = myaws.DeleteFile(key, "theone-profile-images")
					count++
				}
			}(post.Files[i].Name)

		}

		ctx.JSON(http.StatusOK, presentation.Success(lang, "PostDeleted"))
	}
}

func reportPost(service post.UseCase, reportRepo repository.Reports) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		r := struct {
			Reports []int `json:"reports"`
		}{Reports: []int{}}

		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		postID, err := entity.StringToID(ctx.Param("postID"))
		err2 := ctx.ShouldBindJSON(&r)
		lang := user.Lang
		if err2 != nil || err != nil || !validator.IsValidPostReport(r.Reports) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		_, err = service.GetPostByID(postID.Hex())

		if err != nil {
			ctx.JSON(http.StatusNotFound, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		report := entity.ReportPost{PostID: postID, UserID: user.ID, Reports: r.Reports, CreatedAt: time.Now(), Type: "post"}
		err = reportRepo.ReportPost(report)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		ctx.JSON(http.StatusCreated, presentation.Success(lang, "PostReported"))
	}
}

func explorePosts(service post.UseCase, userService user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		coupleIDs, err := userService.ExemptedFromSuggestedAccounts(user.ID, false)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(user.Lang, "SomethingWentWrongInternal"))
			return
		}
		skip, err := strconv.Atoi(ctx.Param("skip"))
		if err != nil {
			skip = 0
		}
		posts, err := service.GetExplorePosts(coupleIDs, user.ID, user.Country, skip)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(user.Lang, "SomethingWentWrongInternal"))
			return
		}
		page := entity.Pagination{
			Next: skip + entity.LimitP,
			End:  len(posts) < entity.LimitP,
		}
		ctx.JSON(http.StatusOK, gin.H{"posts": posts, "pagination": page})
	}
}

func setCloseComments(service post.UseCase, userService user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		postID := ctx.Param("postID")
		switched := ctx.Param("switched")
		pID, err := entity.StringToID(postID)
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		if err != nil || (switched != "ON" && switched != "OFF") {
			ctx.JSON(http.StatusUnprocessableEntity, presentation.Error(user.Lang, "BadRequest"))
			return
		}
		var state bool
		if switched == "ON" {
			state = false
		} else {
			state = true
		}
		u, err := userService.GetUser(user.Name)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(user.Lang, "SomethingWentWrongInternal"))
			return
		}
		err = service.SetClosedComments(pID, u.CoupleID, state)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(user.Lang, "SomethingWentWrong"))
			return
		}
		ctx.JSON(http.StatusOK, presentation.Success(user.Lang, "Comments"+switched))
	}
}

func MakePostHandlers(r *gin.Engine, service post.UseCase, coupleService couple.UseCase, userService user.UseCase, reportsRepo repository.Reports) {
	r.GET("/post/:coupleName/:postID", middlewares.CheckBlocked(coupleService), getPost(service, coupleService)) //tested
	r.GET("/post/comments/:postID/:skip", postComments(service))                                                 //tested
	r.GET("/post/explore/:skip", middlewares.Authenticate(), explorePosts(service, userService))
	r.POST("/post", middlewares.Authenticate(), newPost(service, coupleService, userService)) //tested
	r.POST("/post/:postID/:switched", middlewares.Authenticate(), setCloseComments(service, userService))
	r.POST("/post/comment/:postID", middlewares.Authenticate(), newComment(service, userService, coupleService)) //tested
	r.POST("/post/report/:postID", middlewares.Authenticate(), reportPost(service, reportsRepo))
	r.PATCH("/post/like/:postID", middlewares.Authenticate(), like(service, userService, coupleService)) //tested
	r.PATCH("/post/unlike/:postID", middlewares.Authenticate(), unLikePost(service))                     //tested
	r.PUT("/post/:postID", middlewares.Authenticate(), editPost(service))                                //tested
	r.PATCH("/post/comment/like/:postID/:commentID", middlewares.Authenticate(), likeComment(service))
	r.PATCH("/post/comment/unlike/:postID/:commentID", middlewares.Authenticate(), unlikeComment(service))
	r.DELETE("/post/comment/:postID/:commentID", middlewares.Authenticate(), deletePostComment(service)) //tested
	r.DELETE("/post/:postID", middlewares.Authenticate(), deletePost(service, coupleService))
}
