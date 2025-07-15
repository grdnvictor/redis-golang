# 🧪 Redis Streams Testing Guide

This guide provides practical testing scenarios to validate the Redis Streams implementation in your Go Redis server.

## 🚀 Quick Setup

1. **Start the server:**
```bash
make run
```

2. **Connect with Redis CLI:**
```bash
redis-cli -p 6379
```

3. **Verify server is running:**
```bash
PING
# Expected: PONG
```

---

## 📋 Test Categories

### 1. Basic Stream Operations
### 2. Consumer Groups
### 3. Blocking Operations
### 4. Error Handling
### 5. Performance Testing
### 6. Real-world Scenarios

---

## 🔧 1. Basic Stream Operations

### Test 1.1: XADD - Adding Messages

```bash
# Test auto-generated IDs
XADD mystream * field1 "value1" field2 "value2"
# Expected: "1703188800000-0" (or similar timestamp-based ID)

XADD mystream * user "alice" action "login" timestamp "2024-01-15T10:30:00Z"
# Expected: "1703188800001-0" (next sequential ID)

# Test custom ID
XADD mystream 1000000000000-0 field1 "custom_value"
# Expected: "1000000000000-0"

# Test multiple fields
XADD mystream * f1 "v1" f2 "v2" f3 "v3" f4 "v4"
# Expected: New ID generated

# ✅ PASS: If all commands return valid IDs
```

### Test 1.2: XRANGE - Reading Message Ranges

```bash
# Read all messages
XRANGE mystream - +
# Expected: Array of all stream entries with IDs and field-value pairs

# Read with count limit
XRANGE mystream - + COUNT 2
# Expected: First 2 entries only

# Read specific range
XRANGE mystream 1000000000000-0 +
# Expected: Messages from custom ID onwards
```

### Test 1.3: XREAD - Reading New Messages

```bash
# Read from beginning
XREAD STREAMS mystream 0-0
# Expected: All messages after 0-0

# Read from latest (should be empty initially)
XREAD STREAMS mystream $
# Expected: Empty array

# Add a message in another terminal, then:
XREAD STREAMS mystream $
# Expected: New message
```

### Test 1.4: XLEN - Stream Length

```bash
XLEN mystream
# Expected: Integer (number of messages added)

XLEN nonexistent_stream
# Expected: 0
```

### Test 1.5: XDEL - Delete Messages

```bash
# First, get some IDs
XRANGE mystream - + COUNT 2

# Delete first message (use actual ID from above)
XDEL mystream 1703188800000-0
# Expected: 1 (number of deleted messages)

# Try to delete same message again
XDEL mystream 1703188800000-0
# Expected: 0 (already deleted)

# Delete multiple messages
XDEL mystream 1703188800001-0 1703188800002-0
# Expected: 2 (or actual number deleted)
```

---

## 👥 2. Consumer Groups

### Test 2.1: XGROUP CREATE - Creating Consumer Groups

```bash
# Create a new stream for testing
XADD orders * order_id "1001" customer "alice" amount "99.99"
XADD orders * order_id "1002" customer "bob" amount "149.99"

# Create consumer group from beginning
XGROUP CREATE orders payment_processors 0
# Expected: OK

# Create another consumer group from latest
XGROUP CREATE orders notifications $
# Expected: OK

# Try to create duplicate group
XGROUP CREATE orders payment_processors 0
# Expected: Error about existing group
```

### Test 2.2: XREADGROUP - Reading via Consumer Groups

```bash
# Consumer 1 reads messages
XREADGROUP GROUP payment_processors worker1 COUNT 1 STREAMS orders >
# Expected: Array with 1 message

# Consumer 2 reads messages
XREADGROUP GROUP payment_processors worker2 COUNT 1 STREAMS orders >
# Expected: Array with next message (different from worker1)

# Same consumer reads again
XREADGROUP GROUP payment_processors worker1 COUNT 1 STREAMS orders >
# Expected: Next available message
```

### Test 2.3: XACK - Acknowledging Messages

```bash
# First get a message ID from previous test
XREADGROUP GROUP payment_processors worker1 COUNT 1 STREAMS orders >
# Note the message ID from response

# Acknowledge the message (use actual ID)
XACK orders payment_processors 1703188800000-0
# Expected: 1 (number of acknowledged messages)

# Try to acknowledge non-existent message
XACK orders payment_processors 9999999999999-0
# Expected: 0
```

