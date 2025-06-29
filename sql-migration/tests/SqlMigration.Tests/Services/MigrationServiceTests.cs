using Microsoft.Extensions.Logging;
using NSubstitute;
using SqlMigration.Models;
using SqlMigration.Repositories;
using SqlMigration.Services;

namespace SqlMigration.Tests.Services;

public class MigrationServiceTests
{
    private readonly ISqlScriptsHelpers _sqlScriptsHelpersMock = Substitute.For<ISqlScriptsHelpers>();
    private readonly IScriptExecutor _scriptExecutorMock = Substitute.For<IScriptExecutor>();
    private readonly IMigrationHistoryRepositoryFactory _migrationHistoryRepositoryFactoryMock = Substitute.For<IMigrationHistoryRepositoryFactory>();
    private readonly IMigrationHistoryRepository _migrationHistoryRepositoryMock = Substitute.For<IMigrationHistoryRepository>();
    private readonly ILogger<MigrationService> _loggerMock = Substitute.For<ILogger<MigrationService>>();

    private readonly MigrationService _migrationService;

    public MigrationServiceTests()
    {
        _migrationHistoryRepositoryFactoryMock
            .Create(Arg.Any<string>()).Returns(_migrationHistoryRepositoryMock);

        _migrationService = new MigrationService(
            _sqlScriptsHelpersMock,
            _scriptExecutorMock,
            _migrationHistoryRepositoryFactoryMock,
            _loggerMock);
    }

    [Fact]
    public async Task RunAsync_ShouldCreateHistoryTable_WhenItDoesNotExist()
    {
        // Arrange
        _migrationHistoryRepositoryMock.IsHistoryTableCreatedAsync().Returns(false);

        // Act
        await _migrationService.RunAsync("scripts", "connectionString");

        // Assert
        await _migrationHistoryRepositoryMock.Received(1).CreateHistoryTableAsync();
    }

    [Fact]
    public async Task RunAsync_ShouldNotCreateHistoryTable_WhenItAlreadyExists()
    {
        // Arrange
        _migrationHistoryRepositoryMock.IsHistoryTableCreatedAsync().Returns(true);

        // Act
        await _migrationService.RunAsync("scripts", "connectionString");

        // Assert
        await _migrationHistoryRepositoryMock.DidNotReceive().CreateHistoryTableAsync();
    }

    [Fact]
    public async Task RunAsync_ShouldExecuteNewScripts()
    {
        // Arrange
        var tempDir = Path.Combine(Path.GetTempPath(), Guid.NewGuid().ToString());
        Directory.CreateDirectory(tempDir);
        var scriptFile = Path.Combine(tempDir, "script1.sql");
        File.WriteAllText(scriptFile, "SELECT 1");

        var scriptFiles = new[] { scriptFile };
        _sqlScriptsHelpersMock.ScanForSqlFiles(tempDir).Returns(scriptFiles);
        _sqlScriptsHelpersMock.CalculateHash(scriptFile).Returns("hash1");
        _migrationHistoryRepositoryMock.IsHistoryTableCreatedAsync().Returns(true);
        _migrationHistoryRepositoryMock.GetExecutedScriptsAsync().Returns(new List<ScriptExecutionHistory>());

        // Act
        await _migrationService.RunAsync(tempDir, "connectionString");

        // Assert
        await _scriptExecutorMock.Received(1).ExecuteScript("connectionString", "SELECT 1");
        await _migrationHistoryRepositoryMock.Received(1).AddExecutedScript(Arg.Is<ScriptExecutionHistory>(m => m.ScriptName == "script1.sql" && m.ScriptHash == "hash1"));

        // Cleanup
        Directory.Delete(tempDir, true);
    }

    [Fact]
    public async Task RunAsync_ShouldSkipExecutedScripts_WithSameHash()
    {
        // Arrange
        var scriptFiles = new[] { "scripts/script1.sql" };
        _sqlScriptsHelpersMock.ScanForSqlFiles("scripts").Returns(scriptFiles);
        _sqlScriptsHelpersMock.CalculateHash("scripts/script1.sql").Returns("hash1");
        _migrationHistoryRepositoryMock.IsHistoryTableCreatedAsync().Returns(true);
        _migrationHistoryRepositoryMock.GetExecutedScriptsAsync().Returns(new List<ScriptExecutionHistory>
        {
            new() { ScriptName = "script1.sql", ScriptHash = "hash1" }
        });

        // Act
        await _migrationService.RunAsync("scripts", "connectionString");

        // Assert
        await _scriptExecutorMock.DidNotReceive().ExecuteScript(Arg.Any<string>(), Arg.Any<string>());
        await _migrationHistoryRepositoryMock.DidNotReceive().AddExecutedScript(Arg.Any<ScriptExecutionHistory>());
    }

    [Fact]
    public async Task RunAsync_ShouldFail_WhenExecutedScriptHashChanges()
    {
        // Arrange
        var scriptFiles = new[] { "scripts/script1.sql" };
        _sqlScriptsHelpersMock.ScanForSqlFiles("scripts").Returns(scriptFiles);
        _sqlScriptsHelpersMock.CalculateHash("scripts/script1.sql").Returns("new_hash");
        _migrationHistoryRepositoryMock.IsHistoryTableCreatedAsync().Returns(true);
        _migrationHistoryRepositoryMock.GetExecutedScriptsAsync().Returns(new List<ScriptExecutionHistory>
        {
            new() { ScriptName = "script1.sql", ScriptHash = "old_hash" }
        });

        // Act
        var result = await _migrationService.RunAsync("scripts", "connectionString");

        // Assert
        Assert.Equal(1, result);
        await _scriptExecutorMock.DidNotReceive().ExecuteScript(Arg.Any<string>(), Arg.Any<string>());
        await _migrationHistoryRepositoryMock.DidNotReceive().AddExecutedScript(Arg.Any<ScriptExecutionHistory>());
    }
}
