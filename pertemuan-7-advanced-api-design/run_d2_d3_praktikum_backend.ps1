[CmdletBinding()]
param(
    [switch]$DryRun,
    [switch]$ReadOnlyPreflight,
    [switch]$ExecuteWriteTests,
    [switch]$D3Only
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$ProjectRoot = Split-Path -Parent $PSCommandPath
$EnvFile = Join-Path $ProjectRoot '.env'
$EvidenceDir = Join-Path $ProjectRoot 'evidence'
$RunTag = [DateTime]::UtcNow.ToString('yyyyMMddHHmmssfff')
$BaseUrl = 'http://127.0.0.1:3077'
$TestUsername = 'p7_test'
$TestEmail = "p7_test_$RunTag@example.test"
$TestPassword = "P7Test!$([Guid]::NewGuid().ToString('N').Substring(0,16))9a"
$D2Name = "P7 D2 $RunTag"
$D3Name = "P7 D3 Insert $RunTag"
$D2Nim = 900000000 + [int]([DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds() % 90000000)
$D3BatchNims = @(1..6 | ForEach-Object { $D2Nim + $_ })
$D3InsertNim = $D2Nim + 7
$D2StudentId = $null
$D3StudentId = $null
$D3BatchStudentIds = [System.Collections.Generic.List[int]]::new()
$TestUserId = $null
$BaselineUserIds = @()
$BaselineStudentIds = @()
$DbReady = $false
$Server = $null
$CleanupProblems = [System.Collections.Generic.List[string]]::new()
$ExecutionProblem = $null
$PreviousEnvironment = @{}
$script:HttpClient = $null
$script:LastPsqlExitCode = $null

function Read-DotEnv([string]$Path) {
    if (-not (Test-Path -LiteralPath $Path -PathType Leaf)) {
        throw ".env tidak ditemukan: $Path"
    }
    return ConvertFrom-DotEnvLines @(Get-Content -LiteralPath $Path)
}

function Unwrap-DotEnvValue([string]$Value) {
    # Hanya lepaskan sepasang quote yang benar-benar membungkus seluruh nilai.
    # Karakter lain, termasuk quote di tengah password/URL, dipertahankan apa adanya.
    if ($Value.Length -ge 2) {
        $first = $Value[0]
        $last = $Value[$Value.Length - 1]
        if (($first -eq '"' -and $last -eq '"') -or ($first -eq "'" -and $last -eq "'")) {
            return $Value.Substring(1, $Value.Length - 2)
        }
    }
    return $Value
}

function ConvertFrom-DotEnvLines([string[]]$Lines) {
    $values = @{}
    foreach ($line in $Lines) {
        if ($line -match '^\s*#' -or [string]::IsNullOrWhiteSpace($line)) { continue }
        if ($line -match '^\s*([^=\s]+)\s*=(.*)$') {
            $values[$Matches[1]] = Unwrap-DotEnvValue $Matches[2]
        }
    }
    return $values
}

function Set-ProcessEnvironment([hashtable]$Values) {
    foreach ($key in @('DB_HOST','DB_PORT','DB_USER','DB_PASSWORD','DB_NAME','DB_SSLMODE','DB_MAX_CONNS','JWT_SECRET','JWT_ISSUER','JWT_ACCESS_TTL_MINUTES','JWT_REFRESH_TTL_DAYS','PGPASSWORD','PGSSLMODE')) {
        $envPath = "Env:$key"
        if (Test-Path -LiteralPath $envPath) {
            $PreviousEnvironment[$key] = (Get-Item -LiteralPath $envPath).Value
        } else {
            $PreviousEnvironment[$key] = $null
        }
        if ($Values.ContainsKey($key)) {
            Set-Item -LiteralPath $envPath -Value $Values[$key]
        }
    }
    $PreviousEnvironment['APP_PORT'] = if (Test-Path -LiteralPath 'Env:APP_PORT') { (Get-Item -LiteralPath 'Env:APP_PORT').Value } else { $null }
    Set-Item -LiteralPath 'Env:APP_PORT' -Value '3077'
}

function Restore-ProcessEnvironment {
    foreach ($key in $PreviousEnvironment.Keys) {
        $envPath = "Env:$key"
        if ($null -eq $PreviousEnvironment[$key]) {
            if (Test-Path -LiteralPath $envPath) {
                Remove-Item -LiteralPath $envPath
            }
        } else {
            Set-Item -LiteralPath $envPath -Value $PreviousEnvironment[$key]
        }
    }
}

function Get-SafePsqlFailure([object[]]$Output) {
    $text = ($Output | ForEach-Object { [string]$_ }) -join "`n"
    $category = if ($text -match '(?i)password authentication failed|authentication failed|password gagal') {
        'AUTHENTICATION_PASSWORD_REJECTED'
    } elseif ($text -match '(?i)could not connect|connection refused|timeout|could not translate host') {
        'SERVER_UNREACHABLE'
    } elseif ($text -match '(?i)database .* does not exist') {
        'DATABASE_NOT_FOUND'
    } elseif ($text -match '(?i)permission denied|must be owner|role .* does not exist') {
        'ROLE_OR_ACCESS_REJECTED'
    } elseif ($text -match '(?i)ssl|certificate') {
        'SSL_CONFIGURATION'
    } else {
        'OTHER_PSQL_FAILURE'
    }
    $sqlState = if ($text -match '(?i)SQLSTATE\s*[:=]?\s*([0-9A-Z]{5})') { $Matches[1] } else { 'not_available' }
    return "category=$category; sqlstate=$sqlState"
}

function Invoke-Psql([string]$Sql) {
    # Gunakan flag libpq eksplisit. psql tidak membaca DB_HOST/DB_USER aplikasi.
    # Password hanya diteruskan melalui PGPASSWORD dan tidak pernah diletakkan di argv.
    $args = @('-X','-v','ON_ERROR_STOP=1','-At','-h',$script:PgHost,'-p',$script:PgPort,'-U',$script:PgUser,'-d',$script:DbName,'-c',$Sql)
    $output = & $script:PsqlPath @args 2>&1
    $script:LastPsqlExitCode = [int]$LASTEXITCODE
    if ($LASTEXITCODE -ne 0) {
        throw "psql gagal: $(Get-SafePsqlFailure @($output)); exit_code=$script:LastPsqlExitCode"
    }
    return ConvertFrom-PsqlOutput @($output)
}

function ConvertFrom-PsqlOutput([object[]]$Output) {
    return @($Output | Where-Object { $_ -is [string] -and $_.Trim() -ne '' } | ForEach-Object { $_.Trim() })
}

function ConvertTo-DbScalar([string[]]$Rows) {
    $normalizedRows = @($Rows)
    if ($normalizedRows.Count -ne 1) {
        throw "SELECT scalar mengembalikan $($normalizedRows.Count) baris, bukan tepat satu."
    }
    return [string]$normalizedRows[0]
}

function Get-DbScalar([string]$Sql) {
    $rows = Invoke-Psql $Sql
    return ConvertTo-DbScalar $rows
}

function Invoke-DbCommand([string]$Sql) {
    Assert-WriteAllowed $Sql
    [void](Invoke-Psql $Sql)
}

function Get-IdList([string]$Sql) {
    $value = Get-DbScalar $Sql
    if ([string]::IsNullOrWhiteSpace($value)) { return @() }
    return @($value.Split(',') | ForEach-Object { [int]$_ })
}

function Get-StudentCreatedAt([int]$StudentId) {
    return Get-DbScalar ('SELECT to_char(created_at AT TIME ZONE ''UTC'', ''YYYY-MM-DD"T"HH24:MI:SS.US"Z"'') FROM students WHERE id = ' + $StudentId + ';')
}

function Assert-True([bool]$Condition, [string]$Message) {
    if (-not $Condition) { throw $Message }
}

function Test-SafeTestDatabaseName([string]$Name) {
    if ([string]::IsNullOrWhiteSpace($Name)) { return $false }
    $normalized = $Name.Trim().ToLowerInvariant()
    if ($normalized -in @('praktikum_backend','postgres','template0','template1')) { return $false }
    return $normalized.Contains('test')
}

function Assert-TestDatabaseName([string]$Name) {
    Assert-True (Test-SafeTestDatabaseName $Name) "DB_NAME harus menunjuk database test terpisah dan tidak boleh praktikum_backend. current/DB_NAME=$Name"
}

function Test-WriteSql([string]$Sql) {
    return $Sql -match '(?is)^\s*(INSERT|UPDATE|DELETE|MERGE|TRUNCATE|CREATE|ALTER|DROP|GRANT|REVOKE|COMMENT|COPY)\b'
}

function Assert-WriteAllowed([string]$Sql) {
    if (-not (Test-WriteSql $Sql)) { return }
    Assert-True ([bool]$ExecuteWriteTests) 'Operasi tulis membutuhkan parameter eksplisit -ExecuteWriteTests.'
    Assert-TestDatabaseName $script:DbName
    Assert-True ($script:DbName.Trim().ToLowerInvariant() -ne 'praktikum_backend') 'Operasi tulis ke praktikum_backend ditolak.'
}

function Invoke-Api([string]$Method, [string]$Path, [string]$Token = $null, [string]$Accept = 'application/json', [string]$JsonBody = $null) {
    $request = [System.Net.Http.HttpRequestMessage]::new([System.Net.Http.HttpMethod]::new($Method), "$script:BaseUrl$Path")
    try {
        if ($null -ne $Token) { [void]$request.Headers.TryAddWithoutValidation('Authorization', "Bearer $Token") }
        if ($null -ne $Accept) { [void]$request.Headers.TryAddWithoutValidation('Accept', $Accept) }
        if ($null -ne $JsonBody) { $request.Content = [System.Net.Http.StringContent]::new($JsonBody, [System.Text.Encoding]::UTF8, 'application/json') }
        $response = $script:HttpClient.SendAsync($request).GetAwaiter().GetResult()
        $body = $response.Content.ReadAsStringAsync().GetAwaiter().GetResult()
        $parsed = $null
        if ($body.Trim() -ne '') { $parsed = $body | ConvertFrom-Json -Depth 20 }
        return [pscustomobject]@{ Status = [int]$response.StatusCode; ContentType = [string]$response.Content.Headers.ContentType; Body = $body; Json = $parsed }
    } finally {
        $request.Dispose()
    }
}

function Require-Status($Result, [int]$Expected, [string]$Label) {
    Write-Host "$Label HTTP $($Result.Status)"
    if ($Result.Status -ne $Expected) {
        $code = if ($null -ne $Result.Json) { [string]$Result.Json.code } else { '' }
        throw "$Label mengembalikan HTTP $($Result.Status), diharapkan $Expected. code=$code"
    }
}

function ConvertTo-StudentPage($Result, [string]$Label) {
    Require-Status $result 200 $Label
    $ids = @($result.Json.data | ForEach-Object { [int]$_.id })
    $meta = $result.Json.meta
    $hasMoreProperty = $meta.PSObject.Properties['has_more']
    if ($null -eq $hasMoreProperty) { throw "$Label response tidak memiliki meta.has_more." }
    $hasMore = [bool]$hasMoreProperty.Value
    $cursorProperty = $meta.PSObject.Properties['next_cursor']
    $cursor = $null
    if ($null -ne $cursorProperty -and -not [string]::IsNullOrWhiteSpace([string]$cursorProperty.Value)) {
        $cursor = [string]$cursorProperty.Value
    }
    if ($hasMore -and [string]::IsNullOrWhiteSpace($cursor)) {
        throw "$Label response melanggar kontrak: meta.has_more=true tanpa meta.next_cursor yang terisi."
    }
    if (-not $hasMore) { $cursor = $null }
    $cursorLog = if ($null -eq $cursor) { '<none>' } else { $cursor }
    Write-Host "$Label IDs=[$($ids -join ',')] has_more=$hasMore next_cursor=$cursorLog"
    return [pscustomobject]@{ IDs = $ids; HasMore = $hasMore; NextCursor = $cursor }
}

function Get-StudentPage([string]$Path, [string]$Token, [string]$Label) {
    $result = Invoke-Api -Method 'GET' -Path $Path -Token $Token
    return ConvertTo-StudentPage -Result $result -Label $Label
}

function Get-AllStudentPages([string]$Token, [string]$FirstPath, [string]$Label) {
    $allIds = [System.Collections.Generic.List[int]]::new()
    $seenCursors = [System.Collections.Generic.HashSet[string]]::new()
    $path = $FirstPath
    $pageNumber = 1
    while ($true) {
        $page = Get-StudentPage -Path $path -Token $Token -Label "$Label page $pageNumber"
        foreach ($id in $page.IDs) { $allIds.Add($id) }
        if (-not $page.HasMore) { break }
        Assert-True (-not [string]::IsNullOrWhiteSpace($page.NextCursor)) "$Label page $pageNumber has_more=true tanpa next_cursor"
        Assert-True $seenCursors.Add($page.NextCursor) "$Label page $pageNumber mengulangi next_cursor; pagination dihentikan."
        $encoded = [uri]::EscapeDataString($page.NextCursor)
        $path = "/api/v1/students/?limit=2&cursor=$encoded"
        $pageNumber++
        Assert-True ($pageNumber -le 100) "$Label melebihi batas 100 halaman"
    }
    return @($allIds)
}

function Compare-IdList([int[]]$Actual, [int[]]$Expected, [string]$Label) {
    $actualText = $Actual -join ','
    $expectedText = $Expected -join ','
    Assert-True ($actualText -eq $expectedText) "$Label tidak konsisten. actual=[$actualText] expected=[$expectedText]"
}

function Test-LocalPreflight {
    $parserErrors = 0
    try {
        $dummy = ConvertFrom-DotEnvLines @(
            'PLAIN=postgres://user:p@ss?sslmode=disable',
            'DOUBLE="a$=b&c?d"',
            "SINGLE='quoted value = keep'",
            'EMBEDDED=left"middle"right',
            'EMPTY=""'
        )
        Assert-True ($dummy['PLAIN'] -ceq 'postgres://user:p@ss?sslmode=disable') 'Nilai unquoted berubah.'
        Assert-True ($dummy['DOUBLE'] -ceq 'a$=b&c?d') 'Nilai double-quoted berubah.'
        Assert-True ($dummy['SINGLE'] -ceq 'quoted value = keep') 'Nilai single-quoted berubah.'
        Assert-True ($dummy['EMBEDDED'] -ceq 'left"middle"right') 'Quote di tengah nilai berubah.'
        Assert-True ($dummy['EMPTY'] -ceq '') 'Nilai empty quoted tidak diparse benar.'
        Write-Output 'ENV_QUOTED_VALUES=PASS'
        Write-Output 'ENV_UNQUOTED_VALUES=PASS'
    } catch {
        $parserErrors++
        throw
    } finally {
        Write-Output "PARSER_ERRORS=$parserErrors"
    }

    $rows = ConvertFrom-PsqlOutput @('  p7_test_backend  ', '', '   ')
    $normalizedRows = @($rows)
    Assert-True ($normalizedRows.Count -eq 1 -and $normalizedRows[0] -eq 'p7_test_backend') 'Normalisasi output psql gagal.'
    Assert-True ((ConvertTo-DbScalar $rows) -eq 'p7_test_backend') 'Parsing SELECT scalar gagal.'
    Write-Output 'PSQL_PARSING=PASS'

    $middlePageResponse = [pscustomobject]@{
        Status = 200
        Json = [pscustomobject]@{
            data = @([pscustomobject]@{ id = 101 }, [pscustomobject]@{ id = 100 })
            meta = [pscustomobject]@{ has_more = $true; next_cursor = 'dummy-next-cursor' }
        }
    }
    $middlePage = ConvertTo-StudentPage -Result $middlePageResponse -Label 'LOCAL middle GET helper'
    $middleResults = @($middlePage)
    Assert-True ($middleResults.Count -eq 1) 'GET helper tidak mengembalikan tepat satu objek.'
    Assert-True ($middleResults[0].HasMore -is [bool] -and $middleResults[0].HasMore) 'HasMore halaman tengah tidak dapat diakses sebagai boolean.'
    Assert-True ($middleResults[0].NextCursor -eq 'dummy-next-cursor') 'NextCursor halaman tengah tidak dapat diakses.'
    Assert-True ((@($middleResults[0].IDs) -join ',') -eq '101,100') 'Daftar ID helper tidak benar.'

    $lastPageResponse = [pscustomobject]@{
        Status = 200
        Json = [pscustomobject]@{
            data = @([pscustomobject]@{ id = 99 })
            meta = [pscustomobject]@{ has_more = $false }
        }
    }
    $lastPage = ConvertTo-StudentPage -Result $lastPageResponse -Label 'LOCAL last GET helper'
    Assert-True ((@($lastPage).Count -eq 1 -and -not $lastPage.HasMore -and $null -eq $lastPage.NextCursor)) 'Halaman terakhir tanpa next_cursor tidak diproses normal.'

    $invalidCursorResponse = [pscustomobject]@{
        Status = 200
        Json = [pscustomobject]@{
            data = @([pscustomobject]@{ id = 98 })
            meta = [pscustomobject]@{ has_more = $true }
        }
    }
    $contractError = $null
    try {
        [void](ConvertTo-StudentPage -Result $invalidCursorResponse -Label 'LOCAL invalid GET helper')
    } catch {
        $contractError = $_
    }
    Assert-True ($null -ne $contractError -and $contractError.Exception.Message -match 'melanggar kontrak') 'has_more=true tanpa next_cursor tidak menghasilkan error kontrak.'
    Write-Output 'GET_HELPER_MIDDLE_PAGE=PASS'
    Write-Output 'GET_HELPER_LAST_PAGE_OPTIONAL_CURSOR=PASS'
    Write-Output 'GET_HELPER_INVALID_CURSOR_CONTRACT=PASS'

    $savedEnvironment = @{}
    foreach ($key in @('DB_HOST','DB_PORT','DB_USER','DB_PASSWORD','DB_NAME','DB_SSLMODE','DB_MAX_CONNS','JWT_SECRET','JWT_ISSUER','JWT_ACCESS_TTL_MINUTES','JWT_REFRESH_TTL_DAYS','PGPASSWORD','PGSSLMODE','APP_PORT')) {
        $path = "Env:$key"
        $savedEnvironment[$key] = if (Test-Path -LiteralPath $path) { (Get-Item -LiteralPath $path).Value } else { $null }
    }
    try {
        $script:PreviousEnvironment = @{}
        Set-ProcessEnvironment @{ DB_HOST = 'dummy-host'; DB_PORT = '5432'; DB_USER = 'dummy'; DB_PASSWORD = 'dummy'; DB_NAME = 'p7_test_backend'; DB_SSLMODE = 'disable'; JWT_SECRET = 'dummy-secret-with-at-least-thirty-two-chars' }
        $env:PGPASSWORD = 'dummy'
        $env:PGSSLMODE = 'disable'
        Assert-True ($env:DB_NAME -eq 'p7_test_backend' -and $env:APP_PORT -eq '3077') 'Setup environment dummy gagal.'
    } finally {
        Restore-ProcessEnvironment
    }
    foreach ($key in $savedEnvironment.Keys) {
        $path = "Env:$key"
        if ($null -eq $savedEnvironment[$key]) {
            Assert-True (-not (Test-Path -LiteralPath $path)) "Cleanup environment untuk $key gagal."
        } else {
            Assert-True ((Get-Item -LiteralPath $path).Value -ceq $savedEnvironment[$key]) "Cleanup environment untuk $key mengubah nilai awal."
        }
    }
    Write-Output 'ROLE_SELECT_VERIFICATION=PASS (alur memakai SELECT terpisah)'
    Write-Output 'CLEANUP_DRY_RUN=PASS'

    $finallyExecuted = $false
    try {
        throw 'dummy exception untuk menguji finally'
    } catch {
        # Exception dummy memang diharapkan; bukan operasi database.
    } finally {
        $finallyExecuted = $true
    }
    Assert-True $finallyExecuted 'Blok finally tidak berjalan setelah exception dummy.'
    Write-Output 'FINALLY_ON_EXCEPTION=PASS'
}

if ($DryRun -and ($ReadOnlyPreflight -or $ExecuteWriteTests)) {
    throw '-DryRun tidak boleh digabung dengan -ReadOnlyPreflight atau -ExecuteWriteTests.'
}
if ($ReadOnlyPreflight -and $ExecuteWriteTests) {
    throw '-ReadOnlyPreflight tidak boleh digabung dengan -ExecuteWriteTests.'
}
if (-not $DryRun -and -not $ReadOnlyPreflight -and -not $ExecuteWriteTests) {
    Test-LocalPreflight
    Write-Output 'DEFAULT_SAFE_MODE=PASS: tidak membuka koneksi database, tidak menjalankan HTTP, dan tidak melakukan operasi tulis.'
    Write-Output 'Gunakan -ReadOnlyPreflight untuk cek baca DB test, atau -ExecuteWriteTests untuk menjalankan D2/D3 pada database test terpisah.'
    exit 0
}

if ($DryRun) {
    Test-LocalPreflight
    Write-Output 'DRY-RUN PASS: tidak membuka koneksi database atau HTTP.'
    Write-Output 'Alur tervalidasi: preflight -> register -> role SELECT -> login -> D2 -> D3 -> cleanup finally.'
    exit 0
}

try {
    $config = Read-DotEnv $EnvFile
    foreach ($required in @('DB_HOST','DB_PORT','DB_USER','DB_PASSWORD','DB_NAME','DB_SSLMODE','JWT_SECRET')) {
        Assert-True ($config.ContainsKey($required) -and -not [string]::IsNullOrWhiteSpace($config[$required])) ".env tidak memiliki $required yang dapat dipakai"
    }
    Assert-TestDatabaseName $config['DB_NAME']
    $script:DbName = $config['DB_NAME']
    $script:PgHost = $config['DB_HOST']
    $script:PgPort = $config['DB_PORT']
    $script:PgUser = $config['DB_USER']
    $script:PgSslMode = $config['DB_SSLMODE']
    $script:PsqlPath = (Get-Command psql -ErrorAction Stop).Source
    $script:HttpClient = [System.Net.Http.HttpClient]::new()
    $script:HttpClient.Timeout = [TimeSpan]::FromSeconds(20)
    Set-ProcessEnvironment $config
    $env:PGPASSWORD = $config['DB_PASSWORD']
    $env:PGSSLMODE = $script:PgSslMode

    $actualDatabase = Get-DbScalar 'SELECT current_database();'
    Assert-True ($actualDatabase -eq $script:DbName) "current_database()=$actualDatabase; diharapkan $script:DbName"
    Assert-TestDatabaseName $actualDatabase
    $DbReady = $true
    Write-Output "PREFLIGHT current_database=$actualDatabase"

    if ($ReadOnlyPreflight) {
        $readOnlyUsers = Get-DbScalar 'SELECT COUNT(*) FROM users;'
        $readOnlyStudents = Get-DbScalar 'SELECT COUNT(*) FROM students;'
        Write-Output "READ_ONLY users=$readOnlyUsers students=$readOnlyStudents"
        Write-Output 'READ_ONLY_PREFLIGHT=PASS'
    } else {

    $BaselineUserIds = Get-IdList "SELECT COALESCE(string_agg(id::text, ',' ORDER BY id), '') FROM users;"
    $BaselineStudentIds = Get-IdList "SELECT COALESCE(string_agg(id::text, ',' ORDER BY id), '') FROM students;"
    Write-Output "PREFLIGHT users baseline=[$($BaselineUserIds -join ',')] students baseline=[$($BaselineStudentIds -join ',')]"
    if ($D3Only) {
        Compare-IdList -Actual $BaselineUserIds -Expected @(12,13,35,36,37) -Label 'PREFLIGHT users baseline D3'
        Compare-IdList -Actual $BaselineStudentIds -Expected @(1,3,6,32,33,34) -Label 'PREFLIGHT students baseline D3'
        Write-Output 'PREFLIGHT baseline D3=PASS'
    }
    Assert-True (([int](Get-DbScalar "SELECT COUNT(*) FROM users WHERE username = '$TestUsername';")) -eq 0) 'Akun p7_test sudah ada; STOP untuk menghindari menyentuh akun yang bukan dari eksekusi ini.'
    Assert-True (([int](Get-DbScalar "SELECT COUNT(*) FROM students WHERE name LIKE 'P7 D3 %';")) -eq 0) 'Fixture D3 bertanda P7 masih ada; STOP tanpa menghapus data yang tidak dikenal.'
    $fixtureNims = @($D2Nim) + $D3BatchNims + @($D3InsertNim)
    Assert-True (([int](Get-DbScalar "SELECT COUNT(*) FROM students WHERE nim IN ($($fixtureNims -join ','));")) -eq 0) 'NIM fixture eksekusi ini sudah ada; STOP untuk mencegah konflik.'

    Assert-True (([int](Get-DbScalar "SELECT COUNT(*) FROM pg_constraint WHERE conrelid = 'students'::regclass AND conname = 'students_owner_id_fkey';")) -eq 1) 'Foreign key students.owner_id tidak tersedia.'
    Assert-True (([int](Get-DbScalar "SELECT COUNT(*) FROM pg_constraint WHERE conrelid = 'achievements'::regclass AND confrelid = 'students'::regclass;")) -ge 1) 'Foreign key achievements ke students tidak tersedia.'
    Assert-True (([int](Get-DbScalar "SELECT COUNT(*) FROM pg_constraint WHERE conrelid = 'refresh_tokens'::regclass AND confrelid = 'users'::regclass;")) -ge 1) 'Foreign key refresh_tokens ke users tidak tersedia.'
    Write-Output 'PREFLIGHT ownership/achievements/refresh_tokens foreign keys=PASS'

    $staffPermissions = @(Get-DbScalar "SELECT COALESCE(string_agg(permission_name, ',' ORDER BY permission_name), '') FROM role_permissions WHERE role_name = 'staff';").Split(',')
    Assert-True ($staffPermissions -contains 'student:list') 'Role staff tidak memiliki student:list.'
    Assert-True ($staffPermissions -contains 'student:create') 'Role staff tidak memiliki student:create.'
    $staffHasUpdateAny = $staffPermissions -contains 'student:update:any'
    $staffHasDelete = $staffPermissions -contains 'student:delete'
    Write-Output "PREFLIGHT staff list/create=PASS update_any=$staffHasUpdateAny delete=$staffHasDelete; PATCH memakai ownership, cleanup memakai DELETE SQL terikat ID/owner."

    New-Item -ItemType Directory -Force -Path $EvidenceDir | Out-Null
    $serverOut = Join-Path $EvidenceDir "d2_d3_server_$RunTag.stdout.log"
    $serverErr = Join-Path $EvidenceDir "d2_d3_server_$RunTag.stderr.log"
    $Server = Start-Process -FilePath 'go' -ArgumentList 'run','.' -WorkingDirectory $ProjectRoot -RedirectStandardOutput $serverOut -RedirectStandardError $serverErr -PassThru -WindowStyle Hidden
    $ready = $false
    for ($attempt = 1; $attempt -le 120; $attempt++) {
        try {
            $health = Invoke-Api -Method 'GET' -Path '/api/v1/health' -Accept 'application/json'
            if ($health.Status -eq 200) { $ready = $true; break }
        } catch { }
        Start-Sleep -Milliseconds 250
    }
    Assert-True $ready "Server aplikasi tidak siap. Lihat $serverOut dan $serverErr"
    Write-Output 'PREFLIGHT aplikasi asli /api/v1/health HTTP 200'

    $registerBody = @{ username = $TestUsername; email = $TestEmail; password = $TestPassword } | ConvertTo-Json -Compress
    $registered = Invoke-Api -Method 'POST' -Path '/api/v1/auth/register' -JsonBody $registerBody
    Require-Status $registered 201 'AUTH register p7_test'
    $TestUserId = [int]$registered.Json.data.id
    Write-Output "DATA synthetic user_id=$TestUserId username=$TestUsername"

    Invoke-DbCommand "UPDATE users SET role = 'staff' WHERE id = $TestUserId AND username = '$TestUsername' AND email = '$TestEmail';"
    $verifiedRole = Get-DbScalar "SELECT role FROM users WHERE id = $TestUserId AND username = '$TestUsername' AND email = '$TestEmail';"
    Assert-True ($verifiedRole -eq 'staff') "Role akun sintetis bukan staff: $verifiedRole"
    Write-Output 'AUTH role staff SELECT verification=PASS'

    $loginBody = @{ username = $TestUsername; password = $TestPassword } | ConvertTo-Json -Compress
    $login = Invoke-Api -Method 'POST' -Path '/api/v1/auth/login' -JsonBody $loginBody
    Require-Status $login 200 'AUTH login p7_test'
    $token = [string]$login.Json.data.access_token
    Assert-True (-not [string]::IsNullOrWhiteSpace($token)) 'Login tidak mengeluarkan access token.'
    Write-Output 'AUTH JWT diterbitkan=PASS (nilai token tidak dicetak)'

    if (-not $D3Only) {
    $d2Create = Invoke-Api -Method 'POST' -Path '/api/v1/students/' -Token $token -JsonBody (@{ nim = $D2Nim; name = $D2Name; grade = 80 } | ConvertTo-Json -Compress)
    Require-Status $d2Create 201 'D2 create student'
    $D2StudentId = [int]$d2Create.Json.data.id
    Assert-True (([int](Get-DbScalar "SELECT owner_id FROM students WHERE id = $D2StudentId;")) -eq $TestUserId) 'Student D2 tidak dimiliki akun sementara.'
    Write-Output "DATA synthetic d2_student_id=$D2StudentId nim=$D2Nim owner_id=$TestUserId"

    $d2Original = $d2Create.Json.data
    $patchEmptyName = Invoke-Api -Method 'PATCH' -Path "/api/v1/students/$D2StudentId" -Token $token -JsonBody '{"name":""}'
    Require-Status $patchEmptyName 422 'D2 PATCH name kosong'
    Assert-True ($patchEmptyName.Json.code -eq 'VALIDATION_ERROR') 'D2 PATCH name kosong tidak mengembalikan VALIDATION_ERROR.'
    Write-Output "D2 PATCH name kosong response code=$($patchEmptyName.Json.code) fields.name=$($patchEmptyName.Json.fields.name)"

    $patchGrade = Invoke-Api -Method 'PATCH' -Path "/api/v1/students/$D2StudentId" -Token $token -JsonBody '{"grade":90}'
    Require-Status $patchGrade 200 'D2 PATCH grade 90'
    Assert-True ([double]$patchGrade.Json.data.grade -eq 90) 'D2 PATCH grade tidak menjadi 90.'
    Assert-True ($patchGrade.Json.data.name -eq $d2Original.name) 'D2 PATCH grade mengubah name.'
    Assert-True ([int]$patchGrade.Json.data.nim -eq [int]$d2Original.nim) 'D2 PATCH grade mengubah nim.'
    Assert-True ([bool]$patchGrade.Json.data.is_active -eq [bool]$d2Original.is_active) 'D2 PATCH grade mengubah is_active.'
    Assert-True ([int]$patchGrade.Json.data.owner_id -eq $TestUserId) 'D2 PATCH grade mengubah owner_id.'
    Write-Output 'D2 PATCH grade=90 field lain tetap=PASS'

    $patchNoFields = Invoke-Api -Method 'PATCH' -Path "/api/v1/students/$D2StudentId" -Token $token -JsonBody '{}'
    Require-Status $patchNoFields 400 'D2 PATCH kosong'
    Assert-True ($patchNoFields.Json.code -eq 'BAD_REQUEST') 'D2 PATCH kosong tidak mengembalikan BAD_REQUEST.'
    Write-Output "D2 PATCH kosong response code=$($patchNoFields.Json.code)"

    $patchName = Invoke-Api -Method 'PATCH' -Path "/api/v1/students/$D2StudentId" -Token $token -JsonBody '{"name":"P7 Test Updated"}'
    Require-Status $patchName 200 'D2 PATCH name valid'
    Assert-True ($patchName.Json.data.name -eq 'P7 Test Updated') 'D2 PATCH name valid tidak menyimpan nilai yang diminta.'
    Write-Output "D2 PATCH name valid updated_name=$($patchName.Json.data.name)"
    } else {
        Write-Output 'D2_SKIPPED=PASS (mode D3Only)'
    }

    foreach ($index in 0..($D3BatchNims.Count - 1)) {
        $batchNumber = $index + 1
        $batchName = "P7 D3 Batch $batchNumber $RunTag"
        $batchCreate = Invoke-Api -Method 'POST' -Path '/api/v1/students/' -Token $token -JsonBody (@{ nim = $D3BatchNims[$index]; name = $batchName; grade = (81 + $index) } | ConvertTo-Json -Compress)
        Require-Status $batchCreate 201 "D3 create batch $batchNumber"
        $batchStudentId = [int]$batchCreate.Json.data.id
        Assert-True (([int](Get-DbScalar "SELECT owner_id FROM students WHERE id = $batchStudentId;")) -eq $TestUserId) "Student D3 batch $batchNumber tidak dimiliki akun sementara."
        $D3BatchStudentIds.Add($batchStudentId)
        $batchCreatedAt = Get-StudentCreatedAt $batchStudentId
        Write-Output "DATA synthetic d3_batch=$batchNumber student_id=$batchStudentId nim=$($D3BatchNims[$index]) created_at_utc=$batchCreatedAt owner_id=$TestUserId"
    }

    $beforeInsertFirst = Get-StudentPage -Path '/api/v1/students/?limit=2' -Token $token -Label 'D3 halaman pertama sebelum INSERT'
    Assert-True $beforeInsertFirst.HasMore 'D3 membutuhkan next_cursor setelah halaman pertama; data kurang dari tiga record.'
    Assert-True (-not [string]::IsNullOrWhiteSpace($beforeInsertFirst.NextCursor)) 'D3 halaman pertama tidak memberi next_cursor.'
    $oldCursor = $beforeInsertFirst.NextCursor
    $beforeOrder = Get-IdList "SELECT COALESCE(string_agg(id::text, ',' ORDER BY created_at DESC, id DESC), '') FROM students;"
    Assert-True ($beforeOrder.Count -ge 3) 'D3 membutuhkan sedikitnya tiga student untuk cursor lama.'
    $beforePaginationIds = Get-AllStudentPages -Token $token -FirstPath '/api/v1/students/?limit=2' -Label 'D3 snapshot sebelum INSERT'
    Assert-True ((@($beforePaginationIds | Select-Object -Unique).Count -eq $beforePaginationIds.Count)) 'D3 snapshot sebelum INSERT mengembalikan ID duplikat.'
    Compare-IdList -Actual $beforePaginationIds -Expected $beforeOrder -Label 'D3 urutan snapshot sebelum INSERT'
    Write-Output 'D3 snapshot sebelum INSERT duplikasi=PASS urutan=PASS'

    $d3Create = Invoke-Api -Method 'POST' -Path '/api/v1/students/' -Token $token -JsonBody (@{ nim = $D3InsertNim; name = $D3Name; grade = 90 } | ConvertTo-Json -Compress)
    Require-Status $d3Create 201 'D3 INSERT student baru'
    $D3StudentId = [int]$d3Create.Json.data.id
    Assert-True (([int](Get-DbScalar "SELECT owner_id FROM students WHERE id = $D3StudentId;")) -eq $TestUserId) 'Student D3 tidak dimiliki akun sementara.'
    $d3CreatedAt = Get-StudentCreatedAt $D3StudentId
    Write-Output "DATA synthetic d3_insert_student_id=$D3StudentId nim=$D3InsertNim created_at_utc=$d3CreatedAt owner_id=$TestUserId"

    $afterOrder = Get-IdList "SELECT COALESCE(string_agg(id::text, ',' ORDER BY created_at DESC, id DESC), '') FROM students;"
    $cursorIndex = [array]::IndexOf([int[]]$afterOrder, [int]$beforeInsertFirst.IDs[-1])
    Assert-True ($cursorIndex -ge 0) 'Record batas cursor lama tidak ditemukan pada urutan setelah INSERT.'
    $expectedContinuation = if ($cursorIndex + 1 -lt $afterOrder.Count) { @($afterOrder[($cursorIndex + 1)..($afterOrder.Count - 1)]) } else { @() }
    $newIndex = [array]::IndexOf([int[]]$afterOrder, $D3StudentId)
    Assert-True ($newIndex -ge 0) 'Record baru D3 tidak ditemukan dalam urutan database.'
    $newBeforeOldCursor = $newIndex -lt $cursorIndex
    Write-Output "D3 posisi actual new_index=$newIndex cursor_index=$cursorIndex new_before_old_cursor=$newBeforeOldCursor"

    $continuedIds = Get-AllStudentPages -Token $token -FirstPath "/api/v1/students/?limit=2&cursor=$([uri]::EscapeDataString($oldCursor))" -Label 'D3 lanjutan cursor lama'
    Assert-True ((@($continuedIds | Select-Object -Unique).Count -eq $continuedIds.Count)) 'D3 cursor lama mengembalikan ID duplikat.'
    Compare-IdList -Actual $continuedIds -Expected $expectedContinuation -Label 'D3 urutan lanjutan cursor lama'
    if ($newBeforeOldCursor) {
        Assert-True (-not ($continuedIds -contains $D3StudentId)) 'Record baru berada sebelum cursor lama tetapi muncul pada lanjutan cursor.'
        Write-Output 'D3 INSERT_BEFORE_OLD_CURSOR=PASS record_baru_tidak_muncul_di_lanjutan=PASS'
    } else {
        Write-Output 'D3 INSERT_BEFORE_OLD_CURSOR=INCONCLUSIVE (record baru tidak berada sebelum batas cursor lama)'
    }
    Write-Output "D3 cursor lama duplikasi=PASS urutan=PASS record_baru_muncul=$($continuedIds -contains $D3StudentId)"

    $freshIds = Get-AllStudentPages -Token $token -FirstPath '/api/v1/students/?limit=2' -Label 'D3 pagination baru'
    Assert-True ((@($freshIds | Select-Object -Unique).Count -eq $freshIds.Count)) 'D3 pagination baru mengembalikan ID duplikat.'
    Compare-IdList -Actual $freshIds -Expected $afterOrder -Label 'D3 urutan pagination baru'
    Assert-True ($freshIds -contains $D3StudentId) 'D3 pagination baru tidak memuat record baru.'
    Write-Output 'D3 pagination baru duplikasi=PASS urutan=PASS record_baru_terlihat=PASS'
    Write-Output 'D2_D3_EXECUTION=PASS'
    }
} catch {
    $ExecutionProblem = $_
    Write-Output "EXECUTION_FAIL: $($_.Exception.Message)"
} finally {
    if ($null -ne $Server) {
        try {
            if (-not $Server.HasExited) { Stop-Process -Id $Server.Id -ErrorAction Stop }
            foreach ($listener in @(Get-NetTCPConnection -State Listen -LocalPort 3077 -ErrorAction SilentlyContinue)) {
                $listenerProcess = Get-CimInstance Win32_Process -Filter "ProcessId=$($listener.OwningProcess)"
                if ($null -eq $listenerProcess -or $listenerProcess.Name -ne 'pertemuan-7-advanced-api-design.exe') {
                    throw 'Listener port 3077 bukan executable aplikasi uji yang diharapkan.'
                }
                Stop-Process -Id $listener.OwningProcess -ErrorAction Stop
            }
            Start-Sleep -Milliseconds 100
            Assert-True (@(Get-NetTCPConnection -State Listen -LocalPort 3077 -ErrorAction SilentlyContinue).Count -eq 0) 'Listener aplikasi uji masih tersisa pada port 3077.'
            Write-Output 'CLEANUP aplikasi uji dihentikan=PASS'
        } catch {
            $CleanupProblems.Add("server: $($_.Exception.Message)")
        }
    }
    if ($DbReady -and -not $ReadOnlyPreflight) {
        $createdStudentIds = @((@($D2StudentId) + @($D3BatchStudentIds.ToArray()) + @($D3StudentId)) | Where-Object { $null -ne $_ } | Select-Object -Unique)
        foreach ($studentId in $createdStudentIds) {
            try {
                $achievementCount = [int](Get-DbScalar "SELECT COUNT(*) FROM achievements WHERE student_id = $studentId;")
                Write-Output "CLEANUP student_id=$studentId achievements=$achievementCount"
                if ($achievementCount -ne 0) { throw "achievement terkait ditemukan pada student $studentId; student tidak dihapus demi keamanan." }
                Invoke-DbCommand "DELETE FROM students WHERE id = $studentId AND owner_id = $TestUserId;"
                Assert-True (([int](Get-DbScalar "SELECT COUNT(*) FROM students WHERE id = $studentId;")) -eq 0) "student $studentId masih ada setelah DELETE terikat ID/owner."
                Write-Output "CLEANUP student_id=$studentId DELETE verification=PASS"
            } catch {
                $CleanupProblems.Add("student_id=${studentId}: $($_.Exception.Message)")
            }
        }
        if ($null -ne $TestUserId) {
            try {
                $remainingStudents = [int](Get-DbScalar "SELECT COUNT(*) FROM students WHERE owner_id = $TestUserId;")
                if ($remainingStudents -ne 0) { throw "masih ada $remainingStudents student milik akun sementara; user tidak dihapus." }
                $tokenCount = [int](Get-DbScalar "SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $TestUserId;")
                Write-Output "CLEANUP user_id=$TestUserId refresh_tokens=$tokenCount"
                Invoke-DbCommand "DELETE FROM refresh_tokens WHERE user_id = $TestUserId;"
                Assert-True (([int](Get-DbScalar "SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $TestUserId;")) -eq 0) 'refresh token sementara masih ada.'
                Invoke-DbCommand "DELETE FROM users WHERE id = $TestUserId AND username = '$TestUsername' AND email = '$TestEmail';"
                Assert-True (([int](Get-DbScalar "SELECT COUNT(*) FROM users WHERE id = $TestUserId;")) -eq 0) "user $TestUserId masih ada setelah DELETE terikat ID/email."
                Write-Output "CLEANUP user_id=$TestUserId DELETE verification=PASS"
            } catch {
                $CleanupProblems.Add("user_id=${TestUserId}: $($_.Exception.Message)")
            }
        }
        try {
            $finalUserIds = Get-IdList "SELECT COALESCE(string_agg(id::text, ',' ORDER BY id), '') FROM users;"
            $finalStudentIds = Get-IdList "SELECT COALESCE(string_agg(id::text, ',' ORDER BY id), '') FROM students;"
            Compare-IdList -Actual $finalUserIds -Expected $BaselineUserIds -Label 'CLEANUP users baseline'
            Compare-IdList -Actual $finalStudentIds -Expected $BaselineStudentIds -Label 'CLEANUP students baseline'
            Assert-True (([int](Get-DbScalar "SELECT COUNT(*) FROM users WHERE username = '$TestUsername';")) -eq 0) 'Akun p7_test masih tersisa.'
            Write-Output "CLEANUP final users=[$($finalUserIds -join ',')] students=[$($finalStudentIds -join ',')] p7_test=0"
        } catch {
            $CleanupProblems.Add("baseline verification: $($_.Exception.Message)")
        }
    }
    if ($null -ne $script:HttpClient) { $script:HttpClient.Dispose() }
    Restore-ProcessEnvironment
}

if ($CleanupProblems.Count -gt 0) {
    Write-Output "CLEANUP_FAIL: $($CleanupProblems -join ' | ')"
    throw 'Cleanup gagal; lihat ID sintetis dan detail di atas.'
}
if ($null -ne $ExecutionProblem) { throw $ExecutionProblem }
