# MyLib Module Path Update Script
# Usage: .\update-module.ps1 -RepoURL "github.com/username/mylib"

param(
    [Parameter(Mandatory=$true)]
    [string]$RepoURL,
    
    [Parameter(Mandatory=$false)]
    [switch]$DryRun = $false
)

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "  MyLib Module Path Update" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

# Validate RepoURL format
if ($RepoURL -notmatch '^(github\.com|gitee\.com)/[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+$') {
    Write-Host "ERROR: Invalid repository URL format" -ForegroundColor Red
    Write-Host "  Correct format: github.com/username/repository" -ForegroundColor Yellow
    Write-Host "  Or: gitee.com/username/repository" -ForegroundColor Yellow
    exit 1
}

Write-Host "Target Repository: $RepoURL" -ForegroundColor Green
Write-Host ""

if ($DryRun) {
    Write-Host "[DRY RUN MODE] No files will be modified" -ForegroundColor Yellow
    Write-Host ""
}

# Backup warning
Write-Host "WARNING: Recommended actions before proceeding:" -ForegroundColor Yellow
Write-Host "  1. Commit your current code: git add . && git commit -m 'backup'" -ForegroundColor Gray
Write-Host "  2. Or create a backup branch: git checkout -b backup" -ForegroundColor Gray
Write-Host ""

$continue = Read-Host "Continue? (y/n)"
if ($continue -ne 'y' -and $continue -ne 'Y') {
    Write-Host "Operation cancelled" -ForegroundColor Red
    exit 0
}

Write-Host ""
Write-Host "Processing..." -ForegroundColor Cyan
Write-Host ""

# Step 1: Update go.mod
Write-Host "Step 1: Updating go.mod" -ForegroundColor Cyan

$goModPath = "go.mod"
if (Test-Path $goModPath) {
    $goModContent = Get-Content $goModPath -Raw
    $newGoModContent = $goModContent -replace '^module mylib', "module $RepoURL"
    
    if (!$DryRun) {
        $newGoModContent | Set-Content $goModPath -NoNewline
        Write-Host "  OK: go.mod updated" -ForegroundColor Green
    } else {
        Write-Host "  [Preview] go.mod will be updated to:" -ForegroundColor Yellow
        Write-Host "  module $RepoURL" -ForegroundColor Gray
    }
} else {
    Write-Host "  ERROR: go.mod not found" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Step 2: Find all Go files
Write-Host "Step 2: Finding Go files to update" -ForegroundColor Cyan

$goFiles = Get-ChildItem -Path . -Filter *.go -Recurse -File | Where-Object {
    $_.FullName -notmatch '\\\.git\\' -and
    $_.FullName -notmatch '\\vendor\\'
}

Write-Host "  Found $($goFiles.Count) Go files" -ForegroundColor Gray
Write-Host ""

# Step 3: Update import paths
Write-Host "Step 3: Updating import paths" -ForegroundColor Cyan

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
        $relativePath = $file.FullName.Replace((Get-Location).Path, "").TrimStart('\')
        
        if (!$DryRun) {
            $content | Set-Content $file.FullName -NoNewline
            Write-Host "  OK: $relativePath" -ForegroundColor Green
        } else {
            Write-Host "  [Preview] $relativePath" -ForegroundColor Yellow
        }
        
        $updatedCount++
    }
}

Write-Host ""
Write-Host "  Updated $updatedCount files" -ForegroundColor Green
Write-Host ""

# Step 4: Run go mod tidy
if (!$DryRun) {
    Write-Host "Step 4: Running go mod tidy" -ForegroundColor Cyan
    
    $tidyOutput = & go mod tidy 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  OK: go mod tidy completed" -ForegroundColor Green
    } else {
        Write-Host "  WARNING: go mod tidy encountered issues:" -ForegroundColor Yellow
        Write-Host $tidyOutput -ForegroundColor Gray
    }
    Write-Host ""
}

# Step 5: Test build
if (!$DryRun) {
    Write-Host "Step 5: Testing build" -ForegroundColor Cyan
    
    $buildOutput = & go build ./... 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  OK: Build successful" -ForegroundColor Green
    } else {
        Write-Host "  ERROR: Build failed:" -ForegroundColor Red
        Write-Host $buildOutput -ForegroundColor Gray
        Write-Host ""
        Write-Host "Please fix the errors above and try again" -ForegroundColor Yellow
        exit 1
    }
    Write-Host ""
}

# Summary
Write-Host "================================================" -ForegroundColor Cyan
Write-Host "  Update Complete!" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

if ($DryRun) {
    Write-Host "[DRY RUN MODE] No files were modified" -ForegroundColor Yellow
    Write-Host "  Remove -DryRun parameter to perform actual update" -ForegroundColor Yellow
} else {
    Write-Host "Next Steps:" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "1. Initialize Git repository (if not already):" -ForegroundColor White
    Write-Host "   git init" -ForegroundColor Gray
    Write-Host ""
    Write-Host "2. Add all files:" -ForegroundColor White
    Write-Host "   git add ." -ForegroundColor Gray
    Write-Host ""
    Write-Host "3. Commit changes:" -ForegroundColor White
    Write-Host "   git commit -m `"feat: configure module for $RepoURL`"" -ForegroundColor Gray
    Write-Host ""
    Write-Host "4. Add remote repository:" -ForegroundColor White
    Write-Host "   git remote add origin https://$RepoURL.git" -ForegroundColor Gray
    Write-Host ""
    Write-Host "5. Push to remote:" -ForegroundColor White
    Write-Host "   git branch -M main" -ForegroundColor Gray
    Write-Host "   git push -u origin main" -ForegroundColor Gray
    Write-Host ""
    Write-Host "6. Create version tag:" -ForegroundColor White
    Write-Host "   git tag v1.0.0" -ForegroundColor Gray
    Write-Host "   git push origin v1.0.0" -ForegroundColor Gray
    Write-Host ""
    Write-Host "7. Use in other projects:" -ForegroundColor White
    Write-Host "   go get $RepoURL@v1.0.0" -ForegroundColor Gray
    Write-Host ""
    
    # Show additional config for Gitee
    if ($RepoURL -match '^gitee\.com') {
        Write-Host "Additional configuration for Gitee:" -ForegroundColor Yellow
        Write-Host "   go env -w GOPROXY=https://goproxy.cn,direct" -ForegroundColor Gray
        Write-Host "   go env -w GOPRIVATE=$RepoURL" -ForegroundColor Gray
        Write-Host ""
    }
}

Write-Host "For more information, see: README.md" -ForegroundColor Cyan
Write-Host ""
