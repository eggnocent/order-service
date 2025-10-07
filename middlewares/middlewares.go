package middlewares

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"order-service/clients"
	"order-service/common/response"
	"order-service/config"
	"order-service/constants"
	errConstant "order-service/constants/error"
	"runtime/debug"
	"strings"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func HandlePanic() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("🔥 Recovered from panic: %v\n%s", r, debug.Stack())
				c.JSON(http.StatusInternalServerError, response.Response{
					Status:  constants.Error,
					Message: errConstant.ErrInternalServerError.Error(),
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

func RateLimiter(lmt *limiter.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := tollbooth.LimitByRequest(lmt, c.Writer, c.Request)
		if err != nil {
			logrus.Warnf("🚦 Rate limit triggered: %v", err)
			c.JSON(http.StatusTooManyRequests, response.Response{
				Status:  constants.Error,
				Message: errConstant.ErrToManyRequests.Error(),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func extractBearerToken(token string) string {
	arrayToken := strings.Split(token, " ")
	if len(arrayToken) == 2 && strings.ToLower(arrayToken[0]) == "bearer" {
		return arrayToken[1]
	}
	return ""
}

func responseUnauthorized(c *gin.Context, message string) {
	logrus.Warnf("🔒 Unauthorized: %s", message)
	c.JSON(http.StatusUnauthorized, response.Response{
		Status:  constants.Error,
		Message: message,
	})
	c.Abort()
}

func validateAPIKey(c *gin.Context) error {
	apiKey := c.GetHeader(constants.XApiKey)
	requestAt := c.GetHeader(constants.XRequestAt)
	serviceName := c.GetHeader(constants.XServiceName)
	signatureKey := config.Config.SignatureKey

	validateKey := fmt.Sprintf("%s:%s:%s", serviceName, signatureKey, requestAt)
	hash := sha256.New()
	hash.Write([]byte(validateKey))
	resultHash := hex.EncodeToString(hash.Sum(nil))

	if apiKey != resultHash {
		logrus.Warn("❌ Invalid API Key")
		return errConstant.ErrUnauthorized
	}
	return nil
}

func contains(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

func CheckRole(roles []string, client clients.IClientRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenVal := c.Request.Context().Value(constants.Token)
		token, ok := tokenVal.(string)

		logrus.Infof("🔐 [CheckRole] Token from context: %v", tokenVal)
		if !ok || token == "" {
			logrus.Warn("❌ [CheckRole] Token not found or not string")
			responseUnauthorized(c, errConstant.ErrUnauthorized.Error())
			return
		}

		user, err := client.GetUser().GetUserbyToken(c.Request.Context())
		if err != nil {
			logrus.Warnf("❌ [CheckRole] GetUserbyToken failed: %v", err)
			responseUnauthorized(c, errConstant.ErrUnauthorized.Error())
			return
		}

		logrus.Infof("✅ [CheckRole] Authenticated user: %s with role: %s", user.Username, user.Role)
		logrus.Infof("🔍 [CheckRole] Allowed roles: %v", roles)

		if !contains(roles, user.Role) {
			logrus.Warnf("🚫 [CheckRole] Role '%s' is not allowed", user.Role)
			responseUnauthorized(c, errConstant.ErrUnauthorized.Error())
			return
		}

		userLogin := c.Request.WithContext(context.WithValue(c.Request.Context(), constants.User, user))
		c.Request = userLogin

		logrus.Infof("✅ [CheckRole] Access granted for role: %s", user.Role)
		c.Next()
	}
}

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(constants.Authorization)
		logrus.Infof("🔐 AUTH_HEADER: %s", authHeader)

		token := extractBearerToken(authHeader)
		logrus.Infof("🔐 TOKEN_EXTRACTED: %s", token)

		if token == "" {
			responseUnauthorized(c, "unauthorized: token missing")
			return
		}

		ctx := context.WithValue(c.Request.Context(), constants.Token, token)
		c.Request = c.Request.WithContext(ctx)

		logrus.Infof("🔐 Token injected to context")
		c.Next()
	}
}

// func AuthenticateWithoutToken() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		err := validateAPIKey(c)
// 		if err != nil {
// 			responseUnauthorized(c, err.Error())
// 			return
// 		}
// 		logrus.Info("🔓 API Key validated successfully")
// 		c.Next()
// 	}
// }
