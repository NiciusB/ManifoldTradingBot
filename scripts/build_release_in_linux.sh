#!/bin/bash
set -e

# Remove SwiftLintPlugin which causes release build to fail
sed -i.bak '/.plugin(name: "SwiftLintPlugin", package: "SwiftLint")/d' Package.swift
swift build -c release
rm Package.swift
mv Package.swift.bak Package.swift

rm -rf CLI_build
mkdir -p CLI_build
mkdir -p CLI_build/bin
cp -P .build/release/CLI CLI_build/bin/
cp -P /usr/lib/swift/linux/lib*so* CLI_build/bin/
rm CLI_build/bin/lib_InternalSwiftScan.so
chmod 777 CLI_build/bin/CLI
echo "Run bin/CLI" > CLI_build/readme.txt
