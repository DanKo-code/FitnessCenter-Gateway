package middlewares

import (
	"Gateway/internal/coach/coach_errors"
	"Gateway/internal/coach/dtos"
	"Gateway/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
	"strings"
)

func ValidateCreateCoachMW() gin.HandlerFunc {
	return func(c *gin.Context) {

		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
			return
		}

		name := form.Value["name"]
		description := form.Value["description"]
		services := form.Value["services"]

		if (len(name) != 1 || name[0] == "") ||
			(len(description) != 1 || description[0] == "") ||
			(len(services) != 1 || services[0] == "") {
			logger.ErrorLogger.Printf(coach_errors.OnlyPhotoOptional.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": coach_errors.OnlyPhotoOptional.Error()})
			return
		}

		//name validation
		nameValue := name[0]
		if len(nameValue) < 2 || len(nameValue) > 100 {
			logger.ErrorLogger.Printf("Name must be between 2 and 100 characters long")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name must be between 3 and 100 characters long"})
			return
		}
		allowedNameRegex := `^[a-zA-Zа-яА-Я0-9]+$`
		matched, _ := regexp.MatchString(allowedNameRegex, nameValue)
		if !matched {
			logger.ErrorLogger.Printf("Name can only contain Russian and English letters, digits, and spaces")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name can only contain Russian and English letters, digits, and spaces"})
			return
		}

		//description validation
		descriptionValue := description[0]
		if len(descriptionValue) < 10 || len(nameValue) > 500 {
			logger.ErrorLogger.Printf("Description must be between 10 and 500 characters long")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Description must be between 10 and 500 characters long"})
			return
		}

		//services validation
		servicesIds := strings.Split(services[0], ",")
		if len(servicesIds) == 0 {
			logger.ErrorLogger.Printf("at least one service is required")
			c.JSON(http.StatusBadRequest, gin.H{"error": "at least one service is required"})
			return
		}

		createCoachCommand := &dtos.CreateCoachCommand{
			Name:        nameValue,
			Description: descriptionValue,
			Services:    servicesIds,
		}

		c.Set("CreateCoachCommand", createCoachCommand)

		c.Next()
	}
}
