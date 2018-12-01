package lib

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
)

// Sender is response for gin
func Sender(c *gin.Context, code int, err, data interface{}) {
	responseData := map[string]interface{}{
		"ErrorCode":    code,
		"ErrorMessage": err,
		"Data":         data,
	}
	responseJSON, _ := json.Marshal(responseData)
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.String(200, string(responseJSON))
}

// ProxySender is response for proxy
func ProxySender(c *gin.Context, data interface{}) {
	responseJSON, _ := json.Marshal(data)
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.String(200, string(responseJSON))
}
