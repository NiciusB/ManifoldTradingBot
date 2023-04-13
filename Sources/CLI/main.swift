import Foundation

// Put stdout into line buffering mode, instead of arbitrary default buffer: https://stackoverflow.com/questions/34743607/swift-cannot-output-when-using-nstimer
setlinebuf(stdout)

print("Starting app...")

Task {
    do {
        let mainApp = try await MainApp()
        mainApp.startAppLogicLoopTimer()
    } catch {
        printErr(error)
        exit(1)
    }
}

RunLoop.main.run()
