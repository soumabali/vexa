package ws

import (
	"fmt"
	"io"
	"sync"

	"github.com/klauspost/compress/zstd"
)

// CompressionLevel for zstd.
const CompressionLevel = 3

// CompressorPool manages zstd encoder/decoder pools for concurrent use.
type CompressorPool struct {
	encoders sync.Pool
	decoders sync.Pool
	level    zstd.EncoderLevel
}

// NewCompressorPool creates a new pool with the given zstd level.
func NewCompressorPool() *CompressorPool {
	return &CompressorPool{
		level: zstd.SpeedDefault,
		encoders: sync.Pool{
			New: func() interface{} {
				enc, _ := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedDefault))
				return enc
			},
		},
		decoders: sync.Pool{
			New: func() interface{} {
				dec, _ := zstd.NewReader(nil)
				return dec
			},
		},
	}
}

// GetEncoder returns an encoder from the pool.
func (p *CompressorPool) GetEncoder() *zstd.Encoder {
	return p.encoders.Get().(*zstd.Encoder)
}

// PutEncoder returns an encoder to the pool.
func (p *CompressorPool) PutEncoder(enc *zstd.Encoder) {
	p.encoders.Put(enc)
}

// GetDecoder returns a decoder from the pool.
func (p *CompressorPool) GetDecoder() *zstd.Decoder {
	return p.decoders.Get().(*zstd.Decoder)
}

// PutDecoder returns a decoder to the pool.
func (p *CompressorPool) PutDecoder(dec *zstd.Decoder) {
	p.decoders.Put(dec)
}

// Compress compresses data using zstd.
func Compress(data []byte) ([]byte, error) {
	enc, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedDefault))
	if err != nil {
		return nil, fmt.Errorf("zstd encoder: %w", err)
	}
	defer enc.Close()
	return enc.EncodeAll(data, make([]byte, 0, len(data))), nil
}

// Decompress decompresses zstd data.
func Decompress(data []byte) ([]byte, error) {
	dec, err := zstd.NewReader(nil)
	if err != nil {
		return nil, fmt.Errorf("zstd decoder: %w", err)
	}
	defer dec.Close()
	return dec.DecodeAll(data, nil)
}

// StreamingCompressor wraps an io.Writer with zstd compression.
type StreamingCompressor struct {
	writer *zstd.Encoder
	w      io.Writer
}

// NewStreamingCompressor creates a streaming compressor.
func NewStreamingCompressor(w io.Writer) (*StreamingCompressor, error) {
	enc, err := zstd.NewWriter(w, zstd.WithEncoderLevel(zstd.SpeedDefault))
	if err != nil {
		return nil, fmt.Errorf("zstd writer: %w", err)
	}
	return &StreamingCompressor{writer: enc, w: w}, nil
}

// Write implements io.Writer.
func (sc *StreamingCompressor) Write(p []byte) (n int, err error) {
	return sc.writer.Write(p)
}

// Close closes the compressor.
func (sc *StreamingCompressor) Close() error {
	return sc.writer.Close()
}

// StreamingDecompressor wraps an io.Reader with zstd decompression.
type StreamingDecompressor struct {
	reader *zstd.Decoder
	r      io.Reader
}

// NewStreamingDecompressor creates a streaming decompressor.
func NewStreamingDecompressor(r io.Reader) (*StreamingDecompressor, error) {
	dec, err := zstd.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("zstd reader: %w", err)
	}
	return &StreamingDecompressor{reader: dec, r: r}, nil
}

// Read implements io.Reader.
func (sd *StreamingDecompressor) Read(p []byte) (n int, err error) {
	return sd.reader.Read(p)
}

// Close closes the decompressor.
func (sd *StreamingDecompressor) Close() error {
	sd.reader.Close()
	return nil
}
