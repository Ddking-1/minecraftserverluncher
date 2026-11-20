@echo off
title Alpha Test v2 Selector
cd /d "%~dp0"
echo ============================================================
echo  Alpha Test v2 - Minecraft Launcher ecosystem
echo ============================================================
echo.
echo  Choose what to test:
echo    1 - Web (backend + admin/user web UI)
echo    2 - Launcher (backend + MC server max 2GB + Velocity + launcher)
echo    0 - Exit
echo.
set /p CHOICE="Enter choice (1/2/0): "
if "%CHOICE%"=="1" goto WEB
if "%CHOICE%"=="2" goto LAUNCHER
if "%CHOICE%"=="0" goto END
echo Invalid choice.
goto END

:WEB
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0run_web.ps1"
goto END

:LAUNCHER
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0run_launcher.ps1"
goto END

:END
echo.
echo Done.
pause