package routes

import (
	"naevis/chunkedup"
	"naevis/droping"
	"naevis/feedproxy"
	"naevis/filedrop"
	"naevis/filemgr"
	"naevis/middleware"
	"naevis/posts"
	"naevis/ratelim"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func AddStaticRoutes(router *httprouter.Router) {
	router.ServeFiles("/static/uploads/*filepath", http.Dir("static/uploads"))
}

func AddFiledropRoutes(router *httprouter.Router, rateLimiter *ratelim.RateLimiter) {
	router.POST("/api/v1/filedrop", rateLimiter.Limit(filedrop.UploadHandler))

	router.POST("/filedrop", droping.FiledropHandler)
	router.OPTIONS("/filedrop", droping.OptionsHandler)
	// router.GET("/health", droping.HealthHandler)

	router.PUT("/picture/:entitytype/:entityid", rateLimiter.Limit(middleware.Authenticate(filemgr.EditBanner)))

	router.POST("/posts/upload", rateLimiter.Limit(posts.UploadImage))

	router.POST("/filedrop/uploads/chunk", rateLimiter.Limit(chunkedup.ChunkedUploads))
	router.HEAD("/filedrop/uploads/exists", chunkedup.FileExistsHandler)

	router.PUT("/profile/avatar", rateLimiter.Limit(middleware.Authenticate(filedrop.EditProfilePic)))

	router.PUT("/gallery/:entityType/:entityId/images", rateLimiter.Limit(middleware.Authenticate(filedrop.UpdateGalleryImages)))

	router.PUT("/feedproxy", rateLimiter.Limit(middleware.Authenticate(feedproxy.UpdateTweetPost)))
}
