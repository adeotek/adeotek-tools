using SqlMigration.Services;

namespace SqlMigration.Tests;

public class SqlScriptsHelpersTests
{
    [Fact]
    public void ScanForSqlFiles_ShouldReturnAllSqlFilesInDirectoryAndSubdirectories()
    {
        // Arrange
        var sqlScriptsHelpers = new SqlScriptsHelpers();
        var tempDir = Path.Combine(Path.GetTempPath(), Guid.NewGuid().ToString());
        Directory.CreateDirectory(tempDir);
        Directory.CreateDirectory(Path.Combine(tempDir, "sub"));

        File.WriteAllText(Path.Combine(tempDir, "script1.sql"), "");
        File.WriteAllText(Path.Combine(tempDir, "script2.sql"), "");
        File.WriteAllText(Path.Combine(tempDir, "sub", "script3.sql"), "");
        File.WriteAllText(Path.Combine(tempDir, "script4.txt"), "");

        // Act
        var result = sqlScriptsHelpers.ScanForSqlFiles(tempDir);

        // Assert
        Assert.Equal(3, result.Count());
        Assert.Contains(result, f => f.EndsWith("script1.sql"));
        Assert.Contains(result, f => f.EndsWith("script2.sql"));
        Assert.Contains(result, f => f.EndsWith("script3.sql"));

        // Cleanup
        Directory.Delete(tempDir, true);
    }

    [Fact]
    public void CalculateHash_ShouldReturnCorrectSha256Hash()
    {
        // Arrange
        var sqlScriptsHelpers = new SqlScriptsHelpers();
        var tempFile = Path.GetTempFileName();
        File.WriteAllText(tempFile, "test content");

        // Act
        var result = sqlScriptsHelpers.CalculateHash(tempFile);

        // Assert
        Assert.Equal("6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72", result);

        // Cleanup
        File.Delete(tempFile);
    }
}
