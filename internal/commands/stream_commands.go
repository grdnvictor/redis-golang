package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"redis-go/internal/protocol"
	"redis-go/internal/storage"
)

// handleStreamAddCommand handles XADD command
func (commandRegistry *RedisCommandRegistry) handleStreamAddCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) < 3 {
		return protocolEncoder.WriteErrorResponse("ERREUR : XADD nécessite au moins 3 arguments: XADD stream_name id field value [field value ...]")
	}
	
	streamName := commandArguments[0]
	streamID := commandArguments[1]
	
	// Validate field-value pairs
	if (len(commandArguments)-2)%2 != 0 {
		return protocolEncoder.WriteErrorResponse(fmt.Sprintf("ERREUR : nombre impair d'arguments après l'ID - les champs doivent avoir des valeurs. Reçu %d arguments après l'ID", len(commandArguments)-2))
	}
	
	// Get or create stream
	var stream *storage.RedisStream
	storageValue := redisStorage.GetKeyValue(streamName)
	
	if storageValue == nil {
		// Create new stream
		stream = storage.NewRedisStream()
		redisStorage.SetKeyValue(streamName, stream, storage.RedisStreamType, nil)
	} else {
		// Validate existing key is a stream
		if storageValue.DataType != storage.RedisStreamType {
			return protocolEncoder.WriteErrorResponse("ERREUR : la clé existe mais n'est pas un stream")
		}
		stream = storageValue.StoredData.(*storage.RedisStream)
	}
	
	// Generate or validate ID
	generatedID, err := stream.GenerateStreamID(streamID)
	if err != nil {
		return protocolEncoder.WriteErrorResponse(fmt.Sprintf("ERREUR : %v", err))
	}
	
	// Parse field-value pairs
	fields := make(map[string]string)
	for i := 2; i < len(commandArguments); i += 2 {
		fields[commandArguments[i]] = commandArguments[i+1]
	}
	
	// Add entry to stream
	if err := stream.AddEntry(generatedID, fields); err != nil {
		return protocolEncoder.WriteErrorResponse(fmt.Sprintf("ERREUR : %v", err))
	}
	
	return protocolEncoder.WriteBulkStringResponse(generatedID)
}

// handleStreamRangeCommand handles XRANGE command
func (commandRegistry *RedisCommandRegistry) handleStreamRangeCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) < 3 {
		return protocolEncoder.WriteErrorResponse("ERREUR : XRANGE nécessite au moins 3 arguments: XRANGE stream_name start end [COUNT count]")
	}
	
	streamName := commandArguments[0]
	startID := commandArguments[1]
	endID := commandArguments[2]
	
	// Parse COUNT option
	count := 0
	if len(commandArguments) >= 5 && strings.ToUpper(commandArguments[3]) == "COUNT" {
		var err error
		count, err = strconv.Atoi(commandArguments[4])
		if err != nil {
			return protocolEncoder.WriteErrorResponse("ERREUR : COUNT doit être un nombre")
		}
	}
	
	// Get stream
	storageValue := redisStorage.GetKeyValue(streamName)
	if storageValue == nil {
		return protocolEncoder.WriteArrayResponse([]string{})
	}
	
	if storageValue.DataType != storage.RedisStreamType {
		return protocolEncoder.WriteErrorResponse("ERREUR : la clé n'est pas un stream")
	}
	
	stream := storageValue.StoredData.(*storage.RedisStream)
	
	// Get entries in range
	entries, err := stream.GetRange(startID, endID, count)
	if err != nil {
		return protocolEncoder.WriteErrorResponse(fmt.Sprintf("ERREUR : %v", err))
	}
	
	// Format response
	return commandRegistry.formatStreamEntries(entries, protocolEncoder)
}

