package expo.modules.profileimageprocessor

import android.graphics.Bitmap
import android.graphics.Color
import android.net.Uri
import androidx.exifinterface.media.ExifInterface
import expo.modules.imagepicker.exporters.ProfileImagePickerBoundedCopy
import java.io.ByteArrayInputStream
import java.io.File
import java.io.FileOutputStream
import java.io.InputStream
import kotlin.io.path.createTempDirectory
import org.junit.Assert.assertArrayEquals
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Assert.assertThrows
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config
import org.robolectric.annotation.GraphicsMode

@RunWith(RobolectricTestRunner::class)
@Config(sdk = [24, 28, 36])
@GraphicsMode(GraphicsMode.Mode.NATIVE)
class ProfileImageProcessorModuleTest {
  @Test
  fun pickerCopyAcceptsExactLimitWithoutParsingImageBytes() {
    val output = File.createTempFile("picker-exact-", ".bin")
    try {
      ProfileImagePickerBoundedCopy.copy(
        RepeatingInputStream(ProfileImagePickerBoundedCopy.MAX_BYTES),
        output,
      )
      assertEquals(ProfileImagePickerBoundedCopy.MAX_BYTES, output.length())
    } finally {
      output.delete()
    }
  }

  @Test
  fun pickerCopyRemovesOversizedPartialOutput() {
    val output = File.createTempFile("picker-oversized-", ".bin")
    assertThrows(IllegalArgumentException::class.java) {
      ProfileImagePickerBoundedCopy.copy(
        RepeatingInputStream(ProfileImagePickerBoundedCopy.MAX_BYTES + 1L),
        output,
      )
    }
    assertFalse(output.exists())
  }

  @Test
  fun pickerCopyPreservesMalformedBytesWithoutParsingThem() {
    val malformed = "not an image".encodeToByteArray()
    val output = File.createTempFile("picker-malformed-", ".bin")
    try {
      ProfileImagePickerBoundedCopy.copy(ByteArrayInputStream(malformed), output)
      assertArrayEquals(malformed, output.readBytes())
    } finally {
      output.delete()
    }
  }

  @Test
  fun pickerCopySweepsOnlyBoundedStaleOwnedSources() {
    val cacheDir = createTempDirectory("picker-sweep-").toFile()
    val now = 10L * 60L * 60L * 1000L
    try {
      repeat(40) { index ->
        File(cacheDir, "profile-picker-source-stale-$index.jpg").apply {
          writeText("stale")
          setLastModified(now - 2L * 60L * 60L * 1000L - index)
        }
      }
      val active = File(cacheDir, "profile-picker-source-active.jpg").apply {
        writeText("active")
        setLastModified(now)
      }
      val unrelated = File(cacheDir, "unrelated.jpg").apply { writeText("other") }

      assertEquals(32, ProfileImagePickerBoundedCopy.sweepStale(cacheDir, now))
      assertEquals(8, cacheDir.listFiles()!!.count { it.name.startsWith("profile-picker-source-stale-") })
      assertTrue(active.exists())
      assertTrue(unrelated.exists())
    } finally {
      cacheDir.deleteRecursively()
    }
  }

  @Test
  fun sweepsOnlyBoundedStaleProfileTemporaryFiles() {
    val cacheDir = createTempDirectory("profile-image-sweep-").toFile()
    val now = 10L * 60L * 60L * 1000L
    try {
      repeat(40) { index ->
        File(cacheDir, "profile-image-stale-$index.jpg").apply {
          writeText("stale")
          setLastModified(now - 2L * 60L * 60L * 1000L - index)
        }
      }
      val active = File(cacheDir, "profile-source-active.bin").apply {
        writeText("active")
        setLastModified(now)
      }
      val unrelated = File(cacheDir, "other-stale.bin").apply {
        writeText("unrelated")
        setLastModified(0)
      }

      assertEquals(32, ProfileImageProcessorModule().sweepStaleProfileFiles(cacheDir, now))
      assertEquals(8, cacheDir.listFiles()!!.count { it.name.startsWith("profile-image-stale-") })
      assertTrue(active.exists())
      assertTrue(unrelated.exists())
      assertEquals(8, ProfileImageProcessorModule().sweepStaleProfileFiles(cacheDir, now))
    } finally {
      cacheDir.deleteRecursively()
    }
  }

  @Test
  fun snapshotsExactlyOneImmutableProviderResponse() {
    val firstVersion = byteArrayOf(1, 2, 3, 4)
    val laterVersion = byteArrayOf(9, 9, 9, 9)
    var opens = 0
    val provider = {
      opens += 1
      ByteArrayInputStream(if (opens == 1) firstVersion else laterVersion)
    }
    val target = File.createTempFile("profile-source-test-", ".bin")
    try {
      ProfileImageProcessorModule().snapshotProfileSource(provider(), target)
      assertArrayEquals(firstVersion, target.readBytes())
      assertEquals(1, opens)
    } finally {
      target.delete()
    }
  }

