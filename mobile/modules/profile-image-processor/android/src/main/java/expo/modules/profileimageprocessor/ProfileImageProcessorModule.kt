package expo.modules.profileimageprocessor

import android.graphics.Bitmap
import android.graphics.BitmapFactory
import android.graphics.Matrix
import android.net.Uri
import android.util.Log
import androidx.exifinterface.media.ExifInterface
import expo.modules.kotlin.modules.Module
import expo.modules.kotlin.modules.ModuleDefinition
import java.io.File
import java.io.FileInputStream
import java.io.FileOutputStream
import java.io.InputStream
import java.io.ByteArrayOutputStream
import kotlin.math.max
import kotlin.math.roundToInt

private const val MAX_OUTPUT_DIMENSION = 2048
private const val MAX_OUTPUT_BYTES = 5L * 1024L * 1024L
private const val MAX_SOURCE_BYTES = 25L * 1024L * 1024L
private const val MAX_SOURCE_DIMENSION = 20_000
private const val MAX_SOURCE_PIXELS = 100_000_000L
private const val PROFILE_TEMP_STALE_MILLIS = 60L * 60L * 1000L
private const val PROFILE_TEMP_SWEEP_LIMIT = 32
private const val PROFILE_TEMP_LOG_TAG = "ProfileImageProcessor"

class ProfileImageProcessorModule : Module() {
  override fun definition() = ModuleDefinition {
    Name("ProfileImageProcessor")

    AsyncFunction("prepare") { uri: String ->
      prepare(Uri.parse(uri))
    }
  }

  private fun prepare(uri: Uri): Map<String, Any> {
    val context = appContext.reactContext
      ?: throw IllegalStateException("Application context is unavailable")
    if (uri.scheme != "file" && uri.scheme != "content") {
      throw IllegalArgumentException("Profile images must use a local URI")
    }

    sweepStaleProfileFiles(context.cacheDir)
    val snapshot = File.createTempFile("profile-source-", ".bin", context.cacheDir)
    try {
      context.contentResolver.openInputStream(uri).use { input ->
        requireNotNull(input) { "Profile image cannot be opened" }
        snapshotProfileSource(input, snapshot)
      }
      return prepareSnapshot(snapshot, context.cacheDir)
    } finally {
      deleteTemporaryFile(snapshot)
    }
  }

  internal fun prepareSnapshot(snapshot: File, cacheDir: File): Map<String, Any> {
    validateSourceMagic(snapshot)
    readDeclaredDimensions(snapshot)?.let { (width, height) ->
      validateSourceDimensions(width, height)
    }
    val bounds = BitmapFactory.Options().apply { inJustDecodeBounds = true }
    BitmapFactory.decodeFile(snapshot.absolutePath, bounds)
    validateSourceDimensions(bounds.outWidth, bounds.outHeight)

    val orientation = ExifInterface(snapshot).getAttributeInt(
      ExifInterface.TAG_ORIENTATION,
      ExifInterface.ORIENTATION_NORMAL,
    )
    val decodeOptions = BitmapFactory.Options().apply {
      inSampleSize = sampleSizeFor(bounds.outWidth, bounds.outHeight)
      inPreferredConfig = Bitmap.Config.ARGB_8888
    }
    val decoded = BitmapFactory.decodeFile(snapshot.absolutePath, decodeOptions)
      ?: throw IllegalArgumentException("Profile image cannot be decoded")

    var working = orient(decoded, orientation)
    if (working !== decoded) decoded.recycle()
    val largest = max(working.width, working.height)
    if (largest > MAX_OUTPUT_DIMENSION) {
      val scale = MAX_OUTPUT_DIMENSION.toDouble() / largest.toDouble()
      val width = max(1, (working.width * scale).roundToInt())
      val height = max(1, (working.height * scale).roundToInt())
      val scaled = Bitmap.createScaledBitmap(working, width, height, true)
      if (scaled !== working) working.recycle()
      working = scaled
    }

    val output = File.createTempFile("profile-image-", ".jpg", cacheDir)
    try {
      FileOutputStream(output).use { stream ->
        check(working.compress(Bitmap.CompressFormat.JPEG, 80, stream)) {
          "Profile image could not be encoded"
        }
      }
      if (output.length() <= 0L || output.length() > MAX_OUTPUT_BYTES) {
        throw IllegalArgumentException("Profile image output size is invalid")
      }
      val stripped = stripJpegMetadata(output.readBytes())
      FileOutputStream(output, false).use { it.write(stripped) }
      val outputSize = stripped.size.toLong()
      if (outputSize <= 0L || outputSize > MAX_OUTPUT_BYTES) {
        throw IllegalArgumentException("Profile image output size is invalid")
      }
      return mapOf(
        "uri" to Uri.fromFile(output).toString(),
        "width" to working.width,
        "height" to working.height,
        "sourceWidth" to bounds.outWidth,
        "sourceHeight" to bounds.outHeight,
        "size" to outputSize.toDouble(),
      )
    } catch (error: Throwable) {
      deleteTemporaryFile(output)
      throw error
    } finally {
      working.recycle()
    }
  }

