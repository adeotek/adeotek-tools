using SqlMigration.Services;

namespace SqlMigration.Tests.Services;

public class HashCalculatorTests
{
    [Fact]
    public void CalculateHash_ShouldReturnCorrectSha256Hash()
    {
        // Arrange
        var hashCalculator = new HashCalculator();
        var tempFile = Path.GetTempFileName();
        File.WriteAllText(tempFile, "test content");

        // Act
        var result = hashCalculator.CalculateHash(tempFile);

        // Assert
        Assert.Equal("6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72", result);

        // Cleanup
        File.Delete(tempFile);
    }
}
