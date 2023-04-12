# Remove SwiftLintPlugin which causes release build to fail
sed -i.bak '/.plugin(name: "SwiftLintPlugin", package: "SwiftLint")/d' Package.swift
swift build -c release
rm Package.swift
mv Package.swift.bak Package.swift

rm -rf .build/install
mkdir -p .build/install 
cp -P .build/release/CLI .build/install/
# cp -P /usr/lib/swift/linux/lib*so* .build/install/
