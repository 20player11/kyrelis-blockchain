@echo off
cd /d "%~dp0"
py -3 scripts\devnet.py start %*
if errorlevel 1 pause
