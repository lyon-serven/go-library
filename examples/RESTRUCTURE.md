# Examples目录结构改进建议

## 🚨 当前问题

examples目录下存在多个`package main`和`func main()`，导致编译冲突：

```
examples/
├── cache_test_main.go             # main函数
├── jwtutil_example.go             # main函数
├── multilevel_cache_example.go    # main函数 (冲突!)
└── cache_service_example.go       # main函数 (冲突!)
```

**错误**: `main redeclared in this block`

---

## ✅ 推荐的解决方案

### 方案1: 子目录分离（推荐）⭐⭐⭐⭐⭐

```
examples/
├── basic/
│   └── main.go              # 基础示例
├── cache/
│   └── main.go              # 缓存基础示例
├── multilevel/
│   └── main.go              # 多级缓存示例
├── service/
│   └── main.go              # 服务封装示例
└── jwt/
    └── main.go              # JWT示例
```

**优点**:
- ✅ 每个示例独立编译
- ✅ 可以单独运行
- ✅ 结构清晰，易于查找

**使用方式**:
```bash
# 运行特定示例
go run examples/multilevel/main.go
go run examples/service/main.go

# 或进入目录运行
cd examples/multilevel
go run main.go
```

**迁移步骤**:
```bash
# 1. 创建目录结构
mkdir -p examples/{basic,cache,multilevel,service,jwt}

# 2. 移动文件
mv examples/cache_test_main.go examples/basic/main.go
mv examples/multilevel_cache_example.go examples/multilevel/main.go
mv examples/cache_service_example.go examples/service/main.go
mv examples/jwtutil_example.go examples/jwt/main.go

# 3. 确认每个文件都是 package main
```

---

### 方案2: 命令行工具模式 ⭐⭐⭐⭐

```
examples/
├── cmd/
│   ├── basic/
│   │   └── main.go
│   ├── multilevel/
│   │   └── main.go
│   ├── service/
│   │   └── main.go
│   └── jwt/
│       └── main.go
└── shared/
    ├── models.go         # 共享的数据结构
    └── utils.go          # 共享的工具函数
```

**优点**:
- ✅ 支持代码复用（shared包）
- ✅ 类似标准Go项目结构
- ✅ 易于扩展

**使用方式**:
```bash
go run examples/cmd/multilevel/main.go
go run examples/cmd/service/main.go
```

---

### 方案3: Build Tags ⭐⭐⭐

在每个示例文件顶部添加build tag:

```go
//go:build example_multilevel
// +build example_multilevel

package main
```

```go
//go:build example_service
// +build example_service

package main
```

**使用方式**:
```bash
go run -tags=example_multilevel examples/multilevel_cache_example.go
go run -tags=example_service examples/cache_service_example.go
```

**优点**:
- ✅ 不需要改变目录结构
- ✅ 可以选择性编译

**缺点**:
- ❌ 使用不够直观
- ❌ 需要记住tag名称

---

### 方案4: 示例测试文件 ⭐⭐⭐

将示例改为测试文件中的Example函数:

```go
// examples_test.go
package examples_test

func ExampleCache_Basic() {
    // 基础示例
    // Output:
    // 期望的输出
}

func ExampleCache_MultiLevel() {
    // 多级缓存示例
    // Output:
    // 期望的输出
}
```

**优点**:
- ✅ 可以作为文档
- ✅ go test会验证输出
- ✅ godoc会显示

**使用方式**:
```bash
go test -run=Example
```

---

## 🎯 推荐实施方案1

### 详细步骤

#### 1. 创建新的目录结构

```bash
# 在 examples 目录下执行
mkdir -p multilevel service basic jwt
```

#### 2. 移动并重命名文件

```bash
# 移动多级缓存示例
mv multilevel_cache_example.go multilevel/main.go

# 移动服务示例
mv cache_service_example.go service/main.go

# 移动基础示例
mv cache_test_main.go basic/main.go

# 移动JWT示例
mv jwtutil_example.go jwt/main.go
```

#### 3. 确认每个文件的package声明

