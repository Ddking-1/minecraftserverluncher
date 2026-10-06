# SPDX-License-Identifier: GPL-3.0-only
# NSIS installer for the MinecraftServerLauncher.
#
# Build: place the built MinecraftLauncher.exe and the Qt runtime DLLs
# (Qt6Core.dll, Qt6Network.dll, Qt6Widgets.dll) and the qwindows platform
# plugin alongside this script, then run makensis on it.
Unicode true
Name "MinecraftServerLauncher"
OutFile "MinecraftServerLauncher-0.1.0-setup.exe"
InstallDir "$PROGRAMFILES64\MinecraftServerLauncher"

RequestExecutionLevel admin

Page license
Page directory
Page instfiles

Var MYPATH
Var APPNAME
Var VERSION

Function .onInit
    StrCpy $MYPATH "$EXEDIR\"
    StrCpy $APPNAME "MinecraftServerLauncher"
    StrCpy $VERSION "0.1.0"
FunctionEnd

Section "Main"
    SetOutPath "$INSTDIR"

    File "MinecraftLauncher.exe"
    File "Qt6Core.dll"
    File "Qt6Network.dll"
    File "Qt6Widgets.dll"
    SetOutPath "$INSTDIR\platforms"
    File "platforms\qwindows.dll"
    SetOutPath "$INSTDIR"

    WriteUninstaller "$INSTDIR\uninstall.exe"

    CreateDirectory "$SMPROGRAMS\$APPNAME"
    CreateShortcut "$SMPROGRAMS\$APPNAME\MinecraftLauncher.lnk" "$INSTDIR\MinecraftLauncher.exe"

    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\MinecraftServerLauncher" "DisplayName" $APPNAME
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\MinecraftServerLauncher" "DisplayVersion" $VERSION
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\MinecraftServerLauncher" "Publisher" "Your Publisher"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\MinecraftServerLauncher" "UninstallString" '"$INSTDIR\uninstall.exe"'
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\MinecraftServerLauncher" "InstallLocation" "$INSTDIR"
SectionEnd

Section "Uninstall"
    Delete "$INSTDIR\MinecraftLauncher.exe"
    RMDir /r "$INSTDIR\platforms"
    RMDir "$INSTDIR"

    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\MinecraftServerLauncher"

    RMDir "$SMPROGRAMS\$APPNAME"
SectionEnd
