#!/bin/bash

appName="code2md"
distDir="dist"

rm -rdf "$distDir" 2>/dev/null
mkdir -p "$distDir" || exit 100

xFlag=""
versionNumber=""
if [[ "$1" == v* ]]; then
  xFlag="main.VersionNumber=$1"
  versionNumber="$1"
fi

declare -A targets=(
  ["windows-amd64"]=".exe"
  ["windows-arm64"]=".exe"
  ["darwin-amd64"]=""
  ["darwin-arm64"]=""
  ["linux-amd64"]=""
  ["linux-arm64"]=""
)

build_target() {
  local target=$1
  local suffix=$2
  IFS='-' read -r GOOS GOARCH <<< "$target"

  echo "[$(date +%H:%M:%S)] Building for $GOOS-$GOARCH..."
  export GOOS GOARCH
  if [[ $xFlag != "" ]] && go build -ldflags "-X $xFlag" -o "$distDir/${appName}-${GOOS}-${GOARCH}${suffix}"; then
    echo "[$(date +%H:%M:%S)] Build successful for $GOOS-$GOARCH (Version: $versionNumber)"
    return 0
  elif go build -o "$distDir/${appName}-${GOOS}-${GOARCH}${suffix}"; then
    echo "[$(date +%H:%M:%S)] Build successful for $GOOS-$GOARCH"
    return 0
  else
    echo "[$(date +%H:%M:%S)] Build failed for $GOOS-$GOARCH" >&2
    return 200
  fi
}

export appName
export distDir
export xFlag
export versionNumber
export -f build_target

pids=()
returnStatus=0

for target in "${!targets[@]}"; do
  build_target "$target" "${targets[$target]}" &
  pids+=($!)
done

for pid in "${pids[@]}"; do
  wait "$pid" || returnStatus=$?
done

if [ "$returnStatus" -gt "0" ]; then
  exit "$returnStatus"
fi

echo "Generating Checksums"
cd dist && sha256sum -- * > CHECKSUMS.txt && cd .. && echo "Checksums generated" && exit 0
