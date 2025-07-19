; 脚本由 Inno Setup 脚本向导 生成！
; 有关创建 Inno Setup 脚本文件的详细资料请查阅帮助文档！

#define MyAppName "ESafeTestUnlock"
#define MyAppVersion "1.1"
#define MyAppPublisher "Evek"
#define MyAppIcon "unlock.ico"  ; 图标文件名

[Setup]
; 注: AppId的值为单独标识该应用程序。
; 不要为其他安装程序使用相同的AppId值。
; (生成新的GUID，点击 工具|在IDE中生成GUID。)
AppId={{1167EBF7-917B-40D5-977F-CDFE9D9E3AA1}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
;AppVerName={#MyAppName} {#MyAppVersion}
AppPublisher={#MyAppPublisher}
DefaultDirName={commonpf}\{#MyAppName}
DefaultGroupName={#MyAppName}
OutputBaseFilename={#MyAppName}
Compression=lzma
SolidCompression=yes

[Languages]
Name: "chinesesimp"; MessagesFile: "compiler:Default.isl"

[Files]
Source: "D:\Work\gitee\GoProj\YiSaiTongUnlock\UnlockAll\BeeUnlockOffice.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "D:\Work\gitee\GoProj\YiSaiTongUnlock\UnlockFile\wps.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "D:\Work\gitee\GoProj\YiSaiTongUnlock\Build\{#MyAppIcon}"; DestDir: "{app}"; Flags: ignoreversion

; 注意: 不要在任何共享系统文件上使用“Flags: ignoreversion”

[Registry]
; For folder right-click menu (Unlock folder)
Root: HKCR; Subkey: "Directory\shell\Unlock"; ValueType: string; ValueName: ""; ValueData: "Unlock"; Flags: uninsdeletekey
Root: HKCR; Subkey: "Directory\shell\Unlock\command"; ValueType: string; ValueName: ""; ValueData: """{app}\UnlockAll.exe"" ""%1"""; Flags: uninsdeletekey
Root: HKCR; Subkey: "Directory\shell\Unlock"; ValueType: string; ValueName: "Icon"; ValueData: """{app}\{#MyAppIcon}"""; Flags: uninsdeletekey

; For file right-click menu (Unlock file)
Root: HKCR; Subkey: "*\shell\Unlock"; ValueType: string; ValueName: ""; ValueData: "Unlock"; Flags: uninsdeletekey
Root: HKCR; Subkey: "*\shell\Unlock\command"; ValueType: string; ValueName: ""; ValueData: """{app}\UnlockAll.exe"" ""%1"""; Flags: uninsdeletekey
Root: HKCR; Subkey: "*\shell\Unlock"; ValueType: string; ValueName: "Icon"; ValueData: """{app}\{#MyAppIcon}"""; Flags: uninsdeletekey
