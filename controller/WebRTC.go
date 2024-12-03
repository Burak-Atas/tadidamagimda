package controller

import (
	"nerde_yenir/helpers"
	"net/http"

	"github.com/gin-gonic/gin"
)

// WebRTC bağlantısını dışarıdan alınabilir hale getiriyoruz
func Connection(web *helpers.WebRTC) gin.HandlerFunc {
	return func(c *gin.Context) {
		// SDP verisini query parametresi olarak alıyoruz
		sdp := c.Query("sdp")
		if sdp == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "SDP data is required"})
			return
		}

		// SDP verisini kanal üzerinden gönderiyoruz
		go func() {
			web.SdpChan <- sdp
		}()

		c.JSON(http.StatusOK, gin.H{"status": "SDP received"})
	}
}