// handleStreamReadCommand handles XREAD command
func (commandRegistry *RedisCommandRegistry) handleStreamReadCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) < 3 {
		return protocolEncoder.WriteErrorResponse("ERREUR : XREAD nécessite: XREAD [COUNT count] [BLOCK milliseconds] STREAMS stream_name [stream_name ...] id [id ...]")
	}
	
	count := 0
	blockMs := 0
	streamsIndex := -1
	
	// Parse options
	for i := 0; i < len(commandArguments); i++ {
		switch strings.ToUpper(commandArguments[i]) {
		case "COUNT":
			if i+1 >= len(commandArguments) {
				return protocolEncoder.WriteErrorResponse("ERREUR : COUNT nécessite un nombre")
			}
			var err error
			count, err = strconv.Atoi(commandArguments[i+1])
			if err != nil {
				return protocolEncoder.WriteErrorResponse("ERREUR : COUNT doit être un nombre")
			}
			i++
		case "BLOCK":
			if i+1 >= len(commandArguments) {
				return protocolEncoder.WriteErrorResponse("ERREUR : BLOCK nécessite un nombre")
			}
			var err error
			blockMs, err = strconv.Atoi(commandArguments[i+1])
			if err != nil {
				return protocolEncoder.WriteErrorResponse("ERREUR : BLOCK doit être un nombre")
			}
			i++
		case "STREAMS":
			streamsIndex = i + 1
			break
		}
	}
	
	if streamsIndex == -1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : STREAMS requis")
	}
	
	// Parse stream names and IDs
	remaining := commandArguments[streamsIndex:]
	if len(remaining)%2 != 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre égal de stream names et IDs requis")
	}
	
	streamCount := len(remaining) / 2
	streamNames := remaining[:streamCount]
	streamIDs := remaining[streamCount:]
	
	// Implementation for non-blocking read
	if blockMs == 0 {
		return commandRegistry.executeStreamRead(streamNames, streamIDs, count, redisStorage, protocolEncoder)
	}
	
	// Blocking read implementation
	return commandRegistry.executeBlockingStreamRead(streamNames, streamIDs, count, blockMs, redisStorage, protocolEncoder)
}

// executeStreamRead executes non-blocking stream read
func (commandRegistry *RedisCommandRegistry) executeStreamRead(streamNames, streamIDs []string, count int, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	var results []string
	
	for i, streamName := range streamNames {
		fromID := streamIDs[i]
		
		// Handle special ID "$" (latest)
		if fromID == "$" {
			storageValue := redisStorage.GetKeyValue(streamName)
			if storageValue != nil && storageValue.DataType == storage.RedisStreamType {
				stream := storageValue.StoredData.(*storage.RedisStream)
				fromID = stream.LastID
			} else {
				fromID = "0-0"
			}
		}
		
		// Get stream
		storageValue := redisStorage.GetKeyValue(streamName)
		if storageValue == nil || storageValue.DataType != storage.RedisStreamType {
			continue
		}
		
		stream := storageValue.StoredData.(*storage.RedisStream)
		entries, err := stream.GetEntriesFrom(fromID, count)
		if err != nil {
			continue
		}
		
		if len(entries) > 0 {
			// Add stream name
			results = append(results, streamName)
			
			// Add entries for this stream
			for _, entry := range entries {
				results = append(results, entry.ID)
				for field, value := range entry.Fields {
					results = append(results, field, value)
				}
			}
		}
	}
	
	if len(results) == 0 {
		return protocolEncoder.WriteNullBulkStringResponse()
	}
	
	return protocolEncoder.WriteArrayResponse(results)
}

// executeBlockingStreamRead executes blocking stream read
func (commandRegistry *RedisCommandRegistry) executeBlockingStreamRead(streamNames, streamIDs []string, count, blockMs int, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	// For simplicity, we'll implement a basic blocking mechanism
	// In a full implementation, this would use channels and goroutines
	
	startTime := time.Now()
	timeout := time.Duration(blockMs) * time.Millisecond
	
	for {
		// Try to read
		err := commandRegistry.executeStreamRead(streamNames, streamIDs, count, redisStorage, protocolEncoder)
		if err != nil {
			return err
		}
		
		// Check timeout
		if blockMs > 0 && time.Since(startTime) > timeout {
			return protocolEncoder.WriteNullBulkStringResponse()
		}
		
		// Sleep briefly before retry
		time.Sleep(100 * time.Millisecond)
	}
}

