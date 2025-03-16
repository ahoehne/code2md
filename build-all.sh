#!/bin/bash

# Script to build the application for multiple targets in parallel using background jobs

appName="code2md"
distDir="dist"

# Remove and recreate the dist directory
rm -rdf "$distDir" 2>/dev/null
mkdir -p "$distDir" || exit 100

xFlag=""
versionNumber=""
if [[ "$1" == v* ]] ; then
  xFlag="main.VersionNumber=$1"
  versionNumber="$1"
fi

# Define the target architectures and their respective file suffixes
declare -A targets=(
  ["windows-amd64"]=".exe"
  ["windows-arm64"]=".exe"
  ["darwin-amd64"]=""
  ["darwin-arm64"]=""
  ["linux-amd64"]=""
  ["linux-arm64"]=""
)

# Function to build for a specific target
build_target() {
  local target=$1
  IFS='-' read -r GOOS GOARCH <<< "$target"
  local suffix="${targets[$target]}"

  echo "[$(date +%H:%I:%S)] Building for $GOOS-$GOARCH..."
  export GOOS GOARCH
  if [[ $xFlag != "" ]] && go build -ldflags "-X $xFlag" -o "$distDir/${appName}-${GOOS}-${GOARCH}${suffix}"; then
    echo "[$(date +%H:%I:%S)] Build successful for $GOOS-$GOARCH (Version: $versionNumber)"
    return 0
  elif go build -o "$distDir/${appName}-${GOOS}-${GOARCH}${suffix}"; then
    echo "[$(date +%H:%I:%S)] Build successful for $GOOS-$GOARCH"
    return 0
  else
    echo "[$(date +%H:%I:%S)] Build failed for $GOOS-$GOARCH" >&2
    return 200
  fi
}

export appName
export distDir
export -f build_target
export -A targets

# Array to store PIDs of background jobs
pids=()
returnStatus=0

# Start a build job for each target in the background
for target in "${!targets[@]}"; do
  build_target "$target" &
  pids+=($!)
done

# Wait for all background jobs to finish
for pid in "${pids[@]}"; do
  wait "$pid" || returnStatus=$?
done

# Check if any build failed
if [ "$returnStatus" -gt "0" ]; then
  exit "$returnStatus"
fi

# generate checksums
echo "Generating Checksums"
cd dist && sha256sum -- * > CHECKSUMS.txt && cd .. && echo "Checksums generated" && exit 0
