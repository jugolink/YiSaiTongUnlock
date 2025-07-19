; �ű��� Inno Setup �ű��� ���ɣ�
; �йش��� Inno Setup �ű��ļ�����ϸ��������İ����ĵ���

#define MyAppName "ESafeTestUnlock"
#define MyAppVersion "1.1"
#define MyAppPublisher "Evek"
#define MyAppIcon "unlock.ico"  ; ͼ���ļ���

[Setup]
; ע: AppId��ֵΪ������ʶ��Ӧ�ó���
; ��ҪΪ������װ����ʹ����ͬ��AppIdֵ��
; (�����µ�GUID����� ����|��IDE������GUID��)
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

; ע��: ��Ҫ���κι���ϵͳ�ļ���ʹ�á�Flags: ignoreversion��

[Registry]
; For folder right-click menu (Unlock folder)
Root: HKCR; Subkey: "Directory\shell\Unlock"; ValueType: string; ValueName: ""; ValueData: "Unlock"; Flags: uninsdeletekey
Root: HKCR; Subkey: "Directory\shell\Unlock\command"; ValueType: string; ValueName: ""; ValueData: """{app}\UnlockAll.exe"" ""%1"""; Flags: uninsdeletekey
Root: HKCR; Subkey: "Directory\shell\Unlock"; ValueType: string; ValueName: "Icon"; ValueData: """{app}\{#MyAppIcon}"""; Flags: uninsdeletekey

; For file right-click menu (Unlock file)
Root: HKCR; Subkey: "*\shell\Unlock"; ValueType: string; ValueName: ""; ValueData: "Unlock"; Flags: uninsdeletekey
Root: HKCR; Subkey: "*\shell\Unlock\command"; ValueType: string; ValueName: ""; ValueData: """{app}\UnlockAll.exe"" ""%1"""; Flags: uninsdeletekey
Root: HKCR; Subkey: "*\shell\Unlock"; ValueType: string; ValueName: "Icon"; ValueData: """{app}\{#MyAppIcon}"""; Flags: uninsdeletekey