  internal fun stripJpegMetadata(source: ByteArray): ByteArray {
    if (source.size < 4 || source[0] != 0xff.toByte() || source[1] != 0xd8.toByte()) {
      throw IllegalArgumentException("Profile image output is not JPEG")
    }
    val output = ByteArrayOutputStream(source.size)
    output.write(source, 0, 2)
    var markerStart = 2
    while (markerStart < source.size) {
      if (source[markerStart] != 0xff.toByte()) {
        throw IllegalArgumentException("Profile image output has invalid JPEG markers")
      }
      var markerIndex = markerStart
      while (markerIndex < source.size && source[markerIndex] == 0xff.toByte()) markerIndex++
      if (markerIndex >= source.size) {
        throw IllegalArgumentException("Profile image output is truncated")
      }
      val marker = source[markerIndex].toInt() and 0xff
      if (marker == 0xda) {
        output.write(source, markerStart, source.size - markerStart)
        return output.toByteArray()
      }
      if (marker == 0xd9) {
        if (markerIndex + 1 != source.size) {
          throw IllegalArgumentException("Profile image output has trailing bytes")
        }
        output.write(source, markerStart, source.size - markerStart)
        return output.toByteArray()
      }
      if (marker == 0x00 || marker == 0x01 || marker in 0xd0..0xd7) {
        val markerEnd = markerIndex + 1
        output.write(source, markerStart, markerEnd - markerStart)
        markerStart = markerEnd
        continue
      }
      if (markerIndex + 2 >= source.size) {
        throw IllegalArgumentException("Profile image output is truncated")
      }
      val segmentLength = ((source[markerIndex + 1].toInt() and 0xff) shl 8) or
        (source[markerIndex + 2].toInt() and 0xff)
      val segmentEnd = markerIndex + 1 + segmentLength
      if (segmentLength < 2 || segmentEnd > source.size) {
        throw IllegalArgumentException("Profile image output has an invalid segment")
      }
      if (marker !in 0xe0..0xef && marker != 0xfe) {
        output.write(source, markerStart, segmentEnd - markerStart)
      }
      markerStart = segmentEnd
    }
    throw IllegalArgumentException("Profile image output has no scan")
  }

  internal fun sweepStaleProfileFiles(cacheDir: File, nowMillis: Long = System.currentTimeMillis()): Int {
    val candidates = cacheDir.listFiles()?.asSequence()
      ?.filter { file ->
        file.isFile &&
          (file.name.startsWith("profile-source-") || file.name.startsWith("profile-image-")) &&
          nowMillis - file.lastModified() >= PROFILE_TEMP_STALE_MILLIS
      }
      ?.sortedBy { it.lastModified() }
      ?.take(PROFILE_TEMP_SWEEP_LIMIT)
      ?.toList()
      .orEmpty()
    var deleted = 0
    for (candidate in candidates) {
      if (candidate.delete() || !candidate.exists()) {
        deleted++
      } else {
        Log.w(PROFILE_TEMP_LOG_TAG, "Failed to remove stale profile image temporary file")
      }
    }
    return deleted
  }

  private fun deleteTemporaryFile(file: File) {
    if (file.exists() && !file.delete()) {
      Log.w(PROFILE_TEMP_LOG_TAG, "Failed to remove profile image temporary file")
    }
  }

  internal fun snapshotProfileSource(input: InputStream, target: File) {
    FileOutputStream(target).use { output ->
      val buffer = ByteArray(DEFAULT_BUFFER_SIZE)
      var copied = 0L
      while (true) {
        val read = input.read(buffer)
        if (read < 0) break
        if (read == 0) continue
        copied += read
        if (copied > MAX_SOURCE_BYTES) {
          throw IllegalArgumentException("Profile image source is too large")
        }
        output.write(buffer, 0, read)
      }
      if (copied <= 0L) {
        throw IllegalArgumentException("Profile image source is empty")
      }
    }
  }

  internal fun validateSourceMagic(snapshot: File) {
    val header = ByteArray(12)
    val count = FileInputStream(snapshot).use { it.read(header) }
    val jpeg = count >= 3 && header[0] == 0xff.toByte() &&
      header[1] == 0xd8.toByte() && header[2] == 0xff.toByte()
    val png = count >= 8 && header.copyOfRange(0, 8).contentEquals(
      byteArrayOf(0x89.toByte(), 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a),
    )
    val webp = count >= 12 && header.copyOfRange(0, 4).decodeToString() == "RIFF" &&
      header.copyOfRange(8, 12).decodeToString() == "WEBP"
    val isoMedia = if (count >= 12 && header.copyOfRange(4, 8).decodeToString() == "ftyp") {
      header.copyOfRange(8, 12).decodeToString() in setOf(
        "heic", "heix", "hevc", "hevx", "mif1", "msf1", "avif", "avis",
      )
    } else {
      false
    }
    if (!jpeg && !png && !webp && !isoMedia) {
      throw IllegalArgumentException("Profile image format is unsupported")
    }
  }

