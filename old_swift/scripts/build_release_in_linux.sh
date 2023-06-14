#!/bin/bash
set -e

# Remove SwiftLintPlugin which causes release build to fail
trap "rm Package.swift && mv Package.swift.bak Package.swift; exit" INT TERM EXIT
cp Package.swift Package.swift.bak
sed -i '/.plugin(name: "SwiftLintPlugin", package: "SwiftLint")/d' Package.swift
sed -i '/.package(url: "https:\/\/github.com\/realm\/SwiftLint.git", from: "0.51.0"),/d' Package.swift

swift build -c release

rm -rf CLI_build
mkdir -p CLI_build
mkdir -p CLI_build/bin
cp -P .build/release/CLI CLI_build/bin/
cp -P /usr/lib/swift/linux/lib*so* CLI_build/bin/
rm CLI_build/bin/lib_InternalSwiftScan.so
chmod 777 CLI_build/bin/CLI
echo "Run bin/CLI" > CLI_build/readme.txt
