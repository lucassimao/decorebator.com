import ExpoModulesCore

public class ProfileImageProcessorModule: Module {
  public func definition() -> ModuleDefinition {
    Name("ProfileImageProcessor")

    AsyncFunction("prepare") { (uri: String) -> [String: Any] in
      try self.prepare(uri: uri)
    }
  }

  private func prepare(uri: String) throws -> [String: Any] {
    guard let inputURL = URL(string: uri), inputURL.isFileURL else {
      throw InvalidProfileImageException("Profile images must use a local file URI")
    }
    let outputURL = FileManager.default.temporaryDirectory
      .appendingPathComponent("profile-image-\(UUID().uuidString).jpg")
    do {
      let result = try ProfileImageProcessorCore.prepare(inputURL: inputURL, outputURL: outputURL)
      return [
        "uri": result.outputURL.absoluteString,
        "width": result.width,
        "height": result.height,
        "sourceWidth": result.sourceWidth,
        "sourceHeight": result.sourceHeight,
        "size": result.size,
      ]
    } catch {
      throw InvalidProfileImageException(error.localizedDescription)
    }
  }
}

private class InvalidProfileImageException: GenericException<String> {
  override var reason: String { param }
}
