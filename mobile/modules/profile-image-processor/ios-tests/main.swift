import CoreGraphics
import Foundation
import ImageIO
import UniformTypeIdentifiers

enum SmokeFailure: Error {
  case assertion(String)
}

func requireSmoke(_ condition: @autoclosure () -> Bool, _ message: String) throws {
  if !condition() {
    throw SmokeFailure.assertion(message)
  }
}

func writeJPEG(width: Int, height: Int, orientation: Int? = nil, to url: URL) throws {
  let colorSpace = CGColorSpaceCreateDeviceRGB()
  guard let context = CGContext(
    data: nil,
    width: width,
    height: height,
    bitsPerComponent: 8,
    bytesPerRow: width * 4,
    space: colorSpace,
    bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue
  ) else {
    throw SmokeFailure.assertion("could not create fixture context")
  }
  context.setFillColor(red: 0.15, green: 0.35, blue: 0.75, alpha: 1)
  context.fill(CGRect(x: 0, y: 0, width: width, height: height))
  guard let image = context.makeImage(),
        let destination = CGImageDestinationCreateWithURL(
          url as CFURL, UTType.jpeg.identifier as CFString, 1, nil
        ) else {
    throw SmokeFailure.assertion("could not create fixture destination")
  }
  var properties: [CFString: Any] = [kCGImageDestinationLossyCompressionQuality: 0.9]
  if let orientation {
    properties[kCGImagePropertyOrientation] = orientation
  }
  CGImageDestinationAddImage(destination, image, properties as CFDictionary)
  try requireSmoke(CGImageDestinationFinalize(destination), "could not finalize fixture")
}

func jpegFrameMarkers(_ data: Data) -> Set<UInt8> {
  var result = Set<UInt8>()
  guard data.count > 1 else { return result }
  for index in 0..<(data.count - 1) where data[index] == 0xff {
    let marker = data[index + 1]
    if (0xc0...0xcf).contains(marker) && ![0xc4, 0xc8, 0xcc].contains(marker) {
      result.insert(marker)
    }
  }
  return result
}

func jpegMetadataMarkers(_ data: Data) -> Set<UInt8> {
  var result = Set<UInt8>()
  guard data.count > 1 else { return result }
  for index in 0..<(data.count - 1) where data[index] == 0xff {
    let marker = data[index + 1]
    if (0xe0...0xef).contains(marker) || marker == 0xfe {
      result.insert(marker)
    }
  }
  return result
}

func patchJPEGDimensions(_ url: URL, width: Int, height: Int) throws {
  var bytes = try Data(contentsOf: url)
  var offset = 2
  while offset + 8 < bytes.count {
    guard bytes[offset] == 0xff else { break }
    let marker = bytes[offset + 1]
    let length = Int(bytes[offset + 2]) << 8 | Int(bytes[offset + 3])
    if marker == 0xc0 {
      bytes[offset + 5] = UInt8(height >> 8)
      bytes[offset + 6] = UInt8(height & 0xff)
      bytes[offset + 7] = UInt8(width >> 8)
      bytes[offset + 8] = UInt8(width & 0xff)
      try bytes.write(to: url)
      return
    }
    offset += 2 + length
  }
  throw SmokeFailure.assertion("fixture has no baseline frame marker")
}

let root = FileManager.default.temporaryDirectory
  .appendingPathComponent("profile-image-ios-smoke-\(UUID().uuidString)")
try FileManager.default.createDirectory(at: root, withIntermediateDirectories: true)
defer { try? FileManager.default.removeItem(at: root) }

let pickerExactSource = root.appendingPathComponent("picker-exact-source.bin")
FileManager.default.createFile(atPath: pickerExactSource.path, contents: nil)
let pickerExactHandle = try FileHandle(forWritingTo: pickerExactSource)
try pickerExactHandle.truncate(atOffset: UInt64(ProfileImagePickerBoundedCopy.maxBytes))
try pickerExactHandle.close()
let pickerExactTarget = root.appendingPathComponent("picker-exact-target.bin")
try ProfileImagePickerBoundedCopy.copy(from: pickerExactSource, to: pickerExactTarget)
try requireSmoke(
  (try pickerExactTarget.resourceValues(forKeys: [.fileSizeKey]).fileSize) ==
    ProfileImagePickerBoundedCopy.maxBytes,
  "picker handoff changed an exact-limit source"
)

let pickerOversizedSource = root.appendingPathComponent("picker-oversized-source.bin")
FileManager.default.createFile(atPath: pickerOversizedSource.path, contents: nil)
let pickerOversizedHandle = try FileHandle(forWritingTo: pickerOversizedSource)
try pickerOversizedHandle.truncate(atOffset: UInt64(ProfileImagePickerBoundedCopy.maxBytes + 1))
try pickerOversizedHandle.close()
let pickerOversizedTarget = root.appendingPathComponent("picker-oversized-target.bin")
do {
  try ProfileImagePickerBoundedCopy.copy(from: pickerOversizedSource, to: pickerOversizedTarget)
  throw SmokeFailure.assertion("picker handoff accepted an oversized source")
} catch is ProfileImagePickerBoundedCopyError {}
try requireSmoke(
  !FileManager.default.fileExists(atPath: pickerOversizedTarget.path),
  "failed picker handoff retained a partial file"
)

