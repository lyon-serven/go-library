# MyLib 模块配置脚本
# 使用方法: .\update-module-path.ps1 -RepoURL "github.com/yourusername/mylib"

param(
    [Parameter(Mandatory=$true)]
    [string]$RepoURL,
    
    [Parameter(Mandatory=$false)]
    [switch]$DryRun = $false
)

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "  MyLib 模块路径更新脚本" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

# 验证 RepoURL 格式
if ($RepoURL -notmatch '^(github\.com|gitee\.com)/[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+$') {
    Write-Host "❌ 错误: 仓库URL格式不正确" -ForegroundColor Red
    Write-Host "   正确格式: github.com/username/repository" -ForegroundColor Yellow
    Write-Host "   或: gitee.com/username/repository" -ForegroundColor Yellow
    exit 1
}

Write-Host "📦 目标仓库: $RepoURL" -ForegroundColor Green
Write-Host ""

if ($DryRun) {
    Write-Host "🔍 [预览模式] 不会实际修改文件" -ForegroundColor Yellow
    Write-Host ""
}

# 备份提示
Write-Host "⚠️  建议操作:" -ForegroundColor Yellow
Write-Host "   1. 先提交现有代码: git add . && git commit -m 'backup'" -ForegroundColor Gray
Write-Host "   2. 或创建备份分支: git checkout -b backup" -ForegroundColor Gray
Write-Host ""

$continue = Read-Host "是否继续? (y/n)"
if ($continue -ne 'y' -and $continue -ne 'Y') {
    Write-Host "❌ 操作已取消" -ForegroundColor Red
    exit 0
}

Write-Host ""
Write-Host "开始处理..." -ForegroundColor Cyan
Write-Host ""

# 步骤1: 更新 go.mod
Write-Host "📝 步骤1: 更新 go.mod" -ForegroundColor Cyan

$goModPath = "go.mod"
if (Test-Path $goModPath) {
    $goModContent = Get-Content $goModPath -Raw
    $newGoModContent = $goModContent -replace '^module mylib', "module $RepoURL"
    
    if (!$DryRun) {
        $newGoModContent | Set-Content $goModPath -NoNewline
        Write-Host "   ✅ go.mod 已更新" -ForegroundColor Green
    } else {
        Write-Host "   🔍 [预览] go.mod 将更新为:" -ForegroundColor Yellow
        Write-Host "   module $RepoURL" -ForegroundColor Gray
    }
} else {
    Write-Host "   ❌ go.mod 文件不存在" -ForegroundColor Red
    exit 1
}

Write-Host ""

# 步骤2: 查找所有需要更新的 .go 文件
Write-Host "📝 步骤2: 查找需要更新的 Go 文件" -ForegroundColor Cyan

$goFiles = Get-ChildItem -Path . -Filter *.go -Recurse -File | Where-Object {
    $_.FullName -notmatch '\\\.git\\' -and
    $_.FullName -notmatch '\\vendor\\'
}

Write-Host "   找到 $($goFiles.Count) 个 Go 文件" -ForegroundColor Gray
Write-Host ""

# 步骤3: 更新导入路径
Write-Host "📝 步骤3: 更新导入路径" -ForegroundColor Cyan

$replacements = @{
    '"mylib/cache"' = "`"$RepoURL/cache`""
    '"mylib/config"' = "`"$RepoURL/config`""
    '"mylib/util/authutil"' = "`"$RepoURL/util/authutil`""
    '"mylib/util/cryptoutil"' = "`"$RepoURL/util/cryptoutil`""
    '"mylib/util/httputil"' = "`"$RepoURL/util/httputil`""
    '"mylib/util/timeutil"' = "`"$RepoURL/util/timeutil`""
    '"mylib/stringutil"' = "`"$RepoURL/stringutil`""
    '"mylib/sliceutil"' = "`"$RepoURL/sliceutil`""
    '"mylib/fileutil"' = "`"$RepoURL/fileutil`""
    '"mylib/validator"' = "`"$RepoURL/validator`""
    '"mylib/cryptoutil"' = "`"$RepoURL/cryptoutil`""
    '"mylib/httputil"' = "`"$RepoURL/httputil`""
    '"mylib/timeutil"' = "`"$RepoURL/timeutil`""
    '"mylib/examples/model"' = "`"$RepoURL/examples/model`""
}

$updatedCount = 0
$fileCount = 0

