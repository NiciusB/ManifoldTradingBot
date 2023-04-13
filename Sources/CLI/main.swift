import Foundation

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
