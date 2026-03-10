# Persist Bash history across rebuilds
sudo chown -R $(whoami):$(whoami) /cmdhistory
touch /cmdhistory/.bash_history
ln -sf /cmdhistory/.bash_history ~/.bash_history
echo 'export PROMPT_COMMAND="history -a;${PROMPT_COMMAND}"' >> ~/.bashrc

# Persist Claude Code settings and memory across rebuilds
sudo chown -R $(whoami):$(whoami) /claude-persist
mkdir -p /claude-persist/.claude
ln -sf /claude-persist/.claude ~/.claude