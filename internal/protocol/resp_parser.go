package protocol

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

// RedisSerializationProtocolParser pour le parsing des commandes RESP
type RedisSerializationProtocolParser struct {
	bufferedReader *bufio.Reader
}

// NewRedisSerializationProtocolParser crée un nouveau parser RESP
func NewRedisSerializationProtocolParser(inputReader io.Reader) *RedisSerializationProtocolParser {
	return &RedisSerializationProtocolParser{
		bufferedReader: bufio.NewReader(inputReader),
	}
}

// ParseIncomingCommand parse une commande RESP complète
func (redisParser *RedisSerializationProtocolParser) ParseIncomingCommand() ([]string, error) {
	// Lecture du premier caractère pour déterminer le type
	protocolTypeByte, readError := redisParser.bufferedReader.ReadByte()
	if readError != nil {
		return nil, readError
	}

	switch protocolTypeByte {
	case RedisArrayType:
		return redisParser.parseRedisArray()
	default:
		return nil, fmt.Errorf("expected array type, got %c", protocolTypeByte)
	}
}

// parseRedisArray parse un array RESP (format des commandes)
func (redisParser *RedisSerializationProtocolParser) parseRedisArray() ([]string, error) {
	// Lecture du nombre d'éléments
	arrayLengthString, readError := redisParser.readProtocolLine()
	if readError != nil {
		return nil, fmt.Errorf("failed to read array length: %v", readError)
	}

	arrayLength, parseError := strconv.Atoi(arrayLengthString)
	if parseError != nil {
		return nil, fmt.Errorf("invalid array length: %s", arrayLengthString)
	}

	if arrayLength <= 0 {
		return []string{}, nil
	}

	// Lecture de chaque élément
	arrayElements := make([]string, arrayLength)
	for elementIndex := 0; elementIndex < arrayLength; elementIndex++ {
		elementValue, parseError := redisParser.parseRedisBulkString()
		if parseError != nil {
			return nil, fmt.Errorf("failed to parse element %d: %v", elementIndex, parseError)
		}
		arrayElements[elementIndex] = elementValue
	}

	return arrayElements, nil
}

// parseRedisBulkString parse une bulk string RESP
func (redisParser *RedisSerializationProtocolParser) parseRedisBulkString() (string, error) {
	// Lecture du type (doit être $)
	protocolTypeByte, readError := redisParser.bufferedReader.ReadByte()
	if readError != nil {
		return "", fmt.Errorf("failed to read bulk string type: %v", readError)
	}

	if protocolTypeByte != RedisBulkStringType {
		return "", fmt.Errorf("expected bulk string type, got %c", protocolTypeByte)
	}

	// Lecture de la longueur
	stringLengthString, readError := redisParser.readProtocolLine()
	if readError != nil {
		return "", fmt.Errorf("failed to read bulk string length: %v", readError)
	}

	stringLength, parseError := strconv.Atoi(stringLengthString)
	if parseError != nil {
		return "", fmt.Errorf("invalid bulk string length: %s", stringLengthString)
	}

	// Cas spécial : bulk string null
	if stringLength == -1 {
		return "", nil
	}

	if stringLength < 0 {
		return "", fmt.Errorf("invalid bulk string length: %d", stringLength)
	}

	// Lecture du contenu
	stringContent := make([]byte, stringLength)
	_, readError = io.ReadFull(redisParser.bufferedReader, stringContent)
	if readError != nil {
		return "", fmt.Errorf("failed to read bulk string content: %v", readError)
	}

	// Lecture du CRLF final
	carriageReturnLineFeed := make([]byte, 2)
	_, readError = io.ReadFull(redisParser.bufferedReader, carriageReturnLineFeed)
	if readError != nil {
		return "", fmt.Errorf("failed to read CRLF after bulk string: %v", readError)
	}

	if carriageReturnLineFeed[0] != '\r' || carriageReturnLineFeed[1] != '\n' {
		return "", fmt.Errorf("expected CRLF, got %v", carriageReturnLineFeed)
	}

	return string(stringContent), nil
}

// readProtocolLine lit une ligne complète (jusqu'au CRLF)
func (redisParser *RedisSerializationProtocolParser) readProtocolLine() (string, error) {
	var lineResult []byte

	for {
		currentByte, readError := redisParser.bufferedReader.ReadByte()
		if readError != nil {
			return "", readError
		}

		if currentByte == '\r' {
			// Lire le \n suivant
			nextByte, readError := redisParser.bufferedReader.ReadByte()
			if readError != nil {
				return "", readError
			}
			if nextByte != '\n' {
				return "", fmt.Errorf("expected \\n after \\r, got %c", nextByte)
			}
			break
		}

		lineResult = append(lineResult, currentByte)
	}

	return string(lineResult), nil
}
