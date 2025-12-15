# DES 加密函数完整说明

## 📋 函数列表总览

| 函数名 | 输入/输出 | IV模式 | 编码格式 | 使用场景 |
|--------|----------|--------|----------|----------|
| `DESEncrypt` | `[]byte → []byte` | 随机IV | 原始字节 | 底层API，高级用户 |
| `DESDecrypt` | `[]byte → []byte` | 随机IV | 原始字节 | 底层API，高级用户 |
| `DESEncryptString` | `string → string` | 随机IV | Base64 | **推荐：安全的字符串加密** |
| `DESDecryptString` | `string → string` | 随机IV | Base64 | **推荐：安全的字符串解密** |
| `DESEncryptWithFixedIV` | `[]byte → []byte` | 固定IV | 原始字节 | 兼容模式底层API |
| `DESDecryptWithFixedIV` | `[]byte → []byte` | 固定IV | 原始字节 | 兼容模式底层API |
| `DESEncryptHex` | `string → string` | 固定IV | Hex | **兼容：旧系统对接** |
| `DESDecryptHex` | `string → string` | 固定IV | Hex | **兼容：旧系统对接** |

---

## 🎯 详细说明

### 1️⃣ **DESEncrypt / DESDecrypt**（底层字节API - 随机IV）

```go
func DESEncrypt(data, key []byte) ([]byte, error)
func DESDecrypt(encryptedData, key []byte) ([]byte, error)
```

**特点：**
- ✅ 输入/输出都是 `[]byte`，灵活性最高
- ✅ 使用**随机 IV**（每次加密生成新的 IV）
- ✅ IV 附加在密文前面（前 8 字节是 IV，后面是密文）
- ✅ 安全性高：相同明文每次产生不同密文
- ⚠️ 需要手动处理字节数组

**使用场景：**
- 需要加密二进制数据（文件、图片等）
- 需要自定义编码格式
- 高级用户需要完全控制加密流程

**示例：**
```go
key := []byte("12345678") // 8字节
data := []byte("Hello")
encrypted, _ := cryptoutil.DESEncrypt(data, key)
// encrypted = [IV(8字节) + 密文]

decrypted, _ := cryptoutil.DESDecrypt(encrypted, key)
// decrypted = []byte("Hello")
```

---

### 2️⃣ **DESEncryptString / DESDecryptString**（字符串API - 随机IV）⭐ **推荐新项目**

```go
func DESEncryptString(data, key string) (string, error)
func DESDecryptString(encryptedData, key string) (string, error)
```

**特点：**
- ✅ 输入/输出都是 `string`，最方便使用
- ✅ 使用**随机 IV**（每次加密生成新的 IV）
- ✅ 自动进行 **Base64 编码**（密文可直接存储或传输）
- ✅ 安全性高：相同明文每次产生不同密文
- ✅ **推荐用于所有新项目**

**使用场景：**
- ✅ 加密用户密码、敏感信息
- ✅ API 数据传输加密
- ✅ 数据库敏感字段加密
- ✅ 任何需要安全加密的字符串场景

**示例：**
```go
encrypted1, _ := cryptoutil.DESEncryptString("Hello", "12345678")
// 结果：wdcJHK1qTl6KJcMweZFh5aeO72k8A288

encrypted2, _ := cryptoutil.DESEncryptString("Hello", "12345678")
// 结果：C23mYqhlX8XjieBnfIxmvtea364fx8jH（每次不同！）

decrypted, _ := cryptoutil.DESDecryptString(encrypted1, "12345678")
// 结果："Hello"
```

---

### 3️⃣ **DESEncryptWithFixedIV / DESDecryptWithFixedIV**（底层字节API - 固定IV）

```go
func DESEncryptWithFixedIV(data, key []byte) ([]byte, error)
func DESDecryptWithFixedIV(encryptedData, key []byte) ([]byte, error)
```

**特点：**
- ⚠️ 输入/输出都是 `[]byte`
- ⚠️ 使用**固定 IV**（直接用密钥作为 IV）
- ⚠️ 相同明文产生相同密文
- ⚠️ 安全性较低（容易被模式分析攻击）
- ✅ 兼容某些旧系统的加密方式

**使用场景：**
- 对接旧系统（对方使用固定 IV）
- 需要确定性加密（相同输入必须产生相同输出）
- 不推荐新项目使用

**示例：**
```go
key := []byte("12345678")
data := []byte("Hello")

encrypted1, _ := cryptoutil.DESEncryptWithFixedIV(data, key)
encrypted2, _ := cryptoutil.DESEncryptWithFixedIV(data, key)
// encrypted1 == encrypted2（每次相同！）

decrypted, _ := cryptoutil.DESDecryptWithFixedIV(encrypted1, key)
// decrypted = []byte("Hello")
```

---

### 4️⃣ **DESEncryptHex / DESDecryptHex**（字符串API - 固定IV）⭐ **兼容旧代码**

```go
func DESEncryptHex(data, key string) (string, error)
func DESDecryptHex(encryptedHex, key string) (string, error)
```

**特点：**
- ✅ 输入/输出都是 `string`，方便使用
- ⚠️ 使用**固定 IV**（直接用密钥作为 IV）
- ✅ 自动进行 **Hex 编码**（16进制字符串）
- ⚠️ 相同明文产生相同密文
- ✅ **完全兼容你提供的旧代码**
- ✅ 支持空字符串处理