// handleStreamLengthCommand handles XLEN command
func (commandRegistry *RedisCommandRegistry) handleStreamLengthCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : XLEN nécessite exactement 1 argument: XLEN stream_name")
	}
	
	streamName := commandArguments[0]
	
	// Get stream
	storageValue := redisStorage.GetKeyValue(streamName)
	if storageValue == nil {
		return protocolEncoder.WriteIntegerResponse(0)
	}
	
	if storageValue.DataType != storage.RedisStreamType {
		return protocolEncoder.WriteErrorResponse("ERREUR : la clé n'est pas un stream")
	}
	
	stream := storageValue.StoredData.(*storage.RedisStream)
	return protocolEncoder.WriteIntegerResponse(stream.GetLength())
}

// handleStreamDeleteCommand handles XDEL command
func (commandRegistry *RedisCommandRegistry) handleStreamDeleteCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) < 2 {
		return protocolEncoder.WriteErrorResponse("ERREUR : XDEL nécessite au moins 2 arguments: XDEL stream_name id [id ...]")
	}
	
	streamName := commandArguments[0]
	messageIDs := commandArguments[1:]
	
	// Get stream
	storageValue := redisStorage.GetKeyValue(streamName)
	if storageValue == nil {
		return protocolEncoder.WriteIntegerResponse(0)
	}
	
	if storageValue.DataType != storage.RedisStreamType {
		return protocolEncoder.WriteErrorResponse("ERREUR : la clé n'est pas un stream")
	}
	
	stream := storageValue.StoredData.(*storage.RedisStream)
	
	// Delete entries
	deletedCount := 0
	for _, id := range messageIDs {
		if stream.DeleteEntry(id) {
			deletedCount++
		}
	}
	
	return protocolEncoder.WriteIntegerResponse(int64(deletedCount))
}

// handleStreamGroupCommand handles XGROUP command
func (commandRegistry *RedisCommandRegistry) handleStreamGroupCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) < 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : XGROUP nécessite une sous-commande")
	}
	
	subCommand := strings.ToUpper(commandArguments[0])
	
	switch subCommand {
	case "CREATE":
		return commandRegistry.handleStreamGroupCreate(commandArguments[1:], redisStorage, protocolEncoder)
	case "DESTROY":
		return commandRegistry.handleStreamGroupDestroy(commandArguments[1:], redisStorage, protocolEncoder)
	default:
		return protocolEncoder.WriteErrorResponse(fmt.Sprintf("ERREUR : sous-commande XGROUP inconnue: %s", subCommand))
	}
}

// handleStreamGroupCreate handles XGROUP CREATE command
func (commandRegistry *RedisCommandRegistry) handleStreamGroupCreate(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) < 3 {
		return protocolEncoder.WriteErrorResponse("ERREUR : XGROUP CREATE nécessite: XGROUP CREATE stream_name group_name id")
	}
	
	streamName := commandArguments[0]
	groupName := commandArguments[1]
	startID := commandArguments[2]
	
	// Get stream
	storageValue := redisStorage.GetKeyValue(streamName)
	if storageValue == nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : le stream n'existe pas")
	}
	
	if storageValue.DataType != storage.RedisStreamType {
		return protocolEncoder.WriteErrorResponse("ERREUR : la clé n'est pas un stream")
	}
	
	stream := storageValue.StoredData.(*storage.RedisStream)
	
	// Create consumer group
	if err := stream.CreateConsumerGroup(groupName, startID); err != nil {
		return protocolEncoder.WriteErrorResponse(fmt.Sprintf("ERREUR : %v", err))
	}
	
	return protocolEncoder.WriteSimpleStringResponse("OK")
}

