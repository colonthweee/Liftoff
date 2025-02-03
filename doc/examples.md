# Liftoff Configuration Examples

## Developer Workstation Setup
```yaml
packages:
  chocolatey:
    - git
    - vscode
    - nodejs
    - python
    - docker-desktop

system:
  dark_mode: true
  folders:
    - "${USERPROFILE}/Documents/Projects"
    - "${USERPROFILE}/Documents/Workspace"
    - "${USERPROFILE}/.config"

git:
  repositories:
    - url: "https://github.com/microsoft/terminal"
      path: "${USERPROFILE}/Documents/Projects/terminal"
      branch: "main"
      depth: 1

environment:
  variables:
    JAVA_HOME: "C:/Program Files/Java/jdk-17"
  path_append:
    - "%JAVA_HOME%/bin"
    - "${USERPROFILE}/.local/bin"

file_associations:
  associations:
    .py: "${USERPROFILE}/AppData/Local/Programs/Python/Python39/python.exe"
    .js: "${USERPROFILE}/AppData/Local/Programs/Microsoft VS Code/Code.exe"
