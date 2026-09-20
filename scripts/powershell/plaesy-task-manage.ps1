# plaesy-task-manage.ps1 - Task lifecycle management for autonomous workflows
# Moves task files between status directories (backlog → todo → doing → done/blocked)

param(
    [Parameter(Position = 0)]
    [ValidateSet("list", "list-status", "next", "start", "complete", "block", "unblock", "move", "show")]
    [string]$Command = "list",

    [Parameter(Position = 1)]
    [string]$TaskFile,

    [Parameter(Position = 2)]
    [string]$Status,

    [Parameter(Position = 3)]
    [string]$ToStatus,

    [string]$Reason
)

$ErrorActionPreference = "Stop"

# Find project root (.plaesy directory)
function Find-ProjectRoot {
    $current = Get-Location
    while ($current.Path -ne $current.Drive.Name + ":\") {
        if (Test-Path (Join-Path $current ".plaesy")) {
            return $current.Path
        }
        $current = Split-Path -Parent $current
    }
    throw "Project root not found (.plaesy directory not found)"
}

$ProjectRoot = Find-ProjectRoot
$TasksDir = Join-Path $ProjectRoot ".plaesy" "tasks"

# Helper functions
function Write-Info { Write-Host "[INFO] $args" -ForegroundColor Cyan }
function Write-Success { Write-Host "[✓] $args" -ForegroundColor Green }
function Write-Warning { Write-Host "[!] $args" -ForegroundColor Yellow }
function Write-Error { Write-Host "[ERROR] $args" -ForegroundColor Red; exit 1 }

function Validate-Status {
    param([string]$Status)
    $validStatuses = @("backlog", "todo", "doing", "done", "blocked")
    if ($Status -notin $validStatuses) {
        Write-Error "Invalid status: $Status (must be: $($validStatuses -join ', '))"
    }
}

function Get-NextTask {
    $backlogDir = Join-Path $TasksDir "backlog"
    if (-not (Test-Path $backlogDir)) {
        Write-Error "Backlog directory not found: $backlogDir"
    }

    $tasks = @(Get-ChildItem -Path $backlogDir -Filter "*.md" | Sort-Object -Descending | Select-Object -First 1)
    if ($tasks.Count -eq 0) {
        return $null
    }

    return $tasks[0].Name
}

function Move-Task {
    param(
        [string]$TaskFileParam,
        [string]$FromStatus,
        [string]$ToStatus
    )

    Validate-Status -Status $FromStatus
    Validate-Status -Status $ToStatus

    $fromPath = Join-Path $TasksDir $FromStatus $TaskFileParam
    $toPath = Join-Path $TasksDir $ToStatus $TaskFileParam

    if (-not (Test-Path $fromPath)) {
        Write-Error "Task not found: $fromPath"
    }

    Move-Item -Path $fromPath -Destination $toPath
    Write-Success "Moved: $TaskFileParam ($FromStatus → $ToStatus)"
}

function Pop-NextTask {
    $taskFile = Get-NextTask
    if ($null -eq $taskFile) {
        Write-Warning "No tasks in backlog"
        return $null
    }

    Move-Task -TaskFileParam $taskFile -FromStatus "backlog" -ToStatus "todo"
    return $taskFile
}

function Start-Task {
    param([string]$TaskFileParam)
    Move-Task -TaskFileParam $TaskFileParam -FromStatus "todo" -ToStatus "doing"
}

function Complete-Task {
    param([string]$TaskFileParam)
    Move-Task -TaskFileParam $TaskFileParam -FromStatus "doing" -ToStatus "done"
    Write-Success "Task completed: $TaskFileParam"
}

function Block-Task {
    param(
        [string]$TaskFileParam,
        [string]$Reason = "No reason provided"
    )

    Move-Task -TaskFileParam $TaskFileParam -FromStatus "doing" -ToStatus "blocked"

    # Append block reason to file
    $blockedPath = Join-Path $TasksDir "blocked" $TaskFileParam
    if (Test-Path $blockedPath) {
        Add-Content -Path $blockedPath -Value ""
        Add-Content -Path $blockedPath -Value "## Blocked Reason"
        Add-Content -Path $blockedPath -Value $Reason
    }
}