  internal fun validateSourceDimensions(width: Int, height: Int) {
    if (width <= 0 || height <= 0 || width > MAX_SOURCE_DIMENSION || height > MAX_SOURCE_DIMENSION ||
      width.toLong() > MAX_SOURCE_PIXELS / height.toLong()
    ) {
      throw IllegalArgumentException("Profile image dimensions are unsupported")
    }
  }

  private fun readDeclaredDimensions(snapshot: File): Pair<Int, Int>? {
    FileInputStream(snapshot).buffered().use { input ->
      val first = input.read()
      val second = input.read()
      if (first == 0xff && second == 0xd8) {
        return readJpegDimensions(input)
      }
      if (first == 0x89 && second == 0x50) {
        val rest = ByteArray(22)
        if (input.readFully(rest) && rest.copyOfRange(0, 6).contentEquals(
            byteArrayOf(0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a),
          ) && rest.copyOfRange(10, 14).decodeToString() == "IHDR"
        ) {
          val width = readUnsignedInt(rest, 14)
          val height = readUnsignedInt(rest, 18)
          if (width > Int.MAX_VALUE || height > Int.MAX_VALUE) return Pair(-1, -1)
          return Pair(width.toInt(), height.toInt())
        }
      }
    }
    return null
  }

  private fun readJpegDimensions(input: InputStream): Pair<Int, Int>? {
    while (true) {
      var prefix = input.read()
      while (prefix >= 0 && prefix != 0xff) prefix = input.read()
      if (prefix < 0) return null

      var marker = input.read()
      while (marker == 0xff) marker = input.read()
      if (marker < 0 || marker == 0xd9 || marker == 0xda) return null
      if (marker == 0x00 || marker == 0x01 || marker in 0xd0..0xd8) continue

      val lengthHigh = input.read()
      val lengthLow = input.read()
      if (lengthHigh < 0 || lengthLow < 0) return null
      val payloadLength = (lengthHigh shl 8 or lengthLow) - 2
      if (payloadLength < 0) return null
      if (marker in setOf(0xc0, 0xc1, 0xc2, 0xc3, 0xc5, 0xc6, 0xc7, 0xc9, 0xca, 0xcb, 0xcd, 0xce, 0xcf)) {
        val frame = ByteArray(5)
        if (payloadLength < frame.size || !input.readFully(frame)) return null
        val height = frame[1].toUByte().toInt() shl 8 or frame[2].toUByte().toInt()
        val width = frame[3].toUByte().toInt() shl 8 or frame[4].toUByte().toInt()
        return Pair(width, height)
      }
      if (!input.skipFully(payloadLength.toLong())) return null
    }
  }

  private fun InputStream.readFully(target: ByteArray): Boolean {
    var offset = 0
    while (offset < target.size) {
      val count = read(target, offset, target.size - offset)
      if (count < 0) return false
      if (count == 0) continue
      offset += count
    }
    return true
  }

  private fun InputStream.skipFully(byteCount: Long): Boolean {
    var remaining = byteCount
    while (remaining > 0) {
      val skipped = skip(remaining)
      if (skipped > 0) {
        remaining -= skipped
      } else if (read() < 0) {
        return false
      } else {
        remaining--
      }
    }
    return true
  }

  private fun readUnsignedInt(bytes: ByteArray, offset: Int): Long =
    (bytes[offset].toUByte().toLong() shl 24) or
      (bytes[offset + 1].toUByte().toLong() shl 16) or
      (bytes[offset + 2].toUByte().toLong() shl 8) or
      bytes[offset + 3].toUByte().toLong()

  private fun sampleSizeFor(width: Int, height: Int): Int {
    var sample = 1
    while ((width + sample - 1) / sample > MAX_OUTPUT_DIMENSION ||
      (height + sample - 1) / sample > MAX_OUTPUT_DIMENSION
    ) {
      sample *= 2
    }
    return sample
  }

  private fun orient(source: Bitmap, orientation: Int): Bitmap {
    val matrix = Matrix()
    when (orientation) {
      ExifInterface.ORIENTATION_FLIP_HORIZONTAL -> matrix.setScale(-1f, 1f)
      ExifInterface.ORIENTATION_ROTATE_180 -> matrix.setRotate(180f)
      ExifInterface.ORIENTATION_FLIP_VERTICAL -> matrix.setScale(1f, -1f)
      ExifInterface.ORIENTATION_TRANSPOSE -> {
        matrix.setRotate(90f)
        matrix.postScale(-1f, 1f)
      }
      ExifInterface.ORIENTATION_ROTATE_90 -> matrix.setRotate(90f)
      ExifInterface.ORIENTATION_TRANSVERSE -> {
        matrix.setRotate(-90f)
        matrix.postScale(-1f, 1f)
      }
      ExifInterface.ORIENTATION_ROTATE_270 -> matrix.setRotate(-90f)
      else -> return source
    }
    return Bitmap.createBitmap(source, 0, 0, source.width, source.height, matrix, true)
  }
}
