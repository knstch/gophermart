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
	return func(ctx *gin.Context) {
		supportsGzip := strings.Contains(ctx.GetHeader("Accept-Encoding"), "gzip")
		contentEncodingGzip := strings.Contains(ctx.GetHeader("Content-Encoding"), "gzip")

		if supportsGzip {
			gz, err := gzip.NewWriterLevel(ctx.Writer, gzip.BestSpeed)
			if err != nil {
				logger.ErrorLogger("Error compressing data: ", err)
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			defer gz.Close()

			ctx.Header("Content-Encoding", "gzip")
			ctx.Writer = &gzipResponseWriter{
				ResponseWriter: ctx.Writer,
				writer:         io.MultiWriter(ctx.Writer, gz),
				gzipWriter:     gz,
			}
		}

		if contentEncodingGzip {
			reader, err := gzip.NewReader(ctx.Request.Body)
			if err != nil {
				logger.ErrorLogger("Error decompressing data: ", err)
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			defer reader.Close()
			ctx.Request.Body = reader
		}

		ctx.Next()
	}
}
