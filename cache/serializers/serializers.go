package serializers

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
)

// JSONSerializer 使用 JSON 编码实现 ICacheSerializer 接口
type JSONSerializer struct{}

// NewJSONSerializer 创建一个新的 JSON 序列化器
func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

// Name returns the serializer name
func (js *JSONSerializer) Name() string {
	return "json"
}

// Serialize 将对象转换为 JSON 字节
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

// Deserialize 将 JSON 字节转换回对象
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

// GobSerializer 使用 Gob 编码实现 ICacheSerializer 接口
type GobSerializer struct{}

// NewGobSerializer 创建一个新的 Gob 序列化器
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

// Deserialize 将 Gob 字节转换回对象
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

// StringSerializer 为简单字符串值实现 ICacheSerializer 接口
type StringSerializer struct{}

// NewStringSerializer 创建一个新的字符串序列化器
func NewStringSerializer() *StringSerializer {
	return &StringSerializer{}
}

// Name returns the serializer name
func (ss *StringSerializer) Name() string {
	return "string"
}

// Serialize 将字符串转换为字节
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
		// 尝试转换为字符串
		str := fmt.Sprintf("%v", v)
		return []byte(str), nil
	}
}

// Deserialize 将字节转换回字符串
func (ss *StringSerializer) Deserialize(data []byte, target interface{}) error {
	if data == nil {
		return nil
	}

	// 检查目标是否为字符串指针
	if strPtr, ok := target.(*string); ok {
		*strPtr = string(data)
		return nil
	}

	// 检查目标是否为 []byte 指针
	if bytesPtr, ok := target.(*[]byte); ok {
		*bytesPtr = make([]byte, len(data))
		copy(*bytesPtr, data)
		return nil
	}

	// 检查目标是否为 interface{} 指针
	if interfacePtr, ok := target.(*interface{}); ok {
		*interfacePtr = string(data)
		return nil
	}

	return fmt.Errorf("unsupported target type for string deserialization: %T", target)
}

// BinarySerializer 为原始二进制数据实现 ICacheSerializer 接口
type BinarySerializer struct{}

// NewBinarySerializer 创建一个新的二进制序列化器
func NewBinarySerializer() *BinarySerializer {
	return &BinarySerializer{}
}

// Name returns the serializer name
func (bs *BinarySerializer) Name() string {
	return "binary"
}

// Serialize 如果数据已经是字节，则按原样返回
func (bs *BinarySerializer) Serialize(value interface{}) ([]byte, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case []byte:
		// 返回副本 to avoid issues with shared slices
		result := make([]byte, len(v))
		copy(result, v)
		return result, nil
	case string:
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("binary serializer only supports []byte and string types, got: %T", value)
	}
}

// Deserialize 按原样返回数据
func (bs *BinarySerializer) Deserialize(data []byte, target interface{}) error {
	if data == nil {
		return nil
	}

	// 检查目标是否为 []byte 指针
	if bytesPtr, ok := target.(*[]byte); ok {
		*bytesPtr = make([]byte, len(data))
		copy(*bytesPtr, data)
		return nil
	}

	// 检查目标是否为字符串指针
	if strPtr, ok := target.(*string); ok {
		*strPtr = string(data)
		return nil
	}

	// 检查目标是否为 interface{} 指针
	if interfacePtr, ok := target.(*interface{}); ok {
		// 返回副本 of the bytes
		result := make([]byte, len(data))
		copy(result, data)
		*interfacePtr = result
		return nil
	}

	return fmt.Errorf("binary deserializer only supports *[]byte, *string, and *interface{} target types, got: %T", target)
}

// CompressedJSONSerializer 实现带 JSON + 压缩的 ICacheSerializer 接口
// 注意：这是一个用于演示的占位符。在生产环境中，你可能使用 gzip 压缩。
type CompressedJSONSerializer struct {
	jsonSerializer *JSONSerializer
}

// NewCompressedJSONSerializer 创建一个新的压缩 JSON 序列化器
func NewCompressedJSONSerializer() *CompressedJSONSerializer {
	return &CompressedJSONSerializer{
		jsonSerializer: NewJSONSerializer(),
	}
}

// Name returns the serializer name
func (cjs *CompressedJSONSerializer) Name() string {
	return "compressed-json"
}

// Serialize 将对象转换为压缩的 JSON 字节
func (cjs *CompressedJSONSerializer) Serialize(value interface{}) ([]byte, error) {
	// 首先，序列化为 JSON
	jsonData, err := cjs.jsonSerializer.Serialize(value)
	if err != nil {
		return nil, err
	}

	// 在实际实现中，你应该在这里压缩数据
	// 为了演示，我们只添加一个简单的前缀来表示"压缩"
	compressed := append([]byte("COMPRESSED:"), jsonData...)
	return compressed, nil
}

// Deserialize 将压缩的 JSON 字节转换回对象
func (cjs *CompressedJSONSerializer) Deserialize(data []byte, target interface{}) error {
	if data == nil || len(data) == 0 {
		return nil
	}

	// 检查压缩前缀
	prefix := []byte("COMPRESSED:")
	if bytes.HasPrefix(data, prefix) {
		// 删除前缀（在实际实现中，在这里解压缩）
		jsonData := data[len(prefix):]
		return cjs.jsonSerializer.Deserialize(jsonData, target)
	}

	// 回退到常规 JSON 反序列化
	return cjs.jsonSerializer.Deserialize(data, target)
}
