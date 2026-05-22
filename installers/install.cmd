@echo off
setlocal

set "BASE_URL=https://neeto-downloads.s3.amazonaws.com/cli/NeetoRecord/latest"
set "INSTALL_DIR=%LOCALAPPDATA%\Programs\neetorecord"

set "ARCH=amd64"
if "%PROCESSOR_ARCHITECTURE%"=="ARM64" set "ARCH=arm64"
set "ARCHIVE=neetorecord_windows_%ARCH%.zip"

echo Downloading NeetoRecord CLI...
set "TMPDIR=%TEMP%\neetorecord-install"
if exist "%TMPDIR%" rmdir /s /q "%TMPDIR%"
mkdir "%TMPDIR%"

curl -fsSL "%BASE_URL%/%ARCHIVE%" -o "%TMPDIR%\%ARCHIVE%"
if %errorlevel% neq 0 (
    echo Failed to download NeetoRecord CLI.
    exit /b 1
)

echo Extracting...
powershell -Command "Expand-Archive -Path '%TMPDIR%\%ARCHIVE%' -DestinationPath '%TMPDIR%' -Force"

echo Installing to %INSTALL_DIR%...
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"
copy /y "%TMPDIR%\neetorecord.exe" "%INSTALL_DIR%\neetorecord.exe" >nul

:: Add to user PATH if not already present
echo %PATH% | findstr /i /c:"%INSTALL_DIR%" >nul
if %errorlevel% neq 0 (
    for /f "tokens=2*" %%A in ('reg query "HKCU\Environment" /v Path 2^>nul') do set "USER_PATH=%%B"
    setx PATH "%USER_PATH%;%INSTALL_DIR%" >nul
    echo Added %INSTALL_DIR% to user PATH.
)

rmdir /s /q "%TMPDIR%"

echo NeetoRecord CLI installed successfully. Restart your terminal and run 'neetorecord --help' to get started.
