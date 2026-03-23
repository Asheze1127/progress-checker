FROM mcr.microsoft.com/devcontainers/base:bookworm

ENV PATH=/home/vscode/.local/bin:/home/vscode/.asdf/bin:/home/vscode/.asdf/shims:/go/bin:${PATH}

RUN apt-get update \
    && export DEBIAN_FRONTEND=noninteractive \
    && apt-get install -y --no-install-recommends ca-certificates curl dirmngr git gnupg postgresql-client unzip xz-utils \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /commandhistory /go/bin /home/vscode/.asdf /home/vscode/.local/bin /workspaces/progress-checker/node_modules \
    && touch /commandhistory/.bash_history /commandhistory/.zsh_history \
    && chown -R vscode:vscode /commandhistory /go /home/vscode/.asdf /home/vscode/.local /workspaces \
    && chown vscode:vscode /commandhistory/.bash_history /commandhistory/.zsh_history
