!define REGKEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\Door2docUploadService"
!define PROGRAM_NAME "Upload Service"

Name "Door2doc Upload Service"
OutFile "_installer.exe"

InstallDir "$PROGRAMFILES64\Door2doc\UploadService"

RequestExecutionLevel admin

Function .onInit

  ReadRegStr $R0 HKLM "${REGKEY}" "UninstallString"
  StrCmp $R0 "" done

  MessageBox MB_OKCANCEL|MB_ICONEXCLAMATION \
  "${PROGRAM_NAME} is already installed. $\n$\nClick `OK` to remove the \
  previous version or `Cancel` to cancel this upgrade." \
  IDOK uninst
  Abort

;Run the uninstaller
uninst:
  ClearErrors
  ExecWait '$R0 _?=$INSTDIR' ;Do not copy the uninstaller to a temp file

  IfErrors no_remove_uninstaller done
    ;You can either use Delete /REBOOTOK in the uninstaller or add some code
    ;here to remove the uninstaller. Use a registry key to check
    ;whether the user has chosen to uninstall. If you are using an uninstaller
    ;components page, make sure all sections are uninstalled.
  no_remove_uninstaller:

done:
FunctionEnd


Section
    # Stop and remove service if it already exists
    ExecWait '"$INSTDIR\UploadService.exe" stop'
    ExecWait '"$INSTDIR\UploadService.exe" uninstall'

    SetOutPath $INSTDIR

    File /oname=UploadService.exe  d2d-upload_windows_amd64.exe

    WriteUninstaller "$INSTDIR\uninstall.exe"

    WriteRegStr HKLM "${REGKEY}" "DisplayName" "Door2doc Upload Service"
    WriteRegStr HKLM "${REGKEY}" "Publisher" "Door2doc BV"
    WriteRegStr HKLM "${REGKEY}" "UninstallString" "$\"$INSTDIR\uninstall.exe$\""
    WriteRegStr HKLM "${REGKEY}" "QuietUninstallString" "$\"$INSTDIR\uninstall.exe$\" /S"

    ExecWait '"$INSTDIR\UploadService.exe" install'
    ExecWait '"$INSTDIR\UploadService.exe" start'

SectionEnd

Section "uninstall"
    ExecWait '"$INSTDIR\UploadService.exe" stop'
    ExecWait '"$INSTDIR\UploadService.exe" uninstall'

    DeleteRegKey HKLM "${REGKEY}"
    Delete "$INSTDIR\UploadService.exe"
    Delete "$INSTDIR\uninstall.exe"
    RMDir /r "$INSTDIR"
    RMDir "$PROGRAMFILES64\Door2doc"
SectionEnd