@echo off
setlocal enabledelayedexpansion

:: Default number of clients
set CLIENTS=10

:: Check if an arguments is provided for amount of clients
if not "%1"=="" set CLIENTS=%1

:: Compile
echo Compiling Go program...
go build -o sketcher.exe .
if %ERRORLEVEL% neq 0 (
  echo Compilation failed!
  exit /b
)

:: Run the server
echo Starting Server...
start "SketchServer" sketcher.exe

timeout /t 4 /nobreak >nul

start "SketcherClient" sketcher.exe -consumer

:: Start clients
echo Starting %CLIENTS% headless clients...
for /L %%i in (1,1,%CLIENTS%) do (
  start /b sketcher.exe -client -d "./data/PVS 1/dataset_gps.csv" -name "speed_meters_per_second" -type float
)


:: Stop wating for command to stop clients
echo press any key to kill the clients!
pause >nul
echo killing all the clients!


:: Kill all clients
taskkill /IM sketcher.exe /F >nul 2>&1

echo The clients have been killed
exit
