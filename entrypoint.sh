#!/bin/bash
set -eu

if [ -z "${RUNNER_NAME}" ]; then
  echo "RUNNER_NAME must be set" 1>&2
  exit 1
fi

if [ -z "${RUNNER_REPO}" ]; then
  echo "RUNNER_REPO must be set" 1>&2
  exit 1
fi

if [ -z "${RUNNER_TOKEN}" ]; then
  echo "RUNNER_TOKEN must be set" 1>&2
  exit 1
fi

if [ -n "${RUNNER_SSH_KNOWN_HOSTS:-}" ]; then
  echo "${RUNNER_SSH_KNOWN_HOSTS}" > "$HOME/.ssh/known_hosts"
  chmod 0600 "$HOME/.ssh/known_hosts"
fi

if [ -n "${RUNNER_SSH_ID_RSA:-}" ]; then
  echo "${RUNNER_SSH_ID_RSA}" > "$HOME/.ssh/id_rsa"
  chmod 0600 "$HOME/.ssh/id_rsa"
fi

if [ -n "${RUNNER_SSH_ID_RSA_PUB:-}" ]; then
  echo "${RUNNER_SSH_ID_RSA_PUB}" > "$HOME/.ssh/id_rsa.pub"
  chmod 0600 "$HOME/.ssh/id_rsa.pub"
fi

cd "$HOME"
./config.sh --unattended --replace --name "${RUNNER_NAME}" --url "https://github.com/${RUNNER_REPO}" --token "${RUNNER_TOKEN}"

unset RUNNER_SSH_ID_RSA RUNNER_SSH_ID_RSA_PUB RUNNER_SSH_KNOWN_HOSTS
unset RUNNER_NAME RUNNER_REPO RUNNER_TOKEN
exec ./run.sh --once
