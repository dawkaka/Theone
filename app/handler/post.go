package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/entity"
	"github.com/dawkaka/theone/inter"
	"github.com/dawkaka/theone/pkg/myaws"
	"github.com/dawkaka/theone/pkg/utils"
	"github.com/dawkaka/theone/pkg/validator"
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
		files := form.File["post_image"]
		caption := strings.TrimSpace(ctx.PostForm("caption"))
		fmt.Println(form.Value)
		coupleName := strings.TrimSpace(ctx.PostForm("couple_name"))
		if !validator.IsCaption(caption) || err != nil || !validator.IsCoupleName(coupleName) {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		if !user.HasPartner {
			ctx.JSON(http.StatusForbidden, presentation.Error(lang, "OnlyCoupleCanPost"))
		}
		filesMetadata, err := myaws.UploadMultipleFiles(files)
		fmt.Println(err)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		mentions := utils.ExtracMentions(caption)
		post := entity.Post{
			PostID:      utils.GenerateID(),
			CoupleID:    user.CoupleID,
			InitiatedID: user.ID,
			AcceptedID:  user.PartnerID,
			PostedBy:    user.ID,
			Files:       filesMetadata,
			Caption:     caption,
			Mentioned:   mentions,
			CreatedAt:   time.Now(),
			Likes:       []entity.ID{},
			Comments:    []entity.Comment{},
		}
		_, err = service.CreatePost(&post)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		//Post error created, whether notifications are successful or not user does't need to know
		go func() {
			notif := entity.Notification{
				Type:       "PostMentioned",
				Message:    inter.LocalizeWithUserName(lang, coupleName, "PostMentionedNotif"),
				PostID:     post.PostID,
				CoupleName: coupleName,
			}
			partnerNotif := entity.Notification{
				Type:       "PartnerPosted",
				Message:    inter.LocalizeWithUserName(lang, user.Name, "PartnerNewPostNotif"),
				PostID:     post.PostID,
				CoupleName: coupleName,
			}
			userService.NotifyCouple([2]entity.ID{user.PartnerID, primitive.NewObjectID()}, partnerNotif)
			if len(mentions) > 0 {
				userService.NotifyMultipleUsers(mentions, notif)
			}

		}()

		ctx.JSON(http.StatusCreated, presentation.Success(lang, "NewPostAdded"))
	}

}

func getPost(service post.UseCase, coupleService couple.UseCase) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		coupleName, postID := ctx.Param("coupleName"), ctx.Param("postID")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := user.Lang
		if !validator.IsCoupleName(coupleName) || strings.TrimSpace(postID) == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		couple, err := coupleService.GetCouple(coupleName)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "UserNotFound"))
				return
			}
		}
		post, err := service.GetPost(couple.ID.Hex(), postID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
			return
		}
		p := presentation.Post{
			CoupleName:    couple.CoupleName,
			ID:            post.ID,
			CreatedAt:     post.CreatedAt,
			Caption:       post.Caption,
			LikesCount:    post.LikesCount,
			CommentsCount: post.CommentsCount,
			Files:         post.Files,
		}
		ctx.JSON(http.StatusOK, gin.H{"post": p})
	}
}

func newComment(service post.UseCase, userService user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		postID := ctx.Param("postID")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if strings.TrimSpace(postID) == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		post, err := service.GetPostByID(postID)
		fmt.Println(err)
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
		fmt.Println(err)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		fmt.Println(strings.TrimSpace(comment.Comment))
		if !validator.IsCaption(comment.Comment) {
			ctx.JSON(http.StatusUnprocessableEntity, "InvalidComment")
			return
		}
		comment.UserID = user.ID
		comment.ID = primitive.NewObjectID()
		comment.CreatedAt = time.Now()
		comment.Likes = []entity.ID{}
		err = service.NewComment(postID, comment)
		fmt.Println(err)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		notif := entity.Notification{
			Type:    "comment",
			Message: inter.LocalizeWithUserName(lang, user.Name, "PostCommentNotif"),
		}
		_ = userService.NotifyCouple([2]entity.ID{post.InitiatedID, post.AcceptedID}, notif)

		ctx.JSON(http.StatusCreated, presentation.Success(lang, "CommentAdded"))
	}
}

func like(service post.UseCase, userService user.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		postID := ctx.Param("postID")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if strings.TrimSpace(postID) == "" {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		post, err := service.GetPostByID(postID)
		fmt.Println(err)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusNotFound, presentation.Error(lang, "NotFoundComment"))
				return
			}
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}

		err = service.LikePost(postID, user.ID.Hex())
		fmt.Println(err)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, presentation.Error(lang, "SomethingWentWrongInternal"))
			return
		}
		notif := entity.Notification{
			Type:    "like",
			Message: inter.LocalizeWithUserName(lang, user.Name, "PostLikeNotif"),
		}
		_ = userService.NotifyCouple([2]entity.ID{post.InitiatedID, post.AcceptedID}, notif)
		ctx.JSON(http.StatusCreated, presentation.Success(lang, "PostLiked"))
	}
}

func postComments(service post.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		skip, err := strconv.Atoi(ctx.Param("skip"))
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		postID := ctx.Param("postID")
		comments, err := service.GetComments(postID, skip)
		fmt.Println(err)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "SomethingWentWrong"))
			return
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

func editPostCaption(service post.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		postID := ctx.Param("postID")
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		lang := utils.GetLang(user.Lang, ctx.Request.Header)
		var Caption struct {
			Caption string `json:"caption"`
		}
		err := ctx.ShouldBindJSON(&Caption)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
			return
		}
		err = service.EditCaption(postID, user.CoupleID, Caption.Caption)
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

func deletePost(service post.UseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := sessions.Default(ctx).Get("user").(entity.UserSession)
		postID := ctx.Param("postID")
		lang := user.Lang
		if postID == "" {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, presentation.Error(lang, "BadRequest"))
		}
		err := service.DeletePost(user.CoupleID, postID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusForbidden, "Forbidden")
			}
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, "SomethingWentWrongInternal")
		}
		ctx.JSON(http.StatusOK, presentation.Success(lang, "PostDeleted"))
	}
}

func MakePostHandlers(r *gin.Engine, service post.UseCase, coupleService couple.UseCase, userService user.UseCase) {
	r.GET("/post/:coupleName/:postID", getPost(service, coupleService))      //tested
	r.GET("/post/comments/:postID/:skip", postComments(service))             //tested
	r.POST("/post", newPost(service, coupleService, userService))            //tested
	r.POST("/post/comment/:postID", newComment(service, userService))        //tested
	r.DELETE("/post/comment/:postID/:commentID", deletePostComment(service)) //tested
	r.PATCH("/post/like/:postID", like(service, userService))                //tested
	r.PATCH("/post/unlike/:postID", unLikePost(service))                     //tested
	r.PATCH("/post/edit/:postID", editPostCaption(service))                  //tested
	r.PATCH("/post/comment/like/:postID/:commentID", likeComment(service))
	r.PATCH("/post/comment/unlike/:postID/:commentID", unlikeComment(service))
	r.DELETE("/post/:postID", deletePost(service))
}