  @Test
  fun rejectsEmptyProviderResponse() {
    val target = File.createTempFile("profile-source-test-", ".bin")
    try {
      assertThrows(IllegalArgumentException::class.java) {
        ProfileImageProcessorModule().snapshotProfileSource(ByteArrayInputStream(byteArrayOf()), target)
      }
    } finally {
      target.delete()
    }
  }

  @Test
  fun enforcesExactStreamingSourceByteLimit() {
    val exact = File.createTempFile("profile-source-exact-", ".bin")
    val oversized = File.createTempFile("profile-source-oversized-", ".bin")
    try {
      ProfileImageProcessorModule().snapshotProfileSource(
        RepeatingInputStream(25L * 1024L * 1024L),
        exact,
      )
      assertEquals(25L * 1024L * 1024L, exact.length())
      assertThrows(IllegalArgumentException::class.java) {
        ProfileImageProcessorModule().snapshotProfileSource(
          RepeatingInputStream(25L * 1024L * 1024L + 1L),
          oversized,
        )
      }
    } finally {
      exact.delete()
      oversized.delete()
    }
  }

  @Test
  fun preparesBoundedBaselineJpegFromLargeFixture() {
    withFixture(width = 3000, height = 1000) { source, cacheDir ->
      val result = ProfileImageProcessorModule().prepareSnapshot(source, cacheDir)
      assertEquals(3000, result["sourceWidth"])
      assertEquals(1000, result["sourceHeight"])
      assertTrue((result["width"] as Int) in 1..2048)
      assertTrue((result["height"] as Int) in 1..2048)
      val output = outputFile(result)
      try {
        assertTrue(output.length() in 1..(5L * 1024L * 1024L))
        val markers = jpegFrameMarkers(output.readBytes())
        assertTrue("output must contain baseline SOF0", markers.contains(0xc0))
        assertFalse("output must not contain progressive SOF2", markers.contains(0xc2))
        assertTrue("output must not retain JPEG metadata", jpegMetadataMarkers(output.readBytes()).isEmpty())
        System.getenv("PROFILE_IMAGE_ANDROID_FIXTURE")?.let { compatibilityPath ->
          output.copyTo(File(compatibilityPath), overwrite = true)
        }
      } finally {
        output.delete()
      }
    }
  }

  @Test
  fun appliesExifOrientationBeforeReportingOutputDimensions() {
    withFixture(width = 120, height = 60, orientation = ExifInterface.ORIENTATION_ROTATE_90) { source, cacheDir ->
      val result = ProfileImageProcessorModule().prepareSnapshot(source, cacheDir)
      assertEquals(60, result["width"])
      assertEquals(120, result["height"])
      outputFile(result).delete()
    }
  }

  @Test
  fun rejectsMalformedFixture() {
    val cacheDir = createTempDirectory("profile-image-cache-").toFile()
    try {
      val malformed = File(cacheDir, "malformed.jpg").apply { writeText("not an image") }
      assertThrows(IllegalArgumentException::class.java) {
        ProfileImageProcessorModule().prepareSnapshot(malformed, cacheDir)
      }
    } finally {
      cacheDir.deleteRecursively()
    }
  }

  @Test
  fun rejectsOversizedDimensionFixture() {
    val cacheDir = createTempDirectory("profile-image-cache-").toFile()
    try {
      val oversized = File(cacheDir, "oversized.jpg")
      writeJpegFixture(oversized, 10, 10)
      patchJpegWidth(oversized, 20_001)
      assertThrows(IllegalArgumentException::class.java) {
        ProfileImageProcessorModule().prepareSnapshot(oversized, cacheDir)
      }
    } finally {
      cacheDir.deleteRecursively()
    }
  }

  @Test
  fun rejectsSourcePixelProductAboveBudget() {
    val cacheDir = createTempDirectory("profile-image-cache-").toFile()
    try {
      val oversized = File(cacheDir, "oversized-pixels.jpg")
      writeJpegFixture(oversized, 10, 10)
      patchJpegDimensions(oversized, width = 20_000, height = 6_000)
      assertThrows(IllegalArgumentException::class.java) {
        ProfileImageProcessorModule().prepareSnapshot(oversized, cacheDir)
      }
    } finally {
      cacheDir.deleteRecursively()
    }
  }

  @Test
  fun rejectsSourcePixelProductBeforeRasterAllocation() {
    assertThrows(IllegalArgumentException::class.java) {
      ProfileImageProcessorModule().validateSourceDimensions(20_000, 6_000)
    }
  }

