using NSubstitute;
using SqlMigration.Services;

namespace SqlMigration.Tests;

public class FileScannerTests
{
    [Fact]
    public void ScanForSqlFiles_ShouldReturnAllSqlFilesInDirectoryAndSubdirectories()
    {
        // Arrange
        var fileScanner = new FileScanner();
        var tempDir = Path.Combine(Path.GetTempPath(), Guid.NewGuid().ToString());
        Directory.CreateDirectory(tempDir);
        Directory.CreateDirectory(Path.Combine(tempDir, "sub"));

        File.WriteAllText(Path.Combine(tempDir, "script1.sql"), "");
        File.WriteAllText(Path.Combine(tempDir, "script2.sql"), "");
        File.WriteAllText(Path.Combine(tempDir, "sub", "script3.sql"), "");
        File.WriteAllText(Path.Combine(tempDir, "script4.txt"), "");

        // Act
        var result = fileScanner.ScanForSqlFiles(tempDir);

        // Assert
        Assert.Equal(3, result.Count());
        Assert.Contains(result, f => f.EndsWith("script1.sql"));
        Assert.Contains(result, f => f.EndsWith("script2.sql"));
        Assert.Contains(result, f => f.EndsWith("script3.sql"));

        // Cleanup
        Directory.Delete(tempDir, true);
    }
}