let malformedPickerSource = root.appendingPathComponent("picker-malformed-source.bin")
try Data("metadata probe must happen later".utf8).write(to: malformedPickerSource)
let malformedPickerTarget = root.appendingPathComponent("picker-malformed-target.bin")
try ProfileImagePickerBoundedCopy.copy(from: malformedPickerSource, to: malformedPickerTarget)
try requireSmoke(
  try Data(contentsOf: malformedPickerTarget) == Data(contentsOf: malformedPickerSource),
  "picker handoff decoded or changed malformed input"
)

let missingPickerTarget = root.appendingPathComponent("picker-missing-target.bin")
do {
  try ProfileImagePickerBoundedCopy.copy(
    from: root.appendingPathComponent("picker-missing-source.bin"),
    to: missingPickerTarget
  )
  throw SmokeFailure.assertion("picker handoff accepted a missing representation")
} catch {}
try requireSmoke(
  !FileManager.default.fileExists(atPath: missingPickerTarget.path),
  "failed picker representation retained a partial file"
)

let pickerSweepNow = Date(timeIntervalSince1970: 10 * 60 * 60)
for index in 0..<40 {
  let stale = root.appendingPathComponent("profile-picker-source-stale-\(index).jpg")
  try Data("stale".utf8).write(to: stale)
  try FileManager.default.setAttributes(
    [.modificationDate: pickerSweepNow.addingTimeInterval(-2 * 60 * 60 - Double(index))],
    ofItemAtPath: stale.path
  )
}
let activePickerSource = root.appendingPathComponent("profile-picker-source-active.jpg")
try Data("active".utf8).write(to: activePickerSource)
try FileManager.default.setAttributes(
  [.modificationDate: pickerSweepNow],
  ofItemAtPath: activePickerSource.path
)
try requireSmoke(
  ProfileImagePickerBoundedCopy.sweepStale(in: root, now: pickerSweepNow) == 32,
  "picker source sweep exceeded its per-run bound"
)
let remainingPickerSources = try FileManager.default.contentsOfDirectory(atPath: root.path)
  .filter { $0.hasPrefix("profile-picker-source-stale-") }
try requireSmoke(remainingPickerSources.count == 8, "picker source sweep removed the wrong files")
try requireSmoke(
  FileManager.default.fileExists(atPath: activePickerSource.path),
  "picker source sweep removed an active file"
)

let sweepNow = Date(timeIntervalSince1970: 10 * 60 * 60)
for index in 0..<40 {
  let stale = root.appendingPathComponent("profile-image-stale-\(index).jpg")
  try Data("stale".utf8).write(to: stale)
  try FileManager.default.setAttributes(
    [.modificationDate: sweepNow.addingTimeInterval(-2 * 60 * 60 - Double(index))],
    ofItemAtPath: stale.path
  )
}
let activeTemporary = root.appendingPathComponent("profile-source-active.bin")
try Data("active".utf8).write(to: activeTemporary)
try FileManager.default.setAttributes([.modificationDate: sweepNow], ofItemAtPath: activeTemporary.path)
let unrelatedTemporary = root.appendingPathComponent("other-stale.bin")
try Data("unrelated".utf8).write(to: unrelatedTemporary)
try FileManager.default.setAttributes([.modificationDate: Date(timeIntervalSince1970: 0)], ofItemAtPath: unrelatedTemporary.path)
try requireSmoke(
  ProfileImageProcessorCore.sweepStaleProfileFiles(in: root, now: sweepNow) == 32,
  "stale temporary sweep exceeded its per-run bound"
)
let remainingStale = try FileManager.default.contentsOfDirectory(atPath: root.path)
  .filter { $0.hasPrefix("profile-image-stale-") }
try requireSmoke(remainingStale.count == 8, "stale temporary sweep removed the wrong files")
try requireSmoke(
  FileManager.default.fileExists(atPath: activeTemporary.path) &&
    FileManager.default.fileExists(atPath: unrelatedTemporary.path),
  "temporary sweep removed an active or unrelated file"
)
try requireSmoke(
  ProfileImageProcessorCore.sweepStaleProfileFiles(in: root, now: sweepNow) == 8,
  "stale temporary sweep did not clean the bounded remainder"
)

let exactSource = root.appendingPathComponent("exact-source.bin")
FileManager.default.createFile(atPath: exactSource.path, contents: nil)
let exactHandle = try FileHandle(forWritingTo: exactSource)
try exactHandle.truncate(atOffset: UInt64(maxProfileImageSourceBytes))
try exactHandle.close()
let exactSnapshot = root.appendingPathComponent("exact-snapshot.bin")
try ProfileImageProcessorCore.snapshotSource(inputURL: exactSource, snapshotURL: exactSnapshot)
try requireSmoke(
  (try exactSnapshot.resourceValues(forKeys: [.fileSizeKey]).fileSize) == maxProfileImageSourceBytes,
  "exact source byte limit was not preserved"
)

