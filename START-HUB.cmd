@echo off
cd /d "%~dp0"
py -3 scripts\hub.py
if errorlevel 1 pause
