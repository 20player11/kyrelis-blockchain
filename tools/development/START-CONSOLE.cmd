@echo off
cd /d "%~dp0..\..\operator"
go run ./cmd/kyrelis-operator
if errorlevel 1 pause
