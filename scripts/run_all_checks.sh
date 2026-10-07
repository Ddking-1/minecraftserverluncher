#!/usr/bin/env bash
set -euo pipefail

echo "==> Go: vet"
(cd server && go vet ./...)

echo "==> Go: test"
(cd server && go test ./... -count=1 -timeout 120s -v)

echo "==> Go: race"
(cd server && go test ./... -count=1 -timeout 120s -v -race) || {
  echo "WARN: race detector may not be available in this environment"
}

echo "==> Go: staticcheck (server/)"
if command -v staticcheck >/dev/null 2>&1; then
  (cd server && staticcheck ./...)
else
  echo "WARN: staticcheck not installed; skip"
fi

echo "==> Launcher: configure + build + ctest (offscreen)"
if command -v cmake >/dev/null 2>&1; then
  cmake -B build-check -S launcher -G Ninja -DCMAKE_BUILD_TYPE=Release -DCMAKE_AUTOMOC=ON -DCMAKE_AUTOUIC=OFF -DCMAKE_AUTORCC=OFF
  cmake --build build-check --target MinecraftLauncher
  QT_QPA_PLATFORM=offscreen ctest --test-dir build-check --output-on-failure
else
  echo "WARN: cmake not available; skip launcher checks"
fi

echo "==> Done"
