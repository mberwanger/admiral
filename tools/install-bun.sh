#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="${REPO_ROOT:-"$(realpath "$(dirname "${BASH_SOURCE[0]}")/..")"}"
BUILD_ROOT="${REPO_ROOT}/build"
BUILD_BIN="${BUILD_ROOT}/bin"

NAME=bun
RELEASE=v1.2.5
OSX_X64_RELEASE_SUM=969f03626233168f3cdc17e9af9e7c9de32200073831fc53ab9416035de7ea61
OSX_AARCH64_RELEASE_SUM=55e6520a172e22a37d4822f23afde90b26f36880ce3eaad348699418dbf788bb
LINUX_X64_RELEASE_SUM=88f64bedee330ff4d6328e3e90c669bc7ac5314927c604f85e329ea2a1d1979b
LINUX_AARCH64_RELEASE_SUM=ee1b5cabb5fdbb25640787e329846ef41384ca8699db504f45bfee015f348b58

ARCH=x64

RELEASE_BINARY="${BUILD_BIN}/${NAME}-${RELEASE}"

ensure_binary() {
  if [[ ! -f "${RELEASE_BINARY}" ]]; then
    echo "info: Downloading ${NAME} ${RELEASE} to build environment"
    mkdir -p "${BUILD_BIN}"

    case $(uname -m) in
      arm64) ARCH="aarch64";;
      aarch64) ARCH="aarch64";;
    esac

    case "${OSTYPE}" in
      "darwin"*)
        os_type="darwin"
        if [[ "${ARCH}" == "x64" ]]; then
          sum="${OSX_X64_RELEASE_SUM}"
        else
          sum="${OSX_AARCH64_RELEASE_SUM}"
        fi
        ;;
      "linux"*)
        os_type="linux"
        if [[ "${ARCH}" == "x64" ]]; then
          sum="${LINUX_X64_RELEASE_SUM}"
        else
          sum="${LINUX_AARCH64_RELEASE_SUM}"
        fi
        ;;
      *)
        echo "error: Unsupported OS '${OSTYPE}' for ${NAME} install, please install manually" && exit 1
        ;;
    esac

    release_archive="/tmp/${NAME}-${RELEASE}.zip"
    curl -sSL -o "${release_archive}" \
      "https://github.com/oven-sh/bun/releases/download/bun-${RELEASE}/bun-${os_type}-${ARCH}.zip"
    echo "${sum}" "${release_archive}" | sha256sum --check --quiet -

    release_dir="/tmp/${NAME}-${os_type}"-${ARCH}
    unzip -q "${release_archive}" -d "/tmp"

    find "${BUILD_BIN}" -maxdepth 1 -regex '.*/'${NAME}'-[A-Za-z0-9\.]+$' -exec rm {} \;  # cleanup older versions
    mv "${release_dir}/bun" "${RELEASE_BINARY}"
    chmod +x "${RELEASE_BINARY}"
    ln -s ${NAME}-${RELEASE} ./build/bin/bun

    # Cleanup stale resources.
    rm "${release_archive}"
    rm -rf "${release_dir}"
  fi
}

ensure_fd() {
  if [[ "${OSTYPE}" == *"darwin"* ]]; then
    ulimit -n 1024
  fi
}

ensure_binary
ensure_fd