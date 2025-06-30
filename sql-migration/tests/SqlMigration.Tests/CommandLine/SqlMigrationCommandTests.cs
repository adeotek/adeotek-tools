
using System.CommandLine;
using SqlMigration.CommandLine;

namespace SqlMigration.Tests.CommandLine;

public class SqlMigrationCommandTests
{
    [Fact]
    public async Task ExecuteAsync_ShouldReturnError_WhenTargetPathIsInvalid()
    {
        // Arrange
        var command = new RootCommand();
        SqlMigrationCommand.CommandOptions.ForEach(command.Add);
        var parseResult = command.Parse("--target-path invalid/path");

        // Act
        var result = await SqlMigrationCommand.ExecuteAsync(parseResult, CancellationToken.None);

        // Assert
        Assert.Equal(10, result);
    }

    [Fact]
    public async Task ExecuteAsync_ShouldReturnError_WhenConnectionParametersAreInvalid()
    {
        // Arrange
        var tempDir = Path.Combine(Path.GetTempPath(), Guid.NewGuid().ToString());
        Directory.CreateDirectory(tempDir);
        await File.WriteAllTextAsync(Path.Combine(tempDir, "script.sql"), "SELECT 1");

        var command = new RootCommand();
        SqlMigrationCommand.CommandOptions.ForEach(command.Add);
        var parseResult = command.Parse($"--target-path {tempDir}");

        // Act
        var result = await SqlMigrationCommand.ExecuteAsync(parseResult, CancellationToken.None);

        // Assert
        Assert.Equal(20, result);

        Directory.Delete(tempDir, true);
    }
}
