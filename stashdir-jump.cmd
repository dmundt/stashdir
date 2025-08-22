@echo off
REM stashdir-jump.cmd - change current cmd.exe directory using stashdir
REM Usage examples:
REM   stashdir-jump               - interactive select and cd
REM   stashdir-jump list          - show saved paths
REM   stashdir-jump add [PATH]    - add path (defaults to CWD)
REM   stashdir-jump remove <ARG>  - remove by index or path

set EXE=%~dp0stashdir.exe
if not exist "%EXE%" (
  set EXE=stashdir.exe
)

if "%1"=="list" (
  %EXE% list
  goto :eof
)
if "%1"=="add" (
  if "%2"=="" (
    %EXE% add .
  ) else (
    %EXE% add %2
  )
  goto :eof
)
if "%1"=="remove" (
  if "%2"=="" (
    echo Usage: stashdir-jump remove ^<index|path^>
    goto :eof
  )
  %EXE% remove %2
  goto :eof
)

for /f "usebackq delims=" %%i in (`%EXE% select`) do cd /d "%%i"