// handleStreamGroupDestroy handles XGROUP DESTROY command
func (commandRegistry *RedisCommandRegistry) handleStreamGroupDestroy(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) < 2 {
		return protocolEncoder.WriteErrorResponse("ERREUR : XGROUP DESTROY nécessite: XGROUP DESTROY stream_name group_name")
	}
	
	streamName := commandArguments[0]
	groupName := commandArguments[1]
	
	// Get stream
	storageValue := redisStorage.GetKeyValue(streamName)
	if storageValue == nil {
		return protocolEncoder.WriteIntegerResponse(0)
	}
	
	if storageValue.DataType != storage.RedisStreamType {
		return protocolEncoder.WriteErrorResponse("ERREUR : la clé n'est pas un stream")
	}
	
	stream := storageValue.StoredData.(*storage.RedisStream)
	
	// Delete consumer group
	if _, exists := stream.ConsumerGroups[groupName]; exists {
		delete(stream.ConsumerGroups, groupName)
		return protocolEncoder.WriteIntegerResponse(1)
	}
	
	return protocolEncoder.WriteIntegerResponse(0)
}

// handleStreamReadGroupCommand handles XREADGROUP command
func (commandRegistry *RedisCommandRegistry) handleStreamReadGroupCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) < 5 {
		return protocolEncoder.WriteErrorResponse("ERREUR : XREADGROUP nécessite: XREADGROUP GROUP group_name consumer_name [COUNT count] [BLOCK milliseconds] STREAMS stream_name [stream_name ...] id [id ...]")
	}
	
	// Parse GROUP
	if strings.ToUpper(commandArguments[0]) != "GROUP" {
		return protocolEncoder.WriteErrorResponse("ERREUR : XREADGROUP doit commencer par GROUP")
	}
	
	groupName := commandArguments[1]
	consumerName := commandArguments[2]
	
	count := 0
	blockMs := 0
	streamsIndex := -1
	
	// Parse options
	for i := 3; i < len(commandArguments); i++ {
		switch strings.ToUpper(commandArguments[i]) {
		case "COUNT":
			if i+1 >= len(commandArguments) {
				return protocolEncoder.WriteErrorResponse("ERREUR : COUNT nécessite un nombre")
			}
			var err error
			count, err = strconv.Atoi(commandArguments[i+1])
			if err != nil {
				return protocolEncoder.WriteErrorResponse("ERREUR : COUNT doit être un nombre")
			}
			i++
		case "BLOCK":
			if i+1 >= len(commandArguments) {
				return protocolEncoder.WriteErrorResponse("ERREUR : BLOCK nécessite un nombre")
			}
			var err error
			blockMs, err = strconv.Atoi(commandArguments[i+1])
			if err != nil {
				return protocolEncoder.WriteErrorResponse("ERREUR : BLOCK doit être un nombre")
			}
			i++
		case "STREAMS":
			streamsIndex = i + 1
			break
		}
	}
	
	if streamsIndex == -1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : STREAMS requis")
	}
	
	// Parse stream names and IDs
	remaining := commandArguments[streamsIndex:]
	if len(remaining)%2 != 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre égal de stream names et IDs requis")
	}
	
	streamCount := len(remaining) / 2
	streamNames := remaining[:streamCount]
	streamIDs := remaining[streamCount:]
	
	// Execute read group
	return commandRegistry.executeStreamReadGroup(groupName, consumerName, streamNames, streamIDs, count, blockMs, redisStorage, protocolEncoder)
}

