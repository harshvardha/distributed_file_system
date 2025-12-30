@echo off
echo Starting Distributed File System...
echo.

echo Starting Master Server...
start "DFS Master" powershell -NoExit -Command "cd 'e:\golang_projects\dfs'; go run cmd/master/main.go"
timeout /t 2 /nobreak >nul

echo Starting Chunk Server 1...
start "DFS ChunkServer 1" powershell -NoExit -Command "cd 'e:\golang_projects\dfs'; go run cmd/chunkserver/main.go -port 9001 -storage ./storage1"
timeout /t 1 /nobreak >nul

echo Starting Chunk Server 2...
start "DFS ChunkServer 2" powershell -NoExit -Command "cd 'e:\golang_projects\dfs'; go run cmd/chunkserver/main.go -port 9002 -storage ./storage2"
timeout /t 1 /nobreak >nul

echo Starting Chunk Server 3...
start "DFS ChunkServer 3" powershell -NoExit -Command "cd 'e:\golang_projects\dfs'; go run cmd/chunkserver/main.go -port 9003 -storage ./storage3"

echo.
echo All servers started!
echo Wait 10 seconds for servers to initialize before running client commands.
pause
