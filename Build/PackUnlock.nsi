# Name of the installer
Outfile "YST_Unlock.exe"

# Directory to install to (default is Program Files)
InstallDir $PROGRAMFILES\YST_Unlock

# Name of the application
Name "YST_Unlock"
# Version of the application
InstallDirRegKey HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\YST_Unlock" "Install_Dir"

# App details
VIProductVersion "1.1"
VIAddVersionKey /LANG=1033 ProductName "YST_Unlock"
VIAddVersionKey /LANG=1033 ProductVersion "1.1"
VIAddVersionKey /LANG=1033 Publisher "JohnRey"

# Compression settings
SetCompressor lzma
# SolidCompression removed, as it's not valid in NSIS

# Language settings
LangString DESC_LANG_1033 ${LANG_ENGLISH} "English"
LangString DESC_LANG_2052 ${LANG_CHINESE_SIMPLIFIED} "简体中文"

# Sections (files, registry, shortcuts, etc.)
Section "Install Files"

  # File Install: UnlockAll.exe
  SetOutPath $INSTDIR
  File "E:\MyProject\YiSaiTongUnLock\UnlockAll\UnlockAll.exe"

  # File Install: wps.exe
  File "E:\MyProject\YiSaiTongUnLock\UnlockAll\wps.exe"

SectionEnd

# Create registry keys for "Unlock" context menu integration
Section "Registry Entries"

  # Unlock context menu for directories
  WriteRegStr HKCR "Directory\shell\Unlock" "" "Unlock"
  WriteRegStr HKCR "Directory\shell\Unlock\command" "" '"$INSTDIR\UnlockAll.exe" "$1"'

  # Unlock context menu for files
  WriteRegStr HKCR "*\shell\Unlock" "" "Unlock"
  WriteRegStr HKCR "*\shell\Unlock\command" "" '"$INSTDIR\UnlockAll.exe" "$1"'

SectionEnd

# Uninstaller Section
Section "Uninstall"

  # Remove installed files
  Delete "$INSTDIR\UnlockAll.exe"
  Delete "$INSTDIR\wps.exe"

  # Remove registry keys
  DeleteRegKey HKCR "Directory\shell\Unlock"
  DeleteRegKey HKCR "Directory\shell\Unlock\command"
  DeleteRegKey HKCR "*\shell\Unlock"
  DeleteRegKey HKCR "*\shell\Unlock\command"

  # Remove installation directory (if empty)
  RMDir $INSTDIR

SectionEnd
