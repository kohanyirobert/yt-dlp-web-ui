# Persist Bash history across rebuilds
sudo chown -R $(whoami):$(whoami) /cmdhistory
touch /cmdhistory/.bash_history
ln -sf /cmdhistory/.bash_history ~/.bash_history
echo 'export PROMPT_COMMAND="history -a;${PROMPT_COMMAND}"' >> ~/.bashrc

# Persist GitHub CLI auth across rebuilds
sudo chown -R $(whoami):$(whoami) /home/vscode/.config/gh

# Persist Claude Code settings and memory across rebuilds
sudo chown -R $(whoami):$(whoami) /claude-persist
mkdir -p /claude-persist/.claude
ln -sf /claude-persist/.claude ~/.claude

# Restore .claude.json settings (e.g. theme) if missing
if [ ! -f ~/.claude.json ]; then
    BACKUP_DIR="/claude-persist/.claude/backups"
    if [ -d "$BACKUP_DIR" ]; then
        NEWEST=$(ls -t "$BACKUP_DIR"/.claude.json.backup.* 2>/dev/null | head -1)
        if [ -n "$NEWEST" ]; then
            cp "$NEWEST" ~/.claude.json
        fi
    fi
fi

# Create a stable wrapper for the VS Code remote CLI so EDITOR="code --wait"
# resolves to the host VS Code window (the commit-hash path changes on updates).
REMOTE_CODE=$(ls /home/vscode/.vscode-server/bin/*/bin/remote-cli/code 2>/dev/null | head -1)
if [ -n "$REMOTE_CODE" ]; then
    sudo ln -sf "$REMOTE_CODE" /usr/local/bin/code
fi