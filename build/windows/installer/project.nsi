Unicode true

####
## Please note: Template replacements don't work in this file. They are provided with default defines like
## mentioned underneath.
## If the keyword is not defined, "wails_tools.nsh" will populate them with the values from ProjectInfo.
## If they are defined here, "wails_tools.nsh" will not touch them. This allows to use this project.nsi manually
## from outside of Wails for debugging and development of the installer.
##
## For development first make a wails nsis build to populate the "wails_tools.nsh":
## > wails build --target windows/amd64 --nsis
## Then you can call makensis on this file with specifying the path to your binary:
## For a AMD64 only installer:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app.exe
## For a ARM64 only installer:
## > makensis -DARG_WAILS_ARM64_BINARY=..\..\bin\app.exe
## For a installer with both architectures:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app-amd64.exe -DARG_WAILS_ARM64_BINARY=..\..\bin\app-arm64.exe
####
## The following information is taken from the ProjectInfo file, but they can be overwritten here.
!define INFO_PROJECTNAME    "EasyWinGet"
!define INFO_COMPANYNAME    "DARKBOYJYOTI"
!define INFO_PRODUCTNAME    "EasyWinGet"
!define INFO_PRODUCTVERSION "4.0.0"
!define INFO_COPYRIGHT      "Copyright © 2026 DARKBOYJYOTI"

!define UNINST_KEY_NAME "EasyWinGet"
BrandingText "EasyWinGet by DARKBOYJYOTI"

## Include the wails tools
!include "wails_tools.nsh"

# The version information for this two must consist of 4 parts
VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "DARKBOYJYOTI"
VIAddVersionKey "FileDescription" "EasyWinGet Setup Installer by DARKBOYJYOTI"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "Copyright © 2026 DARKBOYJYOTI"
VIAddVersionKey "ProductName"     "EasyWinGet"
VIAddVersionKey "Comments"        "Windows Package Manager UI by DARKBOYJYOTI"

# Enable HiDPI support. https://nsis.sourceforge.io/Reference/ManifestDPIAware
ManifestDPIAware true

!include "MUI.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
# !define MUI_WELCOMEFINISHPAGE_BITMAP "resources\leftimage.bmp" #Include this to add a bitmap on the left side of the Welcome Page. Must be a size of 164x314
!define MUI_FINISHPAGE_NOAUTOCLOSE # Wait on the INSTFILES page so the user can take a look into the details of the installation steps
!define MUI_ABORTWARNING # This will warn the user if they exit from the installer.

!insertmacro MUI_PAGE_WELCOME # Welcome to the installer page.
# !insertmacro MUI_PAGE_LICENSE "resources\eula.txt" # Adds a EULA page to the installer
!insertmacro MUI_PAGE_DIRECTORY # In which folder install page.
!insertmacro MUI_PAGE_INSTFILES # Installing page.

# Finish Page Checkboxes (Both checked by default)
!define MUI_FINISHPAGE_RUN ""
!define MUI_FINISHPAGE_RUN_TEXT "Launch ${INFO_PRODUCTNAME}"
!define MUI_FINISHPAGE_RUN_FUNCTION LaunchAppFinish
!define MUI_FINISHPAGE_RUN_CHECKED

!define MUI_FINISHPAGE_SHOWREADME ""
!define MUI_FINISHPAGE_SHOWREADME_TEXT "Create Desktop Shortcut"
!define MUI_FINISHPAGE_SHOWREADME_CHECKED
!define MUI_FINISHPAGE_SHOWREADME_FUNCTION CreateDesktopShortcutFinish

!insertmacro MUI_PAGE_FINISH # Finished installation page.

!insertmacro MUI_UNPAGE_INSTFILES # Uinstalling page

!insertmacro MUI_LANGUAGE "English" # Set the Language of the installer