### Test 2.4: XPENDING - Checking Pending Messages

```bash
# Check all pending messages
XPENDING orders payment_processors
# Expected: Array of pending message IDs

# Check pending for specific consumer
XPENDING orders payment_processors worker1
# Expected: Array of messages pending for worker1
```

### Test 2.5: XGROUP DESTROY - Removing Consumer Groups

```bash
# Destroy a consumer group
XGROUP DESTROY orders notifications
# Expected: 1 (group destroyed)

# Try to destroy non-existent group
XGROUP DESTROY orders nonexistent_group
# Expected: 0
```

---

## ⏱️ 3. Blocking Operations

### Test 3.1: XREAD with BLOCK

```bash
# Terminal 1: Start blocking read
XREAD BLOCK 5000 STREAMS mystream $

# Terminal 2: Add a message
XADD mystream * urgent "true" message "immediate_action_required"

# Terminal 1 should receive: The new message immediately
```

### Test 3.2: XREAD with COUNT and BLOCK

```bash
# Terminal 1: Block for multiple messages
XREAD BLOCK 10000 COUNT 3 STREAMS mystream $

# Terminal 2: Add multiple messages
XADD mystream * msg "1"
XADD mystream * msg "2"
XADD mystream * msg "3"

# Terminal 1 should receive: All 3 messages
```

---

## ❌ 4. Error Handling

### Test 4.1: Invalid Commands

```bash
# Too few arguments
XADD mystream
# Expected: Error about insufficient arguments

# Odd number of field-value pairs
XADD mystream * field1 "value1" field2
# Expected: Error about odd number of arguments

# Invalid COUNT
XRANGE mystream - + COUNT abc
# Expected: Error about COUNT being a number

# Invalid BLOCK time
XREAD BLOCK abc STREAMS mystream $
# Expected: Error about BLOCK being a number
```

### Test 4.2: Non-existent Resources

```bash
# Read from non-existent stream
XRANGE nonexistent_stream - +
# Expected: Empty array

# Consumer group operations on non-existent stream
XGROUP CREATE nonexistent_stream mygroup 0
# Expected: Error about stream not existing

# Read from non-existent consumer group
XREADGROUP GROUP nonexistent_group consumer1 STREAMS mystream >
# Expected: Error about group not existing
```

---

## 🚄 5. Performance Testing

### Test 5.1: Bulk Message Adding

```bash
# Add multiple messages quickly
for i in {1..100}; do
  redis-cli -p 6379 XADD perftest * seq $i timestamp $(date +%s)
done

# Check length
XLEN perftest
# Expected: 100

# Read all
XRANGE perftest - +
# Expected: All 100 messages
```

### Test 5.2: Consumer Group Performance

```bash
# Create consumer group
XGROUP CREATE perftest workers 0

# Multiple consumers read in parallel
# Terminal 1:
XREADGROUP GROUP workers worker1 COUNT 10 STREAMS perftest >

# Terminal 2:
XREADGROUP GROUP workers worker2 COUNT 10 STREAMS perftest >

# Terminal 3:
XREADGROUP GROUP workers worker3 COUNT 10 STREAMS perftest >

# Each should get different messages
```

---

## 🌍 6. Real-world Scenarios

### Test 6.1: Order Processing System

```bash
# Create order events stream
XADD orders * event "order_created" order_id "ORD-001" customer_id "123" amount "99.99"
XADD orders * event "payment_received" order_id "ORD-001" payment_method "credit_card"
XADD orders * event "order_shipped" order_id "ORD-001" tracking_number "TRK-456"

# Create processing services
XGROUP CREATE orders payment_service 0
XGROUP CREATE orders shipping_service 0
XGROUP CREATE orders notification_service 0

# Payment service processes orders
XREADGROUP GROUP payment_service worker1 COUNT 5 STREAMS orders >
XACK orders payment_service 1703188800000-0

# Shipping service processes orders
XREADGROUP GROUP shipping_service worker1 COUNT 5 STREAMS orders >
XACK orders shipping_service 1703188800000-0

# Check what's pending
XPENDING orders payment_service
XPENDING orders shipping_service
```

### Test 6.2: Chat System

