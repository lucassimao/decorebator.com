import Foundation
import ImageIO
import UniformTypeIdentifiers

private let maxOutputDimension = 2048
private let maxOutputBytes = 5 * 1024 * 1024
let maxProfileImageSourceBytes = 25 * 1024 * 1024
private let maxSourceDimension = 20_000
private let maxSourcePixels: Int64 = 100_000_000
private let profileTempStaleInterval: TimeInterval = 60 * 60
private let profileTempSweepLimit = 32

struct PreparedProfileImage {
  let outputURL: URL
  let width: Int
  let height: Int
  let sourceWidth: Int
  let sourceHeight: Int
  let size: Int
}

enum ProfileImageProcessorError: Error, LocalizedError {
  case invalid(String)

  var errorDescription: String? {
    switch self {
    case .invalid(let message):
      return message
    }
  }
}

enum ProfileImageProcessorCore {
  static func prepare(inputURL: URL, outputURL: URL) throws -> PreparedProfileImage {
    guard inputURL.isFileURL else {
      throw ProfileImageProcessorError.invalid("Profile images must use a local file URI")
    }

    sweepStaleProfileFiles()
    let snapshotURL = FileManager.default.temporaryDirectory
      .appendingPathComponent("profile-source-\(UUID().uuidString).bin")
    try snapshotSource(inputURL: inputURL, snapshotURL: snapshotURL)
    defer { removeTemporaryFile(snapshotURL) }

    let sourceOptions = [kCGImageSourceShouldCache: false] as CFDictionary
    guard let source = CGImageSourceCreateWithURL(snapshotURL as CFURL, sourceOptions),
          CGImageSourceGetCount(source) == 1,
          let properties = CGImageSourceCopyPropertiesAtIndex(source, 0, sourceOptions) as? [CFString: Any],
          let width = properties[kCGImagePropertyPixelWidth] as? Int,
          let height = properties[kCGImagePropertyPixelHeight] as? Int else {
      throw ProfileImageProcessorError.invalid("Profile image metadata cannot be read")
    }
    guard width > 0, height > 0,
          width <= maxSourceDimension, height <= maxSourceDimension,
          Int64(width) <= maxSourcePixels / Int64(height) else {
      throw ProfileImageProcessorError.invalid("Profile image dimensions are unsupported")
    }

    let thumbnailOptions = [
      kCGImageSourceCreateThumbnailFromImageAlways: true,
      kCGImageSourceCreateThumbnailWithTransform: true,
      kCGImageSourceThumbnailMaxPixelSize: maxOutputDimension,
      kCGImageSourceShouldCacheImmediately: true,
    ] as CFDictionary
    guard let thumbnail = CGImageSourceCreateThumbnailAtIndex(source, 0, thumbnailOptions) else {
      throw ProfileImageProcessorError.invalid("Profile image cannot be decoded")
    }

    guard let destination = CGImageDestinationCreateWithURL(
      outputURL as CFURL,
      UTType.jpeg.identifier as CFString,
      1,
      nil
    ) else {
      throw ProfileImageProcessorError.invalid("Profile image output cannot be created")
    }
    var completed = false
    defer {
      if !completed {
        removeTemporaryFile(outputURL)
      }
    }
    CGImageDestinationAddImage(
      destination,
      thumbnail,
      [kCGImageDestinationLossyCompressionQuality: 0.8] as CFDictionary
    )
    guard CGImageDestinationFinalize(destination) else {
      throw ProfileImageProcessorError.invalid("Profile image could not be encoded")
    }
    guard let encodedSize = try outputURL.resourceValues(forKeys: [.fileSizeKey]).fileSize,
          encodedSize > 0, encodedSize <= maxOutputBytes else {
      throw ProfileImageProcessorError.invalid("Profile image output size is invalid")
    }
    let stripped = try stripJPEGMetadata(Data(contentsOf: outputURL))
    try stripped.write(to: outputURL, options: .atomic)
    let measuredSize = stripped.count
    guard measuredSize > 0, measuredSize <= maxOutputBytes else {
      throw ProfileImageProcessorError.invalid("Profile image output size is invalid")
    }
    completed = true
    return PreparedProfileImage(
      outputURL: outputURL,
      width: thumbnail.width,
      height: thumbnail.height,
      sourceWidth: width,
      sourceHeight: height,
      size: measuredSize
    )
  }