**使用场景：**
- ✅ 解密旧系统的加密数据
- ✅ 对接使用固定IV的第三方系统
- ✅ 迁移期：同时支持新旧加密方式
- ⚠️ 不推荐新功能使用

**示例：**
```go
encrypted1, _ := cryptoutil.DESEncryptHex("Hello", "12345678")
// 结果：738939092eec608e2e2be40ceb3a6ebe

encrypted2, _ := cryptoutil.DESEncryptHex("Hello", "12345678")
// 结果：738939092eec608e2e2be40ceb3a6ebe（每次相同！）

decrypted, _ := cryptoutil.DESDecryptHex(encrypted1, "12345678")
// 结果："Hello"

// 空字符串处理
empty, _ := cryptoutil.DESEncryptHex("", "12345678")
// 结果：""（直接返回空字符串）
```

---

## 🔄 函数关系图

```
┌─────────────────────────────────────────────────────┐
│                  DES 加密函数族                      │
└─────────────────────────────────────────────────────┘
                        │
        ┌───────────────┴───────────────┐
        │                               │
   【随机IV模式】                  【固定IV模式】
   （安全，推荐）                  （兼容旧系统）
        │                               │
   ┌────┴────┐                     ┌────┴────┐
   │         │                     │         │
【底层API】【字符串API】        【底层API】【字符串API】
   │         │                     │         │
   │         │                     │         │
DESEncrypt  DESEncryptString  DESEncryptWithFixedIV  DESEncryptHex
   ↓         ↓                     ↓         ↓
[]byte    Base64                 []byte    Hex
返回      字符串                  返回      字符串
```

---

## 📊 对比测试

```go
package main

import (
	"fmt"
	"gitee.com/wangsoft/go-library/util/cryptoutil"
)

func main() {
	key := "12345678"
	data := "Hello"
	
	// 1. 随机IV + Base64（推荐）
	enc1, _ := cryptoutil.DESEncryptString(data, key)
	enc2, _ := cryptoutil.DESEncryptString(data, key)
	fmt.Printf("随机IV + Base64:\n")
	fmt.Printf("  第1次: %s\n", enc1)
	fmt.Printf("  第2次: %s\n", enc2)
	fmt.Printf("  相同? %v\n\n", enc1 == enc2) // false
	
	// 2. 固定IV + Hex（兼容）
	enc3, _ := cryptoutil.DESEncryptHex(data, key)
	enc4, _ := cryptoutil.DESEncryptHex(data, key)
	fmt.Printf("固定IV + Hex:\n")
	fmt.Printf("  第1次: %s\n", enc3)
	fmt.Printf("  第2次: %s\n", enc4)
	fmt.Printf("  相同? %v\n", enc3 == enc4) // true
}
```

输出：
```
随机IV + Base64:
  第1次: wdcJHK1qTl6KJcMweZFh5aeO72k8A288
  第2次: C23mYqhlX8XjieBnfIxmvtea364fx8jH
  相同? false

固定IV + Hex:
  第1次: 738939092eec608e2e2be40ceb3a6ebe
  第2次: 738939092eec608e2e2be40ceb3a6ebe
  相同? true
```

---

## ✅ 使用建议

### 🎯 新项目开发
```go
// ✅ 推荐：使用随机IV版本
encrypted := cryptoutil.DESEncryptString(data, key)
decrypted := cryptoutil.DESDecryptString(encrypted, key)
```

### 🔄 对接旧系统/解密旧数据
```go
// ✅ 使用固定IV的Hex版本
encrypted := cryptoutil.DESEncryptHex(data, key)
decrypted := cryptoutil.DESDecryptHex(encryptedHex, key)
```

### 🔧 高级场景（处理二进制数据）
```go
// 随机IV（安全）
encrypted := cryptoutil.DESEncrypt(fileBytes, key)

// 固定IV（兼容）
encrypted := cryptoutil.DESEncryptWithFixedIV(fileBytes, key)
```

---

## ⚠️ 安全提醒

1. **DES 已过时**：密钥只有 56 位（8字节），容易被暴力破解
2. **推荐使用 AES**：新项目应该用 `AESEncryptString`（支持 128/192/256位密钥）
3. **固定IV不安全**：仅用于兼容旧系统，新系统必须用随机IV
4. **密钥管理**：不要硬编码密钥，使用环境变量或密钥管理系统

---

## 🚀 升级路径

如果你有旧系统使用 `DESEncryptHex`，可以这样逐步升级：

### 阶段1：兼容期（同时支持新旧）
```go
// 写入：使用新的安全方式
newEncrypted := cryptoutil.AESEncryptString(data, newKey)

// 读取：尝试新方式，失败则用旧方式
decrypted, err := cryptoutil.AESDecryptString(encrypted, newKey)
if err != nil {
	// 降级到旧方式
	decrypted, err = cryptoutil.DESDecryptHex(encrypted, oldKey)
}
```

### 阶段2：数据迁移
```go
// 读取旧数据
oldDecrypted, _ := cryptoutil.DESDecryptHex(oldEncrypted, oldKey)

// 用新方式重新加密
newEncrypted, _ := cryptoutil.AESEncryptString(oldDecrypted, newKey)

// 更新数据库
updateDatabase(id, newEncrypted)
```

### 阶段3：完全切换
```go
// 只使用新的安全方式
encrypted := cryptoutil.AESEncryptString(data, key)
decrypted := cryptoutil.AESDecryptString(encrypted, key)
```