```bash
# Create chat room
XADD chat:room1 * user "alice" message "Hello everyone!" timestamp "2024-01-15T10:30:00Z"
XADD chat:room1 * user "bob" message "Hi Alice!" timestamp "2024-01-15T10:30:15Z"
XADD chat:room1 * user "charlie" message "How's everyone doing?" timestamp "2024-01-15T10:30:30Z"

# Read chat history
XRANGE chat:room1 - +

# Set up real-time reading
XREAD BLOCK 30000 STREAMS chat:room1 $

# In another terminal, add messages:
XADD chat:room1 * user "david" message "Just joined!" timestamp "2024-01-15T10:31:00Z"
```

### Test 6.3: Log Aggregation

```bash
# Different service logs
XADD logs:api * level "INFO" service "auth" message "User login successful" user_id "123"
XADD logs:api * level "ERROR" service "payment" message "Payment failed" error_code "insufficient_funds"
XADD logs:api * level "WARNING" service "inventory" message "Low stock alert" product_id "456"

# Analytics consumer group
XGROUP CREATE logs:api analytics 0
XREADGROUP GROUP analytics processor1 COUNT 10 STREAMS logs:api >

# Alert consumer group
XGROUP CREATE logs:api alerts 0
XREADGROUP GROUP alerts alert_processor COUNT 10 STREAMS logs:api >
```

---

## ✅ Test Results Checklist

### Basic Operations
- [ ] XADD with auto-generated IDs works
- [ ] XADD with custom IDs works
- [ ] XRANGE returns correct message ranges
- [ ] XREAD returns messages from specific positions
- [ ] XLEN returns correct stream length
- [ ] XDEL removes messages correctly

### Consumer Groups
- [ ] XGROUP CREATE creates groups successfully
- [ ] XREADGROUP distributes messages between consumers
- [ ] XACK acknowledges messages properly
- [ ] XPENDING shows pending messages
- [ ] XGROUP DESTROY removes groups

### Error Handling
- [ ] Invalid arguments return appropriate errors
- [ ] Non-existent resources handled gracefully
- [ ] Type mismatches detected correctly

### Performance
- [ ] Handles multiple rapid inserts
- [ ] Consumer groups distribute load effectively
- [ ] Blocking operations work correctly

---

## 🔍 Troubleshooting

### Common Issues

**1. "Stream doesn't exist" errors:**
```bash
# Solution: Create stream first
XADD mystream * field "value"
```

**2. Consumer group already exists:**
```bash
# Solution: Destroy and recreate
XGROUP DESTROY mystream mygroup
XGROUP CREATE mystream mygroup 0
```

**3. No messages in XREAD:**
```bash
# Check stream has messages
XLEN mystream
# Verify correct ID
XRANGE mystream - +
```

**4. Blocking operations not working:**
```bash
# Verify syntax
XREAD BLOCK 5000 STREAMS mystream $
# Use $ for latest messages
```

### Debug Commands

```bash
# Check stream length
XLEN mystream

# See all messages
XRANGE mystream - +

# Check consumer groups
XPENDING mystream mygroup

# Verify server is responsive
PING
```

---

## 📊 Expected Performance Metrics

### Typical Performance (Development Environment)
- **Message insertion**: ~1000-5000 messages/second
- **Range queries**: ~500-2000 queries/second
- **Consumer group reads**: ~300-1000 reads/second
- **Memory usage**: ~100-500 bytes per message

### Scaling Considerations
- Stream length affects range query performance
- Consumer groups add overhead but enable parallel processing
- Blocking operations use minimal resources when idle

---

## 🎯 Test Automation Script

Save this as `test_streaming.sh`:

```bash
#!/bin/bash

echo "🧪 Starting Redis Streams Testing..."

# Test basic operations
echo "Testing basic operations..."
redis-cli -p 6379 XADD test_stream \* field1 "value1" field2 "value2"
redis-cli -p 6379 XLEN test_stream
redis-cli -p 6379 XRANGE test_stream - +

# Test consumer groups
echo "Testing consumer groups..."
redis-cli -p 6379 XGROUP CREATE test_stream test_group 0
redis-cli -p 6379 XREADGROUP GROUP test_group consumer1 COUNT 1 STREAMS test_stream \>

echo "✅ Basic tests completed!"
```

Run with:
```bash
chmod +x test_streaming.sh
./test_streaming.sh
```

---

## 📚 Additional Resources

- **Full Documentation**: `REDIS_STREAMS_README.md`
- **Command Reference**: See `internal/commands/stream_commands.go`
- **Storage Implementation**: See `internal/storage/stream_operations.go`

---

*Happy Testing! 🚀* 