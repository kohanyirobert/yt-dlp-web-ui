# Persist Bash history across rebuilds
sudo chown -R $(whoami):$(whoami) /cmdhistory
touch /cmdhistory/.bash_history
ln -sf /cmdhistory/.bash_history ~/.bash_history
echo 'export PROMPT_COMMAND="history -a;${PROMPT_COMMAND}"' >> ~/.bashrc

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