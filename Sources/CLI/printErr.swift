import Foundation

extension FileHandle: TextOutputStream {
  public func write(_ string: String) {
    let data = Data(string.utf8)
    self.write(data)
  }
}

func printErr(_ message: Any) {
    var standardError = FileHandle.standardError
    print(message, to: &standardError)
}