  static func stripJPEGMetadata(_ source: Data) throws -> Data {
    guard source.count >= 4, source[0] == 0xff, source[1] == 0xd8 else {
      throw ProfileImageProcessorError.invalid("Profile image output is not JPEG")
    }
    var output = Data(source[0..<2])
    var markerStart = 2
    while markerStart < source.count {
      guard source[markerStart] == 0xff else {
        throw ProfileImageProcessorError.invalid("Profile image output has invalid JPEG markers")
      }
      var markerIndex = markerStart
      while markerIndex < source.count, source[markerIndex] == 0xff { markerIndex += 1 }
      guard markerIndex < source.count else {
        throw ProfileImageProcessorError.invalid("Profile image output is truncated")
      }
      let marker = source[markerIndex]
      if marker == 0xda {
        output.append(source[markerStart..<source.count])
        return output
      }
      if marker == 0xd9 {
        guard markerIndex + 1 == source.count else {
          throw ProfileImageProcessorError.invalid("Profile image output has trailing bytes")
        }
        output.append(source[markerStart..<source.count])
        return output
      }
      if marker == 0x00 || marker == 0x01 || (0xd0...0xd7).contains(marker) {
        let markerEnd = markerIndex + 1
        output.append(source[markerStart..<markerEnd])
        markerStart = markerEnd
        continue
      }
      guard markerIndex + 2 < source.count else {
        throw ProfileImageProcessorError.invalid("Profile image output is truncated")
      }
      let segmentLength = Int(source[markerIndex + 1]) << 8 | Int(source[markerIndex + 2])
      let segmentEnd = markerIndex + 1 + segmentLength
      guard segmentLength >= 2, segmentEnd <= source.count else {
        throw ProfileImageProcessorError.invalid("Profile image output has an invalid segment")
      }
      if !(0xe0...0xef).contains(marker), marker != 0xfe {
        output.append(source[markerStart..<segmentEnd])
      }
      markerStart = segmentEnd
    }
    throw ProfileImageProcessorError.invalid("Profile image output has no scan")
  }

  static func snapshotSource(inputURL: URL, snapshotURL: URL) throws {
    guard FileManager.default.createFile(atPath: snapshotURL.path, contents: nil),
          let input = try? FileHandle(forReadingFrom: inputURL),
          let output = try? FileHandle(forWritingTo: snapshotURL) else {
      removeTemporaryFile(snapshotURL)
      throw ProfileImageProcessorError.invalid("Profile image cannot be opened")
    }
    defer {
      try? input.close()
      try? output.close()
    }
    do {
      var copied = 0
      while let chunk = try input.read(upToCount: 8192), !chunk.isEmpty {
        copied += chunk.count
        guard copied <= maxProfileImageSourceBytes else {
          throw ProfileImageProcessorError.invalid("Profile image source is too large")
        }
        try output.write(contentsOf: chunk)
      }
      guard copied > 0 else {
        throw ProfileImageProcessorError.invalid("Profile image source is empty")
      }
    } catch {
      removeTemporaryFile(snapshotURL)
      throw error
    }
  }

  @discardableResult
  static func sweepStaleProfileFiles(
    in directory: URL = FileManager.default.temporaryDirectory,
    now: Date = Date()
  ) -> Int {
    let fileManager = FileManager.default
    let urls: [URL]
    do {
      urls = try fileManager.contentsOfDirectory(
        at: directory,
        includingPropertiesForKeys: [.isRegularFileKey, .contentModificationDateKey],
        options: [.skipsHiddenFiles]
      )
    } catch {
      NSLog("ProfileImageProcessor: failed to enumerate stale temporary files: %@", error.localizedDescription)
      return 0
    }
    let candidates = urls.compactMap { url -> (URL, Date)? in
      guard url.lastPathComponent.hasPrefix("profile-source-") ||
              url.lastPathComponent.hasPrefix("profile-image-") else { return nil }
      guard let values = try? url.resourceValues(forKeys: [.isRegularFileKey, .contentModificationDateKey]),
            values.isRegularFile == true,
            let modified = values.contentModificationDate,
            now.timeIntervalSince(modified) >= profileTempStaleInterval else { return nil }
      return (url, modified)
    }.sorted { $0.1 < $1.1 }.prefix(profileTempSweepLimit)

    var deleted = 0
    for (url, _) in candidates {
      do {
        try fileManager.removeItem(at: url)
        deleted += 1
      } catch {
        NSLog("ProfileImageProcessor: failed to remove stale temporary file: %@", error.localizedDescription)
      }
    }
    return deleted
  }

  private static func removeTemporaryFile(_ url: URL) {
    guard FileManager.default.fileExists(atPath: url.path) else { return }
    do {
      try FileManager.default.removeItem(at: url)
    } catch {
      NSLog("ProfileImageProcessor: failed to remove temporary file: %@", error.localizedDescription)
    }
  }
}