foreach ($file in $goFiles) {
    $content = Get-Content $file.FullName -Raw
    $originalContent = $content
    $fileModified = $false
    
    foreach ($old in $replacements.Keys) {
        $new = $replacements[$old]
        if ($content -match [regex]::Escape($old)) {
            $content = $content -replace [regex]::Escape($old), $new
            $fileModified = $true
        }
    }
    
    if ($fileModified) {
        $fileCount++
        $relativePath = $file.FullName.Replace((Get-Location).Path, "").TrimStart('\')
        
        if (!$DryRun) {
            $content | Set-Content $file.FullName -NoNewline
            Write-Host "   ✅ $relativePath" -ForegroundColor Green
        } else {
            Write-Host "   🔍 [预览] $relativePath" -ForegroundColor Yellow
        }
        
        $updatedCount++
    }
}

Write-Host ""
Write-Host "   已更新 $updatedCount 个文件" -ForegroundColor Green
Write-Host ""

# 步骤4: 运行 go mod tidy
if (!$DryRun) {
    Write-Host "📝 步骤4: 运行 go mod tidy" -ForegroundColor Cyan
    
    $tidyOutput = & go mod tidy 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "   ✅ go mod tidy 成功" -ForegroundColor Green
    } else {
        Write-Host "   ⚠️  go mod tidy 遇到问题:" -ForegroundColor Yellow
        Write-Host $tidyOutput -ForegroundColor Gray
    }
    Write-Host ""
}

# 步骤5: 测试编译
if (!$DryRun) {
    Write-Host "📝 步骤5: 测试编译" -ForegroundColor Cyan
    
    $buildOutput = & go build ./... 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "   ✅ 编译成功" -ForegroundColor Green
    } else {
        Write-Host "   ❌ 编译失败:" -ForegroundColor Red
        Write-Host $buildOutput -ForegroundColor Gray
        Write-Host ""
        Write-Host "请检查上述错误，修复后重新运行" -ForegroundColor Yellow
        exit 1
    }
    Write-Host ""
}

# 完成总结
Write-Host "================================================" -ForegroundColor Cyan
Write-Host "  ✅ 更新完成!" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

if ($DryRun) {
    Write-Host "🔍 这是预览模式，没有实际修改文件" -ForegroundColor Yellow
    Write-Host "   移除 -DryRun 参数来执行实际更新" -ForegroundColor Yellow
} else {
    Write-Host "📋 后续步骤:" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "1. 创建 Git 仓库（如果还没有）:" -ForegroundColor White
    Write-Host "   git init" -ForegroundColor Gray
    Write-Host ""
    Write-Host "2. 添加所有文件:" -ForegroundColor White
    Write-Host "   git add ." -ForegroundColor Gray
    Write-Host ""
    Write-Host "3. 提交更改:" -ForegroundColor White
    Write-Host "   git commit -m `"feat: configure module for $RepoURL`"" -ForegroundColor Gray
    Write-Host ""
    Write-Host "4. 添加远程仓库:" -ForegroundColor White
    Write-Host "   git remote add origin https://$RepoURL.git" -ForegroundColor Gray
    Write-Host ""
    Write-Host "5. 推送到远程:" -ForegroundColor White
    Write-Host "   git branch -M main" -ForegroundColor Gray
    Write-Host "   git push -u origin main" -ForegroundColor Gray
    Write-Host ""
    Write-Host "6. 创建版本标签:" -ForegroundColor White
    Write-Host "   git tag v1.0.0" -ForegroundColor Gray
    Write-Host "   git push origin v1.0.0" -ForegroundColor Gray
    Write-Host ""
    Write-Host "7. 在其他项目中使用:" -ForegroundColor White
    Write-Host "   go get $RepoURL@v1.0.0" -ForegroundColor Gray
    Write-Host ""
    
    # 如果是 Gitee，显示额外配置
    if ($RepoURL -match '^gitee\.com') {
        Write-Host "⚠️  Gitee 额外配置:" -ForegroundColor Yellow
        Write-Host "   go env -w GOPROXY=https://goproxy.cn,direct" -ForegroundColor Gray
        Write-Host "   go env -w GOPRIVATE=$RepoURL" -ForegroundColor Gray
        Write-Host ""
    }
}

Write-Host "📚 更多信息请查看: 配置指南.md" -ForegroundColor Cyan
Write-Host ""
