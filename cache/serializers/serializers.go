package serializers

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
)

// JSONSerializer implements ICacheSerializer using JSON encoding
type JSONSerializer struct{}

// NewJSONSerializer creates a new JSON serializer
func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

// Name returns the serializer name
func (js *JSONSerializer) Name() string {
	return "json"
}

// Serialize converts an object to JSON bytes
func (js *JSONSerializer) Serialize(value interface{}) ([]byte, error) {
	if value == nil {
		return nil, nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return data, nil
}

// Deserialize converts JSON bytes back to an object
func (js *JSONSerializer) Deserialize(data []byte, target interface{}) error {
	if data == nil || len(data) == 0 {
		return nil
	}

	err := json.Unmarshal(data, target)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// GobSerializer implements ICacheSerializer using Gob encoding
type GobSerializer struct{}

// NewGobSerializer creates a new Gob serializer
func NewGobSerializer() *GobSerializer {
	return &GobSerializer{}
}

// Name returns the serializer name
func (gs *GobSerializer) Name() string {
	return "gob"
}

// Serialize converts an object to Gob bytes
func (gs *GobSerializer) Serialize(value interface{}) ([]byte, error) {
	if value == nil {
		return nil, nil
	}

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(value)
	if err != nil {
		return nil, fmt.Errorf("failed to encode with Gob: %w", err)
	}

	return buffer.Bytes(), nil
}

// Deserialize converts Gob bytes back to an object
func (gs *GobSerializer) Deserialize(data []byte, target interface{}) error {
	if data == nil || len(data) == 0 {
		return nil
	}

	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)

	err := decoder.Decode(target)
	if err != nil {
		return fmt.Errorf("failed to decode with Gob: %w", err)
	}

	return nil
}

// StringSerializer implements ICacheSerializer for simple string values
type StringSerializer struct{}

// NewStringSerializer creates a new string serializer
func NewStringSerializer() *StringSerializer {
	return &StringSerializer{}
}

// Name returns the serializer name
func (ss *StringSerializer) Name() string {
	return "string"
}

// Serialize converts a string to bytes
func (ss *StringSerializer) Serialize(value interface{}) ([]byte, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		// Try to convert to string
		str := fmt.Sprintf("%v", v)
		return []byte(str), nil
	}
}

// Deserialize converts bytes back to a string
func (ss *StringSerializer) Deserialize(data []byte, target interface{}) error {
	if data == nil {
		return nil
	}

	// Check if target is a pointer to string
	if strPtr, ok := target.(*string); ok {
		*strPtr = string(data)
		return nil
	}

	// Check if target is a pointer to []byte
	if bytesPtr, ok := target.(*[]byte); ok {
		*bytesPtr = make([]byte, len(data))
		copy(*bytesPtr, data)
		return nil
	}

	// Check if target is a pointer to interface{}
	if interfacePtr, ok := target.(*interface{}); ok {
		*interfacePtr = string(data)
		return nil
	}

	return fmt.Errorf("unsupported target type for string deserialization: %T", target)
}

// BinarySerializer implements ICacheSerializer for raw binary data
type BinarySerializer struct{}

// NewBinarySerializer creates a new binary serializer
func NewBinarySerializer() *BinarySerializer {
	return &BinarySerializer{}
}

// Name returns the serializer name
func (bs *BinarySerializer) Name() string {
	return "binary"
}

// Serialize returns the data as-is if it's already bytes
func (bs *BinarySerializer) Serialize(value interface{}) ([]byte, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case []byte:
		// Return a copy to avoid issues with shared slices
		result := make([]byte, len(v))
		copy(result, v)
		return result, nil
	case string:
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("binary serializer only supports []byte and string types, got: %T", value)
	}
}

// Deserialize returns the data as-is
func (bs *BinarySerializer) Deserialize(data []byte, target interface{}) error {
	if data == nil {
		return nil
	}

	// Check if target is a pointer to []byte
	if bytesPtr, ok := target.(*[]byte); ok {
		*bytesPtr = make([]byte, len(data))
		copy(*bytesPtr, data)
		return nil
	}

	// Check if target is a pointer to string
	if strPtr, ok := target.(*string); ok {
		*strPtr = string(data)
		return nil
	}

	// Check if target is a pointer to interface{}
	if interfacePtr, ok := target.(*interface{}); ok {
		// Return a copy of the bytes
		result := make([]byte, len(data))
		copy(result, data)
		*interfacePtr = result
		return nil
	}

	return fmt.Errorf("binary deserializer only supports *[]byte, *string, and *interface{} target types, got: %T", target)
}

// CompressedJSONSerializer implements ICacheSerializer with JSON + compression
// Note: This is a placeholder for demonstration. In production, you might use gzip compression.
type CompressedJSONSerializer struct {
	jsonSerializer *JSONSerializer
}

// NewCompressedJSONSerializer creates a new compressed JSON serializer
func NewCompressedJSONSerializer() *CompressedJSONSerializer {
	return &CompressedJSONSerializer{
		jsonSerializer: NewJSONSerializer(),
	}
}

// Name returns the serializer name
func (cjs *CompressedJSONSerializer) Name() string {
	return "compressed-json"
}

// Serialize converts an object to compressed JSON bytes
func (cjs *CompressedJSONSerializer) Serialize(value interface{}) ([]byte, error) {
	// First, serialize to JSON
	jsonData, err := cjs.jsonSerializer.Serialize(value)
	if err != nil {
		return nil, err
	}

	// In a real implementation, you would compress the data here
	// For demonstration, we'll just add a simple prefix to indicate "compression"
	compressed := append([]byte("COMPRESSED:"), jsonData...)
	return compressed, nil
}

// Deserialize converts compressed JSON bytes back to an object
func (cjs *CompressedJSONSerializer) Deserialize(data []byte, target interface{}) error {
	if data == nil || len(data) == 0 {
		return nil
	}

	// Check for compression prefix
	prefix := []byte("COMPRESSED:")
	if bytes.HasPrefix(data, prefix) {
		// Remove the prefix (in real implementation, decompress here)
		jsonData := data[len(prefix):]
		return cjs.jsonSerializer.Deserialize(jsonData, target)
	}

	// Fallback to regular JSON deserialization
	return cjs.jsonSerializer.Deserialize(data, target)
}
