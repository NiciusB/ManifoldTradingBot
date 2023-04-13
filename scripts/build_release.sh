# Remove SwiftLintPlugin which causes release build to fail
sed -i.bak '/.plugin(name: "SwiftLintPlugin", package: "SwiftLint")/d' Package.swift
swift build -c release
rm Package.swift
mv Package.swift.bak Package.swift

rm -rf .build/CLI_build .build/CLI_build.tar.gz
mkdir -p .build/CLI_build 
cp -P .build/release/CLI .build/CLI_build/
cp -P /usr/lib/swift/linux/lib*so* .build/CLI_build/
tar -zcvf .build/CLI_build.tar.gz .build/CLI_build/
