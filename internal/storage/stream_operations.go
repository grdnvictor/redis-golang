package storage

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// StreamEntry represents a single entry in a Redis stream
type StreamEntry struct {
	ID       string            // Format: "1609459200000-0"
	Fields   map[string]string // Key-value pairs
	Added    time.Time         // When the entry was added
}

// ConsumerInfo represents a consumer within a consumer group
type ConsumerInfo struct {
	Name         string
	LastSeen     time.Time
	PendingCount int
	PendingIDs   map[string]time.Time // message_id -> delivery_time
}

// ConsumerGroup represents a consumer group for a stream
type ConsumerGroup struct {
	Name          string
	LastDelivered string                    // Last delivered message ID
	Consumers     map[string]*ConsumerInfo  // consumer_name -> ConsumerInfo
	PendingIDs    map[string]string         // message_id -> consumer_name
	CreatedAt     time.Time
	mutex         sync.RWMutex
}

// RedisStream represents a Redis stream data structure
type RedisStream struct {
	Entries        []StreamEntry              // Time-ordered entries
	ConsumerGroups map[string]*ConsumerGroup  // group_name -> ConsumerGroup
	LastID         string                     // Last generated ID
	Length         int64                      // Current length
	mutex          sync.RWMutex
}

// NewRedisStream creates a new Redis stream
func NewRedisStream() *RedisStream {
	return &RedisStream{
		Entries:        make([]StreamEntry, 0),
		ConsumerGroups: make(map[string]*ConsumerGroup),
		LastID:         "0-0",
		Length:         0,
	}
}

// GenerateStreamID generates a new stream ID
func (stream *RedisStream) GenerateStreamID(userID string) (string, error) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()

	currentTime := time.Now().UnixMilli()
	
	if userID == "*" {
		// Auto-generate ID
		lastTimestamp, lastSequence := parseStreamID(stream.LastID)
		
		if currentTime > lastTimestamp {
			stream.LastID = fmt.Sprintf("%d-0", currentTime)
		} else {
			stream.LastID = fmt.Sprintf("%d-%d", lastTimestamp, lastSequence+1)
		}
		return stream.LastID, nil
	}
	
	// User provided ID - validate it's greater than lastID
	if userID != "0-0" && !isIDGreater(userID, stream.LastID) {
		return "", fmt.Errorf("The ID specified in XADD must be greater than the last generated ID")
	}
	
	stream.LastID = userID
	return userID, nil
}

// AddEntry adds a new entry to the stream
func (stream *RedisStream) AddEntry(id string, fields map[string]string) error {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()

	entry := StreamEntry{
		ID:     id,
		Fields: fields,
		Added:  time.Now(),
	}

	stream.Entries = append(stream.Entries, entry)
	stream.Length++
	
	return nil
}

// GetRange returns entries within a range
func (stream *RedisStream) GetRange(start, end string, count int) ([]StreamEntry, error) {
	stream.mutex.RLock()
	defer stream.mutex.RUnlock()

	var results []StreamEntry
	
	for _, entry := range stream.Entries {
		if start != "-" && !isIDGreaterOrEqual(entry.ID, start) {
			continue
		}
		if end != "+" && !isIDGreaterOrEqual(end, entry.ID) {
			continue
		}
		
		results = append(results, entry)
		if count > 0 && len(results) >= count {
			break
		}
	}
	
	return results, nil
}

// GetEntriesFrom returns entries from a specific ID onwards
func (stream *RedisStream) GetEntriesFrom(fromID string, count int) ([]StreamEntry, error) {
	stream.mutex.RLock()
	defer stream.mutex.RUnlock()

	var results []StreamEntry
	
	for _, entry := range stream.Entries {
		if isIDGreater(entry.ID, fromID) {
			results = append(results, entry)
			if count > 0 && len(results) >= count {
				break
			}
		}
	}
	
	return results, nil
}

// DeleteEntry removes an entry from the stream
func (stream *RedisStream) DeleteEntry(id string) bool {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()

	for i, entry := range stream.Entries {
		if entry.ID == id {
			stream.Entries = append(stream.Entries[:i], stream.Entries[i+1:]...)
			stream.Length--
			return true
		}
	}
	return false
}

// GetLength returns the current stream length
func (stream *RedisStream) GetLength() int64 {
	stream.mutex.RLock()
	defer stream.mutex.RUnlock()
	return stream.Length
}

// CreateConsumerGroup creates a new consumer group
func (stream *RedisStream) CreateConsumerGroup(groupName, startID string) error {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()

	if _, exists := stream.ConsumerGroups[groupName]; exists {
		return fmt.Errorf("Consumer Group name already exists")
	}

	group := &ConsumerGroup{
		Name:          groupName,
		LastDelivered: startID,
		Consumers:     make(map[string]*ConsumerInfo),
		PendingIDs:    make(map[string]string),
		CreatedAt:     time.Now(),
	}

	stream.ConsumerGroups[groupName] = group
	return nil
}

