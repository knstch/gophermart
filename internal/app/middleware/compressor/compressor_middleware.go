// Package compressor provides a middleware to compress responses and decompress requests.
package compressor

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/knstch/gophermart/internal/app/logger"
)

type gzipResponseWriter struct {
	gin.ResponseWriter
	writer     io.Writer
	gzipWriter *gzip.Writer
}

func WithCompressor() gin.HandlerFunc {
	return func(c *gin.Context) {
		supportsGzip := strings.Contains(c.GetHeader("Accept-Encoding"), "gzip")
		contentEncodingGzip := strings.Contains(c.GetHeader("Content-Encoding"), "gzip")

		if supportsGzip {
			gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestSpeed)
			if err != nil {
				logger.ErrorLogger("Error compressing data: ", err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			defer gz.Close()

			c.Header("Content-Encoding", "gzip")
			c.Writer = &gzipResponseWriter{
				ResponseWriter: c.Writer,
				writer:         io.MultiWriter(c.Writer, gz),
				gzipWriter:     gz,
			}
		}

		if contentEncodingGzip {
			reader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				logger.ErrorLogger("Error decompressing data: ", err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			defer reader.Close()
			c.Request.Body = reader
		}

		c.Next()
	}
}
