#!/usr/bin/env bash
set -euo pipefail

workspace_dir="/workspaces/progress-checker"
tool_versions_file="${workspace_dir}/.tool-versions"
asdf_dir="/home/vscode/.asdf"
asdf_bin="${asdf_dir}/bin/asdf"
asdf_version="0.18.0"

append_if_missing() {
  local file="$1"
  local line="$2"

  if ! grep -Fqx "$line" "$file"; then
    printf '%s\n' "$line" >> "$file"
  fi
}

setup_asdf() {
  local asdf_arch
  local temp_dir

  if [ -x "${asdf_bin}" ]; then
    return
  fi

  asdf_arch="$(dpkg --print-architecture)"
  case "${asdf_arch}" in
    amd64|arm64)
      ;;
    *)
      echo "Unsupported architecture: ${asdf_arch}" >&2
      exit 1
      ;;
  esac

  temp_dir="$(mktemp -d)"
  mkdir -p "${asdf_dir}/bin"
  curl -fsSLo "${temp_dir}/asdf.tar.gz" "https://github.com/asdf-vm/asdf/releases/download/v${asdf_version}/asdf-v${asdf_version}-linux-${asdf_arch}.tar.gz"
  tar -xzf "${temp_dir}/asdf.tar.gz" -C "${asdf_dir}/bin" asdf
  chmod +x "${asdf_bin}"
  rm -rf "${temp_dir}"
}

setup_asdf_plugins() {
  # `asdf plugin add` fails if the plugin is already installed.
  cut -d' ' -f1 "${tool_versions_file}" | grep -v '^#' | xargs -n1 "${asdf_bin}" plugin add || true
}

install_tools_from_tool_versions() {
  "${asdf_bin}" install
  "${asdf_bin}" reshim
}

prepare_nodejs_plugin() {
  if [ -x "${asdf_dir}/plugins/nodejs/bin/import-release-team-keyring" ]; then
    bash "${asdf_dir}/plugins/nodejs/bin/import-release-team-keyring"
  fi
}

setup_nodejs() {
  corepack enable
  corepack prepare pnpm@latest --activate

  if [ -f package.json ]; then
    pnpm install
  fi
}

install_go_tool() {
  local binary_name="$1"
  local package_name="$2"

  if [ -x "/go/bin/${binary_name}" ]; then
    return
  fi

  GOBIN=/go/bin go install "${package_name}"
}

setup_golang() {
  install_go_tool gopls golang.org/x/tools/gopls@latest
  install_go_tool dlv github.com/go-delve/delve/cmd/dlv@latest
  install_go_tool staticcheck honnef.co/go/tools/cmd/staticcheck@latest
  install_go_tool gotests github.com/cweill/gotests/gotests@latest
  install_go_tool gomodifytags github.com/fatih/gomodifytags@latest
  install_go_tool impl github.com/josharian/impl@latest
}

setup_shrc() {
  touch /home/vscode/.asdfrc /home/vscode/.bashrc /home/vscode/.zshrc

  append_if_missing /home/vscode/.asdfrc "legacy_version_file = yes"
  append_if_missing /home/vscode/.bashrc 'export HISTFILE=/commandhistory/.bash_history'
  append_if_missing /home/vscode/.bashrc 'PROMPT_COMMAND="history -a"'
  append_if_missing /home/vscode/.bashrc 'export ASDF_DATA_DIR=/home/vscode/.asdf'
  append_if_missing /home/vscode/.bashrc 'export PATH="$HOME/.local/bin:${ASDF_DATA_DIR:-$HOME/.asdf}/bin:${ASDF_DATA_DIR:-$HOME/.asdf}/shims:$PATH"'
  append_if_missing /home/vscode/.bashrc '[[ -f "${ASDF_DATA_DIR:-$HOME/.asdf}/plugins/golang/set-env.bash" ]] && . "${ASDF_DATA_DIR:-$HOME/.asdf}/plugins/golang/set-env.bash"'
  append_if_missing /home/vscode/.bashrc 'command -v direnv >/dev/null 2>&1 && eval "$(direnv hook bash)"'
  append_if_missing /home/vscode/.zshrc 'export HISTFILE=/commandhistory/.zsh_history'
  append_if_missing /home/vscode/.zshrc 'HISTSIZE=10000'
  append_if_missing /home/vscode/.zshrc 'SAVEHIST=10000'
  append_if_missing /home/vscode/.zshrc 'setopt APPEND_HISTORY SHARE_HISTORY INC_APPEND_HISTORY'
  append_if_missing /home/vscode/.zshrc 'export ASDF_DATA_DIR=/home/vscode/.asdf'
  append_if_missing /home/vscode/.zshrc 'export PATH="$HOME/.local/bin:${ASDF_DATA_DIR:-$HOME/.asdf}/bin:${ASDF_DATA_DIR:-$HOME/.asdf}/shims:$PATH"'
  append_if_missing /home/vscode/.zshrc '[[ -f "${ASDF_DATA_DIR:-$HOME/.asdf}/plugins/golang/set-env.zsh" ]] && . "${ASDF_DATA_DIR:-$HOME/.asdf}/plugins/golang/set-env.zsh"'
  append_if_missing /home/vscode/.zshrc 'command -v direnv >/dev/null 2>&1 && eval "$(direnv hook zsh)"'
}

# Prepare paths backed by persistent volumes from docker-compose.yml.
# The asdf directory is kept across rebuilds for reusable tool state.
mkdir -p /commandhistory /go/bin "${asdf_dir}" "${workspace_dir}/node_modules"
touch /commandhistory/.bash_history /commandhistory/.zsh_history

cd "${workspace_dir}"

setup_asdf
export PATH="$HOME/.local/bin:${asdf_dir}/bin:${asdf_dir}/shims:/go/bin:${PATH}"
setup_asdf_plugins
prepare_nodejs_plugin
install_tools_from_tool_versions

# TODO: gh-35 Install aws-cdk from infra/package.json instead of managing it with asdf.
setup_nodejs
setup_golang
setup_shrc