## The following two statements can be used to sign the installer and the uninstaller. The path to the binaries are provided in %1
#!uninstfinalize 'signtool --file "%1"'
#!finalize 'signtool --file "%1"'

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe" # Name of the installer's file.
!ifdef WAILS_INSTALL_SCOPE
  !if "${WAILS_INSTALL_SCOPE}" == "user"
    InstallDir "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
  !else
    InstallDir "$PROGRAMFILES64\${INFO_PRODUCTNAME}"
  !endif
!else
  InstallDir "$PROGRAMFILES64\${INFO_PRODUCTNAME}"
!endif # Default installing folder ($PROGRAMFILES is Program Files folder).
ShowInstDetails show # This will always show the installation details.

Function .onInit
   !insertmacro wails.checkArchitecture
FunctionEnd

Function LaunchAppFinish
    # Launch as the standard interactive logged-in user (prevents high-integrity UIPI block)
    Exec '"$WINDIR\explorer.exe" "$INSTDIR\${PRODUCT_EXECUTABLE}"'
FunctionEnd

Function CreateDesktopShortcutFinish
    !insertmacro wails.setShellContext
    SetOutPath $INSTDIR
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}" "" "$INSTDIR\${PRODUCT_EXECUTABLE}" 0
    ${If} ${REQUEST_EXECUTION_LEVEL} == "admin"
        SetShellVarContext current
        CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}" "" "$INSTDIR\${PRODUCT_EXECUTABLE}" 0
        SetShellVarContext all
    ${EndIf}
    System::Call 'shell32::SHChangeNotify(i 0x08000000, i 0, i 0, i 0)'
FunctionEnd

Section
    !insertmacro wails.setShellContext

    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR

    !insertmacro wails.files
    File "Pin-Taskbar.ps1"

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}" "" "$INSTDIR\${PRODUCT_EXECUTABLE}" 0
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}" "" "$INSTDIR\${PRODUCT_EXECUTABLE}" 0
    ${If} ${REQUEST_EXECUTION_LEVEL} == "admin"
        SetShellVarContext current
        CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}" "" "$INSTDIR\${PRODUCT_EXECUTABLE}" 0
        SetShellVarContext all
    ${EndIf}
    System::Call 'shell32::SHChangeNotify(i 0x08000000, i 0, i 0, i 0)'

    # Pin to Windows 10/11 Taskbar by default
    DetailPrint "Pinning ${INFO_PRODUCTNAME} to Taskbar..."
    nsExec::Exec 'powershell.exe -NoProfile -ExecutionPolicy Bypass -WindowStyle Hidden -File "$INSTDIR\Pin-Taskbar.ps1" -Pin "$INSTDIR\${PRODUCT_EXECUTABLE}" -Silent'

    # Remove any old stray directories from previous versions
    RMDir /r "$INSTDIR\app"
    RMDir /r "$INSTDIR\data"
    RMDir /r "$INSTDIR\jobs"

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols

    !insertmacro wails.writeUninstaller
SectionEnd

Section "uninstall"
    !insertmacro wails.setShellContext

    # Unpin from Taskbar before removing files
    DetailPrint "Unpinning ${INFO_PRODUCTNAME} from Taskbar..."
    ${If} ${FileExists} "$INSTDIR\Pin-Taskbar.ps1"
        nsExec::Exec 'powershell.exe -NoProfile -ExecutionPolicy Bypass -WindowStyle Hidden -File "$INSTDIR\Pin-Taskbar.ps1" -Unpin "${INFO_PRODUCTNAME}*" -Silent'
        Delete "$INSTDIR\Pin-Taskbar.ps1"
    ${EndIf}

    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}" # Remove the WebView2 DataPath
    RMDir /r "$LOCALAPPDATA\${INFO_PRODUCTNAME}" # Remove application data and logs

    RMDir /r $INSTDIR

    SetShellVarContext all
    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    SetShellVarContext current
    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    System::Call 'shell32::SHChangeNotify(i 0x08000000, i 0, i 0, i 0)'

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    !insertmacro wails.deleteUninstaller
SectionEnd
