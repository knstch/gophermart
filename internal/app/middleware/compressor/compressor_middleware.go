package compressor

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/knstch/gophermart/internal/app/logger"
)

type gzipWriter struct {
	res http.ResponseWriter
	zw  *gzip.Writer
}

func newGzipWriter(res http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		res: res,
		zw:  gzip.NewWriter(res),
	}
}

func (gw *gzipWriter) Header() http.Header {
	return gw.res.Header()
}

func (gw *gzipWriter) Write(b []byte) (int, error) {
	return gw.zw.Write(b)
}

func (gw *gzipWriter) WriteHeader(statusCode int) {
	gw.res.Header().Set("Content-Encoding", "gzip")
	gw.res.WriteHeader(statusCode)
}

func (gw *gzipWriter) Close() error {
	return gw.zw.Close()
}

type gzipReader struct {
	req io.ReadCloser
	zr  *gzip.Reader
}

func newCompressReader(req io.ReadCloser) (*gzipReader, error) {
	zr, err := gzip.NewReader(req)
	if err != nil {
		return nil, err
	}

	return &gzipReader{
		req: req,
		zr:  zr,
	}, nil
}

func (gr *gzipReader) Read(b []byte) (n int, err error) {
	return gr.zr.Read(b)
}

func (gr *gzipReader) Close() error {
	if err := gr.req.Close(); err != nil {
		return err
	}
	return gr.zr.Close()
}

// Сжимает данные
func GzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		originalRes := res
		supportsGzip := strings.Contains(req.Header.Get("Accept-Encoding"), "gzip")
		contentEncodingGzip := strings.Contains(req.Header.Get("Content-Encoding"), "gzip")
		if supportsGzip {
			compressedRes := newGzipWriter(res)
			originalRes = compressedRes
			defer compressedRes.Close()
		}
		if contentEncodingGzip {
			decompressedReq, err := newCompressReader(req.Body)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				logger.ErrorLogger("Error during decompression: ", err)
				return
			}
			req.Body = decompressedReq
			defer decompressedReq.Close()
		}
		h.ServeHTTP(originalRes, req)
	})
}
