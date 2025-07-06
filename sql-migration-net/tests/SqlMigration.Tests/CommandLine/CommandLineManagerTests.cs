
using SqlMigration.CommandLine;

namespace SqlMigration.Tests.CommandLine;

public class CommandLineManagerTests
{
    [Fact]
    public async Task ExecuteCommandAsync_ShouldProcessCommandLineArguments()
    {
        // Arrange
        var tempDir = Path.Combine(Path.GetTempPath(), Guid.NewGuid().ToString());
        Directory.CreateDirectory(tempDir);
        await File.WriteAllTextAsync(Path.Combine(tempDir, "script.sql"), "SELECT 1");

        var args = new[]
        {
            "--target-path", tempDir,
            "--provider", "SQLite",
            "--connection-string", "DataSource=:memory:"
        };

        // Act
        var result = await CommandLineManager.ExecuteCommandAsync(args);

        // Assert
        Assert.Equal(0, result);

        Directory.Delete(tempDir, true);
    }

    [Fact]
    public async Task ExecuteCommandAsync_ShouldProcessEnvironmentVariables()
    {
        // Arrange
        var tempDir = Path.Combine(Path.GetTempPath(), Guid.NewGuid().ToString());
        Directory.CreateDirectory(tempDir);
        await File.WriteAllTextAsync(Path.Combine(tempDir, "script.sql"), "SELECT 1");

        Environment.SetEnvironmentVariable($"{CommandLineManager.EnvironmentVariablesPrefix}TARGET_PATH", tempDir);
        Environment.SetEnvironmentVariable($"{CommandLineManager.EnvironmentVariablesPrefix}PROVIDER", "SQLite");
        Environment.SetEnvironmentVariable($"{CommandLineManager.EnvironmentVariablesPrefix}CONNECTION_STRING", "DataSource=:memory:");

        // Act
        var result = await CommandLineManager.ExecuteCommandAsync([]);

        // Assert
        Assert.Equal(0, result);

        Directory.Delete(tempDir, true);
        Environment.SetEnvironmentVariable($"{CommandLineManager.EnvironmentVariablesPrefix}TARGET_PATH", null);
        Environment.SetEnvironmentVariable($"{CommandLineManager.EnvironmentVariablesPrefix}PROVIDER", null);
        Environment.SetEnvironmentVariable($"{CommandLineManager.EnvironmentVariablesPrefix}CONNECTION_STRING", null);
    }
}
