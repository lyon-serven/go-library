# 测试所有示例的 PowerShell 脚本

Write-Host "=== 开始测试 go-library 示例 ===" -ForegroundColor Cyan
Write-Host ""

$ErrorActionPreference = "Continue"
$OriginalLocation = Get-Location

function Test-Example {
    param(
        [string]$Name,
        [string]$Path,
        [string]$File
    )
    
    Write-Host "【测试 $Name】" -ForegroundColor Green
    Write-Host "目录: $Path" -ForegroundColor Gray
    
    try {
        Set-Location $Path
        $output = go run $File 2>&1
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ $Name 测试通过" -ForegroundColor Green
            Write-Host $output
        } else {
            Write-Host "❌ $Name 测试失败" -ForegroundColor Red
            Write-Host $output
        }
    }
    catch {
        Write-Host "❌ $Name 测试异常: $_" -ForegroundColor Red
    }
    finally {
        Set-Location $OriginalLocation
    }
    
    Write-Host ""
    Write-Host "----------------------------------------" -ForegroundColor Gray
    Write-Host ""
}

# 测试 JWT 工具
Test-Example -Name "JWT 工具" `
             -Path "examples\jwtutil" `
             -File "jwtutil_example.go"

# 测试基础缓存
Test-Example -Name "基础缓存" `
             -Path "examples\cache" `
             -File "cache_basic_example.go"

# 测试多级缓存
Test-Example -Name "多级缓存" `
             -Path "examples\cache" `
             -File "multilevel_cache_example.go"

# 注意：数据库示例需要配置数据库连接，默认跳过
Write-Host "⚠️  数据库示例需要配置数据库连接信息，已跳过" -ForegroundColor Yellow
Write-Host "   如需测试，请手动修改配置后运行：" -ForegroundColor Gray
Write-Host "   - cd examples\db\gorm && go run gorm_gen_example.go" -ForegroundColor Gray
Write-Host "   - cd examples\db\xorm && go run xorm_gen_example.go" -ForegroundColor Gray
Write-Host ""

Write-Host "=== 测试完成 ===" -ForegroundColor Cyan
