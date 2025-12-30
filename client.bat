@echo off
echo Distributed File System Client
echo.

if "%1"=="" goto usage
if "%1"=="upload" goto upload
if "%1"=="download" goto download
if "%1"=="list" goto list
goto usage

:upload
if "%2"=="" goto usage
if "%3"=="" goto usage
go run cmd/client/main.go upload -file %2 -name %3
goto end

:download
if "%2"=="" goto usage
if "%3"=="" goto usage
go run cmd/client/main.go download -name %2 -output %3
goto end

:list
go run cmd/client/main.go list
goto end

:usage
echo Usage:
echo   client.bat upload ^<local_file^> ^<remote_name^>
echo   client.bat download ^<remote_name^> ^<local_file^>
echo   client.bat list
echo.
echo Examples:
echo   client.bat upload testfile.txt myfile.txt
echo   client.bat download myfile.txt downloaded.txt
echo   client.bat list
goto end

:end
