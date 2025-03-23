#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(realpath "$(dirname "${BASH_SOURCE[0]}")/..")"
BUILD_ROOT="${REPO_ROOT}/build"
BUILD_BIN="${BUILD_ROOT}/bin"

NAME=buf
RELEASE=v1.50.1
OSX_RELEASE_256=28acc2c4c6b5698367f39d1de90b2259ce0d2d48ee4a17b636d69aa9498138d2
LINUX_RELEASE_256=96f701c56484899c85dd0fe3895fad431226e20312f2bd5af2e33a87d260d4ce

ARCH=x86_64

RELEASE_BINARY="${BUILD_BIN}/${NAME}-${RELEASE}"

main() {
  ensure_binary

  "${RELEASE_BINARY}" "$@"
}

ensure_binary() {
  if [[ ! -f "${RELEASE_BINARY}" ]]; then
    echo "info: Downloading ${NAME} ${RELEASE} to build environment"
    mkdir -p "${BUILD_BIN}"

    case "${OSTYPE}" in
      "darwin"*) os_type="Darwin"; sum="${OSX_RELEASE_256}" ;;
      "linux"*) os_type="Linux"; sum="${LINUX_RELEASE_256}" ;;
      *) echo "error: Unsupported OS '${OSTYPE}' for ${NAME} install, please install manually" && exit 1 ;;
    esac

    release_archive="/tmp/${NAME}-${RELEASE}.tar.gz"
    curl -sSL -o "${release_archive}" \
      "https://github.com/bufbuild/buf/releases/download/${RELEASE}/buf-${os_type}-${ARCH}.tar.gz"
    echo "${sum}" ${release_archive} | sha256sum --check --quiet -

    release_tmp_dir="/tmp/${NAME}-${RELEASE}"
    mkdir -p "${release_tmp_dir}"
    tar -xzf "${release_archive}" --strip=1 -C "${release_tmp_dir}"

    if [[ ! -f "${RELEASE_BINARY}" ]]; then
      find "${BUILD_BIN}" -maxdepth 0 -regex '.*/'${NAME}'-[A-Za-z0-9\.]+$' -exec rm {} \;  # cleanup older versions
      mv "${release_tmp_dir}/bin/${NAME}" "${RELEASE_BINARY}"
    fi

    # Cleanup stale resources.
    rm "${release_archive}"
    rm -rf "${release_tmp_dir}"
  fi
}

main "$@"