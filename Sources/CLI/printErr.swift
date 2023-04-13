import Foundation

extension FileHandle: TextOutputStream {
  public func write(_ string: String) {
    let data = Data(string.utf8)
    self.write(data)
  }
}

func printErr(_ items: Any...) {
    var standardError = FileHandle.standardError
    print(items, to: &standardError)
}
