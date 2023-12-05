@echo off

REM Configuration
set SERVICE_NAME=shugosha
set BINARY_NAME=shugosha
set BUILD_DIR=.\build
set GO_CMD=go

REM Check if a command line argument was provided
if "%1"=="" goto help

if "%1" == "init" (
    Import-Module posh-git
    goto end
)

REM Run the project
if "%1"=="run" (
    cd cmd/shugosha
    go run main.go wire_gen.go
    goto end
)

REM Build the project
if "%1"=="build" (
    %GO_CMD% build -o %BUILD_DIR%\%BINARY_NAME% .\cmd\%SERVICE_NAME%\main.go
    goto end
)

REM Run tests
if "%1"=="test" (
    %GO_CMD% test .\...
    goto end
)

REM Clean build artifacts
if "%1"=="clean" (
    %GO_CMD% clean
    if exist %BUILD_DIR% rmdir /s /q %BUILD_DIR%
    goto end
)

REM Format the code
if "%1"=="fmt" (
    %GO_CMD% fmt .\...
    goto end
)

REM Vet the code
if "%1"=="vet" (
    %GO_CMD% vet .\...
    goto end
)

REM Display help
:help
echo Available commands:
echo   build  - Build the project binary
echo   test   - Run tests
echo   clean  - Clean build artifacts
echo   fmt    - Format the code
echo   vet    - Vet the code
echo   run    - Run the project
echo   init   - Init power shell modules
goto end

:end
