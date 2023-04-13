#!/bin/bash
set -e

# Remove SwiftLintPlugin which causes release build to fail
sed -i.bak '/.plugin(name: "SwiftLintPlugin", package: "SwiftLint")/d' Package.swift
swift build -c release
rm Package.swift
mv Package.swift.bak Package.swift

rm -rf CLI_build
mkdir -p CLI_build 
cp -P .build/release/CLI CLI_build/
cp -P /usr/lib/swift/linux/lib*so* CLI_build/
