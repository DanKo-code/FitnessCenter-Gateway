package middlewares

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"strings"
)

type Claims struct {
	UserId string `json:"user_id"`
	Role   string `json:"role"`
	Exp    int64  `json:"exp"`
	jwt.RegisteredClaims
}

func VerifyAccessToken(accessToken string) (*Claims, error) {
	secretKey := os.Getenv("JWT_SECRET")
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func VerifyAccessTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			return
		}

		token := tokenParts[1]
		claims, err := VerifyAccessToken(token)
		if err != nil {
			fmt.Println("Error verifying token:", err)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("Role", claims.Role)
		c.Set("UserIdFromToken", claims.UserId)

		c.Next()
	}
}
