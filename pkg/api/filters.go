// TODO - refactor to work with testing

package api

import (
	"net/http"
	"strings"

	"github.com/AustinBayley/activity_tracker_api/pkg/service"
	"github.com/gin-gonic/gin"
)

const (
	UserCtxKey string = "userContext"
)

type RequestContext struct {
	UserID service.ID `json:"userID"`
	Admin  bool       `json:"admin"`
	Email  string     `json:"email"`
}

// ActorFilter updates the context with details of the user making the request.
func (a *API) ActorFilter(req *gin.Context) {
	email := req.GetHeader("X-Auth-Request-Email")
	groups := req.GetHeader("X-Auth-Request-Groups")

	req.Set(UserCtxKey, RequestContext{
		Email: email,
	})

	if email == "" {
		req.Next()
		return
	}

	// We need to get the ID of the user but don't want to do anything with an error
	// as some routes do not require a user to be authenticated.
	user, _ := a.users.GetByEmail(req, email)

	if user == nil {
		req.Next()
		return
	}

	req.Set(UserCtxKey, RequestContext{
		UserID: user.ID,
		Admin:  strings.Contains(groups, a.adminGroup) && user.ID != "",
		Email:  email,
	})

	req.Next()
}

func GetActorContext(req *gin.Context) (RequestContext, bool) {
	rawCtx, ok := req.Get(UserCtxKey)
	if !ok {
		return RequestContext{}, false
	}

	reqCtx, ok := rawCtx.(RequestContext)
	if !ok {
		return RequestContext{}, false
	}

	return reqCtx, true
}

// Only allow requests with tokens
func (a *API) HasAuthFilter(req *gin.Context) {
	ctx, ok := GetActorContext(req)
	if !ok {
		req.JSON(http.StatusUnauthorized, ErrorResponse{
			Cause: NotAuthorised,
		})
		return
	}

	if ctx.UserID == "" {
		req.JSON(http.StatusUnauthorized, ErrorResponse{
			Cause: NotAuthorised,
		})
		return
	}

	req.Next()
}

// Only allow admin users
// Return not found if the user is not an admin so as not to expose information
// about admin routes.
func (a *API) AdminAuthFilter(req *gin.Context) {
	reqCtx, ok := GetActorContext(req)
	if !ok {
		req.JSON(http.StatusNotFound, ErrorResponse{
			Cause: NotFound,
		})
		return
	}

	if !reqCtx.Admin {
		req.JSON(http.StatusNotFound, ErrorResponse{
			Cause: NotFound,
		})
		return
	}

	req.Next()
}
