package compressor

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/knstch/gophermart/internal/app/logger"
)

// A writer struct containing http.ResponseWriter interface
// and gzip.Writer.
type gzipWriter struct {
	res http.ResponseWriter
	zw  *gzip.Writer
}

// A builder function accepting original http.ResponseWriter and
// returning gzipWriter struct.
func newGzipWriter(res http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		res: res,
		zw:  gzip.NewWriter(res),
	}
}

// A gzipWriter struct method implementing original Header method in http.ResponseWriter.
func (gw *gzipWriter) Header() http.Header {
	return gw.res.Header()
}

// A gzipWriter struct method compressing data using gzip.Writer interface.
func (gw *gzipWriter) Write(b []byte) (int, error) {
	return gw.zw.Write(b)
}

// A gzipWriter struct method setting Content-Encoding to gzip using http.ResponseWriter interface.
func (gw *gzipWriter) WriteHeader(statusCode int) {
	gw.res.Header().Set("Content-Encoding", "gzip")
	gw.res.WriteHeader(statusCode)
}

// A gzipWriter struct method closing response using gzip.Writer interface.
func (gw *gzipWriter) Close() error {
	return gw.zw.Close()
}

// A struct implementing io.ReadCloser and gzip.Reader interface
type gzipReader struct {
	req io.ReadCloser
	zr  *gzip.Reader
}

// A builder function returning gzipReader struct and error. Inside of the function we
// read request using gzip interface.
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

// A gzipReader struct method implementing gzip.Reader to read data from a request.
func (gr *gzipReader) Read(b []byte) (n int, err error) {
	return gr.zr.Read(b)
}

// A gzipReader struct method closing readstream.
func (gr *gzipReader) Close() error {
	if err := gr.req.Close(); err != nil {
		return err
	}
	return gr.zr.Close()
}

// A middleware compressing data using gzip if a receiver accepts this compression type
// and if a server accepts this content encoding.
func WithCompressor(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		originalRes := res
		supportsGzip := strings.Contains(req.Header.Get("Accept-Encoding"), "gzip")
		contentEncodingGzip := strings.Contains(req.Header.Get("Content-Encoding"), "gzip")
		if supportsGzip {
			compressedRes := newGzipWriter(res)
			defer compressedRes.Close()
			originalRes = compressedRes
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