let oversizedSource = root.appendingPathComponent("oversized-source.bin")
FileManager.default.createFile(atPath: oversizedSource.path, contents: nil)
let oversizedHandle = try FileHandle(forWritingTo: oversizedSource)
try oversizedHandle.truncate(atOffset: UInt64(maxProfileImageSourceBytes + 1))
try oversizedHandle.close()
let oversizedSnapshot = root.appendingPathComponent("oversized-snapshot.bin")
do {
  try ProfileImageProcessorCore.snapshotSource(inputURL: oversizedSource, snapshotURL: oversizedSnapshot)
  throw SmokeFailure.assertion("oversized source bytes were accepted")
} catch is ProfileImageProcessorError {}
try requireSmoke(!FileManager.default.fileExists(atPath: oversizedSnapshot.path), "failed snapshot was retained")

let largeInput = root.appendingPathComponent("large.jpg")
let largeOutput = root.appendingPathComponent("large-output.jpg")
try writeJPEG(width: 3000, height: 1000, to: largeInput)
let large = try ProfileImageProcessorCore.prepare(inputURL: largeInput, outputURL: largeOutput)
try requireSmoke(large.sourceWidth == 3000 && large.sourceHeight == 1000, "source dimensions changed")
try requireSmoke(max(large.width, large.height) <= 2048, "output dimensions exceed limit")
try requireSmoke(large.size > 0 && large.size <= 5 * 1024 * 1024, "output byte size is invalid")
let markers = jpegFrameMarkers(try Data(contentsOf: largeOutput))
try requireSmoke(markers.contains(0xc0), "output is not baseline JPEG")
try requireSmoke(!markers.contains(0xc2), "output is progressive JPEG")
try requireSmoke(
  jpegMetadataMarkers(try Data(contentsOf: largeOutput)).isEmpty,
  "output retained JPEG metadata"
)

let metadataHeavy = root.appendingPathComponent("metadata-heavy.jpg")
try FileManager.default.copyItem(at: largeInput, to: metadataHeavy)
let metadataHandle = try FileHandle(forWritingTo: metadataHeavy)
try metadataHandle.truncate(atOffset: UInt64(maxProfileImageSourceBytes + 1))
try metadataHandle.close()
let metadataOutput = root.appendingPathComponent("metadata-heavy-output.jpg")
do {
  _ = try ProfileImageProcessorCore.prepare(inputURL: metadataHeavy, outputURL: metadataOutput)
  throw SmokeFailure.assertion("oversized metadata-heavy image was accepted")
} catch is ProfileImageProcessorError {}
try requireSmoke(!FileManager.default.fileExists(atPath: metadataOutput.path), "oversized image created output")

let orientedInput = root.appendingPathComponent("oriented.jpg")
let orientedOutput = root.appendingPathComponent("oriented-output.jpg")
try writeJPEG(width: 120, height: 60, orientation: 6, to: orientedInput)
let oriented = try ProfileImageProcessorCore.prepare(inputURL: orientedInput, outputURL: orientedOutput)
try requireSmoke(oriented.width == 60 && oriented.height == 120, "EXIF orientation was not applied")

let malformed = root.appendingPathComponent("malformed.jpg")
try Data("not an image".utf8).write(to: malformed)
do {
  _ = try ProfileImageProcessorCore.prepare(
    inputURL: malformed,
    outputURL: root.appendingPathComponent("malformed-output.jpg")
  )
  throw SmokeFailure.assertion("malformed fixture was accepted")
} catch is ProfileImageProcessorError {}

if let compatibilityPath = ProcessInfo.processInfo.environment["PROFILE_IMAGE_IOS_FIXTURE"] {
  let compatibilityURL = compatibilityPath.hasPrefix("/")
    ? URL(fileURLWithPath: compatibilityPath)
    : FileManager.default.urls(for: .documentDirectory, in: .userDomainMask)[0]
      .appendingPathComponent(compatibilityPath)
  try? FileManager.default.removeItem(at: compatibilityURL)
  try FileManager.default.copyItem(at: largeOutput, to: compatibilityURL)
}

let oversized = root.appendingPathComponent("oversized.jpg")
try writeJPEG(width: 20_001, height: 1, to: oversized)
do {
  _ = try ProfileImageProcessorCore.prepare(
    inputURL: oversized,
    outputURL: root.appendingPathComponent("oversized-output.jpg")
  )
  throw SmokeFailure.assertion("oversized source dimensions were accepted")
} catch is ProfileImageProcessorError {}

let oversizedPixels = root.appendingPathComponent("oversized-pixels.jpg")
try writeJPEG(width: 10, height: 10, to: oversizedPixels)
try patchJPEGDimensions(oversizedPixels, width: 20_000, height: 6_000)
do {
  _ = try ProfileImageProcessorCore.prepare(
    inputURL: oversizedPixels,
    outputURL: root.appendingPathComponent("oversized-pixels-output.jpg")
  )
  throw SmokeFailure.assertion("oversized source pixel product was accepted")
} catch is ProfileImageProcessorError {}
