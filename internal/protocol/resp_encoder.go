package protocol

import (
	"fmt"
	"io"
)

// RedisSerializationProtocolEncoder pour l'encodage des réponses RESP
type RedisSerializationProtocolEncoder struct {
	outputWriter io.Writer
}

// NewRedisSerializationProtocolEncoder crée un nouveau encoder RESP
func NewRedisSerializationProtocolEncoder(outputWriter io.Writer) *RedisSerializationProtocolEncoder {
	return &RedisSerializationProtocolEncoder{outputWriter: outputWriter}
}

// WriteSimpleStringResponse écrit une simple string (+OK)
func (redisEncoder *RedisSerializationProtocolEncoder) WriteSimpleStringResponse(responseString string) error {
	_, writeError := fmt.Fprintf(redisEncoder.outputWriter, "+%s\r\n", responseString)
	return writeError
}

// WriteErrorResponse écrit une erreur (-ERR message)
func (redisEncoder *RedisSerializationProtocolEncoder) WriteErrorResponse(errorMessage string) error {
	_, writeError := fmt.Fprintf(redisEncoder.outputWriter, "-%s\r\n", errorMessage)
	return writeError
}

// WriteIntegerResponse écrit un entier (:123)
func (redisEncoder *RedisSerializationProtocolEncoder) WriteIntegerResponse(integerValue int64) error {
	_, writeError := fmt.Fprintf(redisEncoder.outputWriter, ":%d\r\n", integerValue)
	return writeError
}

// WriteBulkStringResponse écrit une bulk string ($5\r\nhello\r\n)
func (redisEncoder *RedisSerializationProtocolEncoder) WriteBulkStringResponse(bulkString string) error {
	_, writeError := fmt.Fprintf(redisEncoder.outputWriter, "$%d\r\n%s\r\n", len(bulkString), bulkString)
	return writeError
}

// WriteNullBulkStringResponse écrit une bulk string null ($-1\r\n)
func (redisEncoder *RedisSerializationProtocolEncoder) WriteNullBulkStringResponse() error {
	_, writeError := fmt.Fprintf(redisEncoder.outputWriter, "$-1\r\n")
	return writeError
}

// WriteArrayResponse écrit un array (*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n)
func (redisEncoder *RedisSerializationProtocolEncoder) WriteArrayResponse(arrayElements []string) error {
	if _, writeError := fmt.Fprintf(redisEncoder.outputWriter, "*%d\r\n", len(arrayElements)); writeError != nil {
		return writeError
	}

	for _, arrayElement := range arrayElements {
		if writeError := redisEncoder.WriteBulkStringResponse(arrayElement); writeError != nil {
			return writeError
		}
	}

	return nil
}
