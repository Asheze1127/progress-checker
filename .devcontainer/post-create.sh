#!/usr/bin/env bash
set -euo pipefail

workspace_dir="/workspaces/progress-checker"

cd "${workspace_dir}"

if ! command -v aws >/dev/null 2>&1; then
  echo "AWS CLI is not available yet. Rebuild the container and rerun postCreate if needed."
  exit 0
fi

aws_profile="${AWS_PROFILE:-default}"

if [ "${aws_profile}" = "default" ]; then
  profile_args=()
  verify_command="aws sts get-caller-identity"
  configure_command="aws configure"
else
  profile_args=(--profile "${aws_profile}")
  verify_command="aws sts get-caller-identity --profile ${aws_profile}"
  configure_command="aws configure --profile ${aws_profile}"
fi

if aws sts get-caller-identity "${profile_args[@]}" >/dev/null 2>&1; then
  echo "AWS CLI authentication is ready for profile '${aws_profile}'."
  exit 0
fi

sso_start_url="$(aws configure get sso_start_url "${profile_args[@]}" 2>/dev/null || true)"
if [ -n "${sso_start_url}" ]; then
  auth_command="aws sso login --profile ${aws_profile}"
else
  auth_command="${configure_command}"
fi

cat <<EOF
AWS CLI authentication is not ready for profile '${aws_profile}'.
Run the following command in the container:

  ${auth_command}

After authentication, verify it with:

  ${verify_command}
EOF