```go
// multilevel/main.go
package main

// service/main.go
package main

// basic/main.go
package main

// jwt/main.go
package main
```

#### 4. 创建README

```bash
# examples/README.md
cat > README.md << 'EOF'
# Examples

## 运行示例

### 基础缓存示例
```bash
go run basic/main.go
```

### 多级缓存示例
```bash
go run multilevel/main.go
```

### 服务封装示例
```bash
go run service/main.go
```

### JWT工具示例
```bash
go run jwt/main.go
```

## 示例说明

- **basic**: 缓存包的基础使用
- **multilevel**: 多级缓存L1→L2→DB的完整示例
- **service**: 如何在服务中封装使用缓存
- **jwt**: JWT工具类的使用示例
EOF
```

#### 5. 更新导入路径（如有必要）

确保所有示例都正确导入:
```go
import (
    "gitee.com/wangsoft/go-library/cache"
    "gitee.com/wangsoft/go-library/cache/providers"
    "gitee.com/wangsoft/go-library/cache/serializers"
)
```

---

## 🔧 快速修复脚本

创建一个脚本自动完成迁移:

```bash
#!/bin/bash
# migrate_examples.sh

echo "开始迁移示例文件..."

cd examples

# 创建目录
mkdir -p multilevel service basic jwt

# 移动文件
if [ -f "multilevel_cache_example.go" ]; then
    mv multilevel_cache_example.go multilevel/main.go
    echo "✅ 已移动 multilevel_cache_example.go"
fi

if [ -f "cache_service_example.go" ]; then
    mv cache_service_example.go service/main.go
    echo "✅ 已移动 cache_service_example.go"
fi

if [ -f "cache_test_main.go" ]; then
    mv cache_test_main.go basic/main.go
    echo "✅ 已移动 cache_test_main.go"
fi

if [ -f "jwtutil_example.go" ]; then
    mv jwtutil_example.go jwt/main.go
    echo "✅ 已移动 jwtutil_example.go"
fi

# 创建README
cat > README.md << 'EOF'
# Examples

运行示例:
- 基础: `go run basic/main.go`
- 多级缓存: `go run multilevel/main.go`
- 服务封装: `go run service/main.go`
- JWT: `go run jwt/main.go`
EOF

echo "✅ 迁移完成!"
echo ""
echo "新的目录结构:"
tree -L 2
```

**使用方式**:
```bash
chmod +x migrate_examples.sh
./migrate_examples.sh
```

---

## 📋 迁移检查清单

- [ ] 创建子目录结构
- [ ] 移动所有示例文件
- [ ] 重命名为main.go
- [ ] 确认package声明正确
- [ ] 更新导入路径
- [ ] 测试每个示例可以独立运行
- [ ] 创建examples/README.md
- [ ] 更新主README.md中的示例链接

---

## 🎯 迁移后的最终结构

```
go-library/
├── cache/
│   ├── interfaces.go
│   ├── manager.go
│   ├── providers/
│   ├── serializers/
│   └── docs/
├── util/
│   ├── jwtutil/
│   └── ...
└── examples/
    ├── README.md              📖 示例说明
    ├── basic/
    │   └── main.go           ✅ 基础缓存示例
    ├── multilevel/
    │   └── main.go           ✅ 多级缓存示例
    ├── service/
    │   └── main.go           ✅ 服务封装示例
    └── jwt/
        └── main.go           ✅ JWT示例
```

---

## ✅ 优点总结

采用方案1后:

1. **清晰的结构**: 每个示例独立目录
2. **易于运行**: `go run multilevel/main.go`
3. **易于维护**: 修改一个示例不影响其他
4. **易于扩展**: 添加新示例只需新建目录
5. **符合惯例**: 符合Go社区的常见做法

---

## 🚀 立即行动

**推荐执行顺序**:

1. ✅ 创建migrate_examples.sh脚本
2. ✅ 执行脚本完成迁移
3. ✅ 测试每个示例
4. ✅ 更新文档链接
5. ✅ 提交代码

**预计耗时**: 10-15分钟

---

**现在就开始重构吧！** 🎯