  @Test
  fun preparesPngFixtureAsBoundedBaselineJpeg() {
    val cacheDir = createTempDirectory("profile-image-cache-").toFile()
    val source = File(cacheDir, "source.png")
    try {
      val bitmap = Bitmap.createBitmap(320, 180, Bitmap.Config.ARGB_8888)
      bitmap.eraseColor(Color.rgb(180, 80, 30))
      FileOutputStream(source).use { output ->
        assertTrue(bitmap.compress(Bitmap.CompressFormat.PNG, 100, output))
      }
      bitmap.recycle()

      val result = ProfileImageProcessorModule().prepareSnapshot(source, cacheDir)
      assertEquals(320, result["width"])
      assertEquals(180, result["height"])
      val output = outputFile(result)
      try {
        val markers = jpegFrameMarkers(output.readBytes())
        assertTrue(markers.contains(0xc0))
        assertFalse(markers.contains(0xc2))
      } finally {
        output.delete()
      }
    } finally {
      cacheDir.deleteRecursively()
    }
  }

  private fun withFixture(
    width: Int,
    height: Int,
    orientation: Int? = null,
    assertion: (File, File) -> Unit,
  ) {
    val cacheDir = createTempDirectory("profile-image-cache-").toFile()
    val source = File(cacheDir, "source.jpg")
    try {
      writeJpegFixture(source, width, height)
      if (orientation != null) {
        ExifInterface(source).apply {
          setAttribute(ExifInterface.TAG_ORIENTATION, orientation.toString())
          saveAttributes()
        }
      }
      assertion(source, cacheDir)
    } finally {
      cacheDir.deleteRecursively()
    }
  }

  private fun writeJpegFixture(target: File, width: Int, height: Int) {
    val bitmap = Bitmap.createBitmap(width, height, Bitmap.Config.ARGB_8888)
    bitmap.eraseColor(Color.rgb(40, 90, 180))
    FileOutputStream(target).use { output ->
      assertTrue(bitmap.compress(Bitmap.CompressFormat.JPEG, 90, output))
    }
    bitmap.recycle()
  }

  private fun outputFile(result: Map<String, Any>): File =
    File(requireNotNull(Uri.parse(result["uri"] as String).path))

  private fun patchJpegWidth(target: File, width: Int) {
    patchJpegDimensions(target, width = width, height = null)
  }

  private fun patchJpegDimensions(target: File, width: Int, height: Int?) {
    val bytes = target.readBytes()
    var offset = 2
    while (offset + 8 < bytes.size) {
      if (bytes[offset].toInt() and 0xff != 0xff) break
      val marker = bytes[offset + 1].toInt() and 0xff
      val length = ((bytes[offset + 2].toInt() and 0xff) shl 8) or (bytes[offset + 3].toInt() and 0xff)
      if (marker == 0xc0) {
        if (height != null) {
          bytes[offset + 5] = (height ushr 8).toByte()
          bytes[offset + 6] = height.toByte()
        }
        bytes[offset + 7] = (width ushr 8).toByte()
        bytes[offset + 8] = width.toByte()
        target.writeBytes(bytes)
        return
      }
      offset += 2 + length
    }
    throw AssertionError("fixture has no SOF0 marker")
  }

  private class RepeatingInputStream(private var remaining: Long) : InputStream() {
    override fun read(): Int {
      if (remaining <= 0) return -1
      remaining -= 1
      return 0x5a
    }

    override fun read(buffer: ByteArray, offset: Int, length: Int): Int {
      if (remaining <= 0) return -1
      val count = minOf(length.toLong(), remaining).toInt()
      buffer.fill(0x5a, offset, offset + count)
      remaining -= count
      return count
    }
  }

  private fun jpegFrameMarkers(bytes: ByteArray): Set<Int> {
    val result = mutableSetOf<Int>()
    var index = 0
    while (index + 1 < bytes.size) {
      if (bytes[index].toInt() and 0xff == 0xff) {
        val marker = bytes[index + 1].toInt() and 0xff
        if (marker in 0xc0..0xcf && marker !in setOf(0xc4, 0xc8, 0xcc)) result += marker
      }
      index += 1
    }
    return result
  }

  private fun jpegMetadataMarkers(bytes: ByteArray): Set<Int> {
    val result = mutableSetOf<Int>()
    var index = 0
    while (index + 1 < bytes.size) {
      if (bytes[index].toInt() and 0xff == 0xff) {
        val marker = bytes[index + 1].toInt() and 0xff
        if (marker in 0xe0..0xef || marker == 0xfe) result += marker
      }
      index++
    }
    return result
  }
}
