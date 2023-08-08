package test

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"

	"github.com/andybalholm/brotli"
)

func DecompressBrotliFile(filePath string) []byte {
	content, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	return DecompressBrotli(content)
}

func DecompressBrotli(content []byte) []byte {
	var buf bytes.Buffer
	buf.ReadFrom(brotli.NewReader(bytes.NewReader(content)))
	return buf.Bytes()
}

func DecompressGzipFile(filePath string) []byte {
	content, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	return DecompressGzip(content)
}

func DecompressGzip(content []byte) []byte {
	var buf bytes.Buffer
	gzipReader, err := gzip.NewReader(bytes.NewReader(content))
	if err != nil {
		panic(err)
	}
	buf.ReadFrom(gzipReader)
	return buf.Bytes()
}

func CompressToFile(content []byte, file string) {
	compressedContent := compress(content, func(buffer *bytes.Buffer) io.WriteCloser {
		return brotli.NewWriterLevel(buffer, brotli.BestCompression)
	})
	err := os.WriteFile(file+".br", compressedContent, 0644)
	if err != nil {
		panic(err)
	}
	compressedContent = compress(content, func(buffer *bytes.Buffer) io.WriteCloser {
		writer, _ := gzip.NewWriterLevel(buffer, gzip.BestCompression)
		return writer
	})
	err = os.WriteFile(file+".gz", compressedContent, 0644)
	if err != nil {
		panic(err)
	}
}

func compress(content []byte, compression func(buffer *bytes.Buffer) io.WriteCloser) []byte {
	var buffer bytes.Buffer
	writer := compression(&buffer)
	writer.Write(content)
	writer.Close()
	return buffer.Bytes()
}
