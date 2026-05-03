eval "$(/opt/homebrew/bin/brew shellenv)"

export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"  # This loads nvm
[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"  # This loads nvm bash_completion
export PATH=$PATH:$(go env GOPATH)/bin
export PATH="/opt/homebrew/opt/postgresql@15/bin:$PATH"
export PATH="$PATH:/Users/a.nasstya/flutter/bin"
# The following lines have been added by Docker Desktop to enable Docker CLI completions.
fpath=(/Users/a.nasstya/.docker/completions $fpath)
autoload -Uz compinit
compinit
# End of Docker CLI completions
# Flutter
export PATH="$PATH:/Users/a.nasstya/flutter/bin"

# Android SDK
export ANDROID_HOME="/Users/a.nasstya/Library/Android/sdk"
export PATH="$PATH:$ANDROID_HOME/emulator"
export PATH="$PATH:$ANDROID_HOME/platform-tools"
export PATH="$PATH:$ANDROID_HOME/tools"
export PATH="$PATH:$ANDROID_HOME/tools/bin"
