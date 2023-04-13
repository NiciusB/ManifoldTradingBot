#!/bin/bash
set -e

# Remove SwiftLintPlugin which causes linux builds to fail
sed -i.bak '/.plugin(name: "SwiftLintPlugin", package: "SwiftLint")/d' Package.swift
swift test
