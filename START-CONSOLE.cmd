@echo off
cd /d "%~dp0operator"
go run ./cmd/kyrelis-operator
if errorlevel 1 pause