// executeStreamReadGroup executes consumer group read
func (commandRegistry *RedisCommandRegistry) executeStreamReadGroup(groupName, consumerName string, streamNames, streamIDs []string, count, blockMs int, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	var results []string
	
	for i, streamName := range streamNames {
		fromID := streamIDs[i]
		
		// Get stream
		storageValue := redisStorage.GetKeyValue(streamName)
		if storageValue == nil || storageValue.DataType != storage.RedisStreamType {
			continue
		}
		
		stream := storageValue.StoredData.(*storage.RedisStream)
		
		// Check if ">" is used (read new messages)
		if fromID == ">" {
			entries, err := stream.ReadFromGroup(groupName, consumerName, count, false)
			if err != nil {
				continue
			}
			
			if len(entries) > 0 {
				// Add stream name
				results = append(results, streamName)
				
				// Add entries for this stream
				for _, entry := range entries {
					results = append(results, entry.ID)
					for field, value := range entry.Fields {
						results = append(results, field, value)
					}
				}
			}
		}
	}
	
	if len(results) == 0 {
		return protocolEncoder.WriteNullBulkStringResponse()
	}
	
	return protocolEncoder.WriteArrayResponse(results)
}

// handleStreamAckCommand handles XACK command
func (commandRegistry *RedisCommandRegistry) handleStreamAckCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) < 3 {
		return protocolEncoder.WriteErrorResponse("ERREUR : XACK nécessite au moins 3 arguments: XACK stream_name group_name id [id ...]")
	}
	
	streamName := commandArguments[0]
	groupName := commandArguments[1]
	messageIDs := commandArguments[2:]
	
	// Get stream
	storageValue := redisStorage.GetKeyValue(streamName)
	if storageValue == nil {
		return protocolEncoder.WriteIntegerResponse(0)
	}
	
	if storageValue.DataType != storage.RedisStreamType {
		return protocolEncoder.WriteErrorResponse("ERREUR : la clé n'est pas un stream")
	}
	
	stream := storageValue.StoredData.(*storage.RedisStream)
	
	// Acknowledge messages
	ackedCount, err := stream.AckMessage(groupName, messageIDs)
	if err != nil {
		return protocolEncoder.WriteErrorResponse(fmt.Sprintf("ERREUR : %v", err))
	}
	
	return protocolEncoder.WriteIntegerResponse(int64(ackedCount))
}

// handleStreamPendingCommand handles XPENDING command
func (commandRegistry *RedisCommandRegistry) handleStreamPendingCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) < 2 {
		return protocolEncoder.WriteErrorResponse("ERREUR : XPENDING nécessite au moins 2 arguments: XPENDING stream_name group_name [start end count] [consumer_name]")
	}
	
	streamName := commandArguments[0]
	groupName := commandArguments[1]
	
	// Get stream
	storageValue := redisStorage.GetKeyValue(streamName)
	if storageValue == nil {
		return protocolEncoder.WriteArrayResponse([]string{})
	}
	
	if storageValue.DataType != storage.RedisStreamType {
		return protocolEncoder.WriteErrorResponse("ERREUR : la clé n'est pas un stream")
	}
	
	stream := storageValue.StoredData.(*storage.RedisStream)
	
	// Get pending messages
	consumerName := ""
	if len(commandArguments) > 2 {
		consumerName = commandArguments[2]
	}
	
	pendingIDs, err := stream.GetPendingMessages(groupName, consumerName)
	if err != nil {
		return protocolEncoder.WriteErrorResponse(fmt.Sprintf("ERREUR : %v", err))
	}
	
	return protocolEncoder.WriteArrayResponse(pendingIDs)
}

// formatStreamEntries formats stream entries for response
func (commandRegistry *RedisCommandRegistry) formatStreamEntries(entries []storage.StreamEntry, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(entries) == 0 {
		return protocolEncoder.WriteArrayResponse([]string{})
	}
	
	// Format: [id1, field1, value1, field2, value2, id2, field1, value1, field2, value2]
	var responseArray []string
	for _, entry := range entries {
		// Add entry ID
		responseArray = append(responseArray, entry.ID)
		
		// Add fields and values
		for field, value := range entry.Fields {
			responseArray = append(responseArray, field, value)
		}
	}
	
	return protocolEncoder.WriteArrayResponse(responseArray)
} 