function Unblock-Task {
    param([string]$TaskFileParam)
    Move-Task -TaskFileParam $TaskFileParam -FromStatus "blocked" -ToStatus "todo"
    Write-Success "Task unblocked: $TaskFileParam"
}

function List-Tasks {
    param([string]$StatusParam)

    Validate-Status -Status $StatusParam

    $statusDir = Join-Path $TasksDir $StatusParam
    if (-not (Test-Path $statusDir)) {
        Write-Warning "No tasks in $StatusParam"
        return
    }

    $tasks = @(Get-ChildItem -Path $statusDir -Filter "*.md")
    $count = $tasks.Count

    Write-Host "$StatusParam ($count)" -ForegroundColor Cyan
    if ($count -gt 0) {
        foreach ($task in $tasks) {
            Write-Host "  - $($task.BaseName)" -ForegroundColor White
        }
    }
}

function List-AllTasks {
    Write-Host "`n📋 Task Summary`n" -ForegroundColor Cyan

    foreach ($status in @("backlog", "todo", "doing", "done", "blocked")) {
        List-Tasks -StatusParam $status
        Write-Host ""
    }
}

function Show-Task {
    param(
        [string]$TaskFileParam,
        [string]$StatusParam
    )

    $filePath = $null

    if ([string]::IsNullOrEmpty($StatusParam)) {
        # Search for task in all status directories
        foreach ($s in @("backlog", "todo", "doing", "done", "blocked")) {
            $testPath = Join-Path $TasksDir $s $TaskFileParam
            if (Test-Path $testPath) {
                $filePath = $testPath
                $StatusParam = $s
                break
            }
        }
    }
    else {
        $filePath = Join-Path $TasksDir $StatusParam $TaskFileParam
    }

    if (-not (Test-Path $filePath)) {
        Write-Error "Task not found: $TaskFileParam"
    }

    Write-Host "📄 Task: $TaskFileParam" -ForegroundColor Cyan
    Write-Host "Status: $StatusParam" -ForegroundColor Cyan
    Write-Host ""
    Get-Content -Path $filePath
}

# Main command dispatcher
switch ($Command) {
    "list" {
        List-AllTasks
    }
    "list-status" {
        if ([string]::IsNullOrEmpty($Status)) {
            Write-Error "Usage: plaesy-task-manage.ps1 list-status {backlog|todo|doing|done|blocked}"
        }
        List-Tasks -StatusParam $Status
    }
    "next" {
        $task = Pop-NextTask
        if ($null -ne $task) {
            Write-Host $task
        }
    }
    "start" {
        if ([string]::IsNullOrEmpty($TaskFile)) {
            Write-Error "Usage: plaesy-task-manage.ps1 start <task_file>"
        }
        Start-Task -TaskFileParam $TaskFile
    }
    "complete" {
        if ([string]::IsNullOrEmpty($TaskFile)) {
            Write-Error "Usage: plaesy-task-manage.ps1 complete <task_file>"
        }
        Complete-Task -TaskFileParam $TaskFile
    }
    "block" {
        if ([string]::IsNullOrEmpty($TaskFile)) {
            Write-Error "Usage: plaesy-task-manage.ps1 block <task_file> [reason]"
        }
        Block-Task -TaskFileParam $TaskFile -Reason $Reason
    }
    "unblock" {
        if ([string]::IsNullOrEmpty($TaskFile)) {
            Write-Error "Usage: plaesy-task-manage.ps1 unblock <task_file>"
        }
        Unblock-Task -TaskFileParam $TaskFile
    }
    "move" {
        if ([string]::IsNullOrEmpty($TaskFile) -or [string]::IsNullOrEmpty($Status) -or [string]::IsNullOrEmpty($ToStatus)) {
            Write-Error "Usage: plaesy-task-manage.ps1 move <task_file> <from_status> <to_status>"
        }
        Move-Task -TaskFileParam $TaskFile -FromStatus $Status -ToStatus $ToStatus
    }
    "show" {
        if ([string]::IsNullOrEmpty($TaskFile)) {
            Write-Error "Usage: plaesy-task-manage.ps1 show <task_file> [status]"
        }
        Show-Task -TaskFileParam $TaskFile -StatusParam $Status
    }
    default {
        Write-Error "Usage: plaesy-task-manage.ps1 {list|list-status|next|start|complete|block|unblock|move|show}"
    }
}