// ReadFromGroup reads entries for a consumer group
func (stream *RedisStream) ReadFromGroup(groupName, consumerName string, count int, noAck bool) ([]StreamEntry, error) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()

	group, exists := stream.ConsumerGroups[groupName]
	if !exists {
		return nil, fmt.Errorf("No such consumer group")
	}

	group.mutex.Lock()
	defer group.mutex.Unlock()

	// Create consumer if not exists
	if _, exists := group.Consumers[consumerName]; !exists {
		group.Consumers[consumerName] = &ConsumerInfo{
			Name:         consumerName,
			LastSeen:     time.Now(),
			PendingCount: 0,
			PendingIDs:   make(map[string]time.Time),
		}
	}

	consumer := group.Consumers[consumerName]
	consumer.LastSeen = time.Now()

	var results []StreamEntry
	
	// Get entries after the last delivered ID
	for _, entry := range stream.Entries {
		if isIDGreater(entry.ID, group.LastDelivered) {
			results = append(results, entry)
			
			// Track pending messages if ACK is required
			if !noAck {
				group.PendingIDs[entry.ID] = consumerName
				consumer.PendingIDs[entry.ID] = time.Now()
				consumer.PendingCount++
			}
			
			group.LastDelivered = entry.ID
			
			if count > 0 && len(results) >= count {
				break
			}
		}
	}
	
	return results, nil
}

// AckMessage acknowledges a message for a consumer group
func (stream *RedisStream) AckMessage(groupName string, messageIDs []string) (int, error) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()

	group, exists := stream.ConsumerGroups[groupName]
	if !exists {
		return 0, fmt.Errorf("No such consumer group")
	}

	group.mutex.Lock()
	defer group.mutex.Unlock()

	ackedCount := 0
	
	for _, messageID := range messageIDs {
		if consumerName, exists := group.PendingIDs[messageID]; exists {
			// Remove from pending
			delete(group.PendingIDs, messageID)
			
			// Update consumer info
			if consumer, exists := group.Consumers[consumerName]; exists {
				delete(consumer.PendingIDs, messageID)
				consumer.PendingCount--
			}
			
			ackedCount++
		}
	}
	
	return ackedCount, nil
}

// GetPendingMessages returns pending messages for a consumer group
func (stream *RedisStream) GetPendingMessages(groupName string, consumerName string) ([]string, error) {
	stream.mutex.RLock()
	defer stream.mutex.RUnlock()

	group, exists := stream.ConsumerGroups[groupName]
	if !exists {
		return nil, fmt.Errorf("No such consumer group")
	}

	group.mutex.RLock()
	defer group.mutex.RUnlock()

	var pendingIDs []string
	
	if consumerName == "" {
		// Return all pending messages
		for messageID := range group.PendingIDs {
			pendingIDs = append(pendingIDs, messageID)
		}
	} else {
		// Return pending messages for specific consumer
		if consumer, exists := group.Consumers[consumerName]; exists {
			for messageID := range consumer.PendingIDs {
				pendingIDs = append(pendingIDs, messageID)
			}
		}
	}
	
	sort.Strings(pendingIDs)
	return pendingIDs, nil
}

// Helper functions for stream ID comparison
func parseStreamID(id string) (timestamp int64, sequence int64) {
	parts := strings.Split(id, "-")
	if len(parts) != 2 {
		return 0, 0
	}
	
	timestamp, _ = strconv.ParseInt(parts[0], 10, 64)
	sequence, _ = strconv.ParseInt(parts[1], 10, 64)
	return
}

func isIDGreater(id1, id2 string) bool {
	ts1, seq1 := parseStreamID(id1)
	ts2, seq2 := parseStreamID(id2)
	
	if ts1 > ts2 {
		return true
	}
	if ts1 == ts2 && seq1 > seq2 {
		return true
	}
	return false
}

func isIDGreaterOrEqual(id1, id2 string) bool {
	if id1 == id2 {
		return true
	}
	return isIDGreater(id1, id2)
}

// GetStreamInfo returns information about the stream
func (stream *RedisStream) GetStreamInfo() map[string]interface{} {
	stream.mutex.RLock()
	defer stream.mutex.RUnlock()

	info := map[string]interface{}{
		"length":           stream.Length,
		"last-generated-id": stream.LastID,
		"groups":           len(stream.ConsumerGroups),
	}

	if stream.Length > 0 {
		info["first-entry"] = stream.Entries[0]
		info["last-entry"] = stream.Entries[len(stream.Entries)-1]
	}

	return info
} 