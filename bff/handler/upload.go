package handler

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	userv1 "webook/api/proto/gen/user/v1"
	ijwt "webook/bff/handler/jwt"
	"webook/pkg/ginx"
	"webook/pkg/logger"
)

type OSSConfig struct {
	Endpoint     string
	AvatarBucket string
	CoverBucket  string
	ImageBucket  string
}

type UploadHandler struct {
	s3Client *s3.S3
	ossCfg   OSSConfig
	userSvc  userv1.UserServiceClient
	l        logger.LoggerV1
}

func NewUploadHandler(s3Client *s3.S3, ossCfg OSSConfig, userSvc userv1.UserServiceClient, l logger.LoggerV1) *UploadHandler {
	return &UploadHandler{s3Client: s3Client, ossCfg: ossCfg, userSvc: userSvc, l: l}
}

func (h *UploadHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/upload")
	g.POST("/avatar", ginx.WrapClaims(h.l, h.UploadAvatar))
	g.POST("/image", ginx.WrapClaims(h.l, h.UploadImage))
	g.POST("/cover", ginx.WrapClaims(h.l, h.UploadCover))
}

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

func (h *UploadHandler) UploadAvatar(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		return ginx.Result{Code: 4, Msg: "请选择文件"}, nil
	}
	defer file.Close()

	if header.Size > 2*1024*1024 {
		return ginx.Result{Code: 4, Msg: "头像文件不能超过2MB"}, nil
	}

	contentType := header.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		return ginx.Result{Code: 4, Msg: "只支持 jpg/png/webp/gif 格式"}, nil
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	key := fmt.Sprintf("%d/avatar%s", uc.UserId, ext)

	data, err := io.ReadAll(file)
	if err != nil {
		return ginx.Result{Code: 5, Msg: "读取文件失败"}, err
	}

	_, err = h.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(h.ossCfg.AvatarBucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "上传失败"}, err
	}

	url := fmt.Sprintf("%s/%s/%s", h.ossCfg.Endpoint, h.ossCfg.AvatarBucket, key)

	_, err = h.userSvc.EditNoSensitive(ctx.Request.Context(), &userv1.EditNoSensitiveRequest{
		User: &userv1.User{
			Id:        uc.UserId,
			AvatarUrl: url,
		},
	})
	if err != nil {
		h.l.Error("更新头像URL失败", logger.Error(err))
	}

	return ginx.Result{Code: 2, Msg: "OK", Data: map[string]string{"url": url}}, nil
}

func (h *UploadHandler) UploadImage(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		return ginx.Result{Code: 4, Msg: "请选择文件"}, nil
	}
	defer file.Close()

	if header.Size > 10*1024*1024 {
		return ginx.Result{Code: 4, Msg: "图片文件不能超过10MB"}, nil
	}

	contentType := header.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		return ginx.Result{Code: 4, Msg: "只支持 jpg/png/webp/gif 格式"}, nil
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	key := fmt.Sprintf("articles/%s%s", uuid.New().String(), ext)

	data, err := io.ReadAll(file)
	if err != nil {
		return ginx.Result{Code: 5, Msg: "读取文件失败"}, err
	}

	_, err = h.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(h.ossCfg.ImageBucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "上传失败"}, err
	}

	url := fmt.Sprintf("%s/%s/%s", h.ossCfg.Endpoint, h.ossCfg.ImageBucket, key)
	return ginx.Result{Code: 2, Msg: "OK", Data: map[string]string{"url": url}}, nil
}

func (h *UploadHandler) UploadCover(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		return ginx.Result{Code: 4, Msg: "请选择文件"}, nil
	}
	defer file.Close()

	articleIdStr := ctx.Query("article_id")
	articleId, _ := strconv.ParseInt(articleIdStr, 10, 64)
	if articleId <= 0 {
		return ginx.Result{Code: 4, Msg: "article_id 参数错误"}, nil
	}

	if header.Size > 5*1024*1024 {
		return ginx.Result{Code: 4, Msg: "封面文件不能超过5MB"}, nil
	}

	contentType := header.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		return ginx.Result{Code: 4, Msg: "只支持 jpg/png/webp/gif 格式"}, nil
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	key := fmt.Sprintf("articles/%d/cover%s", articleId, ext)

	data, err := io.ReadAll(file)
	if err != nil {
		return ginx.Result{Code: 5, Msg: "读取文件失败"}, err
	}

	_, err = h.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(h.ossCfg.CoverBucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "上传失败"}, err
	}

	url := fmt.Sprintf("%s/%s/%s", h.ossCfg.Endpoint, h.ossCfg.CoverBucket, key)
	return ginx.Result{Code: 2, Msg: "OK", Data: map[string]string{"url": url}}, nil
}
