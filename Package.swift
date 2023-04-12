// swift-tools-version: 5.8
// The swift-tools-version declares the minimum version of Swift required to build this package.

import PackageDescription

let package = Package(
    name: "CLI",
    platforms: [
        .macOS(.v13)
    ],
    products: [
        .executable(
            name: "CLI",
            targets: ["CLI"]
        )
    ],
    dependencies: [
        .package(
            url: "https://github.com/Flight-School/AnyCodable",
            from: "0.6.0"
        ),
        .package(
            url: "https://github.com/scinfu/SwiftSoup.git",
            from: "2.0.0"
        ),
        .package(url: "https://github.com/emilioschepis/swift-dotenv.git", from: "1.0.0"),
        .package(
            url: "https://github.com/Alamofire/Alamofire.git",
            from: "5.0.0"
        ),
        .package(
            url: "https://github.com/realm/SwiftLint.git",
            from: "0.51.0"
        )
    ],
    targets: [
        // Targets are the basic building blocks of a package, defining a module or a test suite.
        // Targets can depend on other targets in this package and products from dependencies.
        .executableTarget(
            name: "CLI",
            dependencies: [
                "AnyCodable",
                "SwiftSoup",
                "Alamofire",
                .product(name: "Dotenv", package: "swift-dotenv")
            ],
            plugins: [
                .plugin(name: "SwiftLintPlugin", package: "SwiftLint")
            ]
        ),
        .testTarget(
            name: "CLITests",
            dependencies: ["CLI"]
        )
    ]
)

for target in package.targets {
  target.swiftSettings = target.swiftSettings ?? []
  target.swiftSettings?.append(
    .unsafeFlags([
      // "-Xfrontend", "-warn-concurrency",
      "-Xfrontend", "-enable-actor-data-race-checks",
      "-enable-bare-slash-regex"
    ])
  )
}
