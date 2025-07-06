using Microsoft.Extensions.Logging;
using NSubstitute;
using NSubstitute.ExceptionExtensions;
using SqlMigration.Models;
using SqlMigration.Repositories;
using SqlMigration.Services;

namespace SqlMigration.Tests.Services;

public class MigrationServiceTests
{
    private readonly ISqlScriptsHelpers _sqlScriptsHelpersMock = Substitute.For<ISqlScriptsHelpers>();
    private readonly IMigrationHistoryRepositoryFactory _migrationHistoryRepositoryFactoryMock = Substitute.For<IMigrationHistoryRepositoryFactory>();
    private readonly IMigrationHistoryRepository _migrationHistoryRepositoryMock = Substitute.For<IMigrationHistoryRepository>();
    private readonly ILogger<MigrationService> _loggerMock = Substitute.For<ILogger<MigrationService>>();

    private MigrationService _migrationService;

    public MigrationServiceTests()
    {
        _migrationHistoryRepositoryFactoryMock
            .Create(Arg.Any<ConnectionParameters>())
            .Returns(_migrationHistoryRepositoryMock);

        _migrationService = new MigrationService(
            _sqlScriptsHelpersMock,
            _migrationHistoryRepositoryFactoryMock,
            _loggerMock);
    }

    [Fact]
    public async Task RunAsync_ShouldCreateHistoryTable_WhenItDoesNotExist()
    {
        // Arrange
        var connectionParams = new ConnectionParameters("SQLite", "DataSource=:memory:", null, null, null, null, null);
        _migrationHistoryRepositoryMock.IsHistoryTableCreatedAsync().Returns(false);
        _sqlScriptsHelpersMock.ScanForSqlFiles(Arg.Any<string>()).Returns(new List<string>());

        // Act
        await _migrationService.RunAsync("scripts", connectionParams);

        // Assert
        await _migrationHistoryRepositoryMock.Received(1).CreateHistoryTableAsync();
    }

    [Fact]
    public async Task RunAsync_ShouldNotCreateHistoryTable_WhenItExists()
    {
        // Arrange
        var connectionParams = new ConnectionParameters("SQLite", "DataSource=:memory:", null, null, null, null, null);
        _migrationHistoryRepositoryMock.IsHistoryTableCreatedAsync().Returns(true);
        _sqlScriptsHelpersMock.ScanForSqlFiles(Arg.Any<string>()).Returns(new List<string>());

        // Act
        await _migrationService.RunAsync("scripts", connectionParams);

        // Assert
        await _migrationHistoryRepositoryMock.DidNotReceive().CreateHistoryTableAsync();
    }

    [Fact]
    public async Task RunAsync_ShouldExecuteNewScripts()
    {
        // Arrange
        var connectionParams = new ConnectionParameters("SQLite", "DataSource=:memory:", null, null, null, null, null);
        const string scriptContent = "SELECT 1";
        const string scriptHash = "hash1";

        _migrationHistoryRepositoryMock.IsHistoryTableCreatedAsync().Returns(true);
        _migrationHistoryRepositoryMock.GetExecutedScriptsAsync().Returns(new List<ScriptExecutionHistory>());

        var tempFile = Path.GetTempFileName();
        await File.WriteAllTextAsync(tempFile, scriptContent);
        _sqlScriptsHelpersMock.ScanForSqlFiles(Arg.Any<string>()).Returns(new List<string> { tempFile });
        _sqlScriptsHelpersMock.CalculateHash(tempFile).Returns(scriptHash);
        _sqlScriptsHelpersMock.ExecuteScriptAsync(scriptContent, connectionParams).Returns(Task.CompletedTask);

        // Act
        await _migrationService.RunAsync(Path.GetDirectoryName(tempFile)!, connectionParams);

        // Assert
        await _sqlScriptsHelpersMock.Received(1).ExecuteScriptAsync(scriptContent, connectionParams);
        await _migrationHistoryRepositoryMock.Received(1).UpsertExecutedScriptAsync(Arg.Is<ScriptExecutionHistory>(s => s.ScriptFile == Path.GetFileName(tempFile) && s.ScriptHash == scriptHash));

        File.Delete(tempFile);
    }

    [Fact]
    public async Task RunAsync_ShouldSkipExecutedScripts()
    {
        // Arrange
        var connectionParams = new ConnectionParameters("SQLite", "DataSource=:memory:", null, null, null, null, null);
        const string scriptFile = "script1.sql";
        const string scriptHash = "hash1";

        _migrationHistoryRepositoryMock.IsHistoryTableCreatedAsync().Returns(true);
        _migrationHistoryRepositoryMock.GetExecutedScriptsAsync().Returns(new List<ScriptExecutionHistory>
        {
            new() { ScriptFile = scriptFile, ScriptHash = scriptHash }
        });

        var tempDir = Path.GetTempPath();
        var tempFile = Path.Combine(tempDir, scriptFile);
        await File.WriteAllTextAsync(tempFile, "content");

        _sqlScriptsHelpersMock.ScanForSqlFiles(Arg.Any<string>()).Returns(new List<string> { tempFile });
        _sqlScriptsHelpersMock.CalculateHash(tempFile).Returns(scriptHash);

        // Act
        await _migrationService.RunAsync(tempDir, connectionParams);

        // Assert
        await _sqlScriptsHelpersMock.DidNotReceive().ExecuteScriptAsync(Arg.Any<string>(), Arg.Any<ConnectionParameters>());
        await _migrationHistoryRepositoryMock.DidNotReceive().UpsertExecutedScriptAsync(Arg.Any<ScriptExecutionHistory>());

        File.Delete(tempFile);
    }

    [Fact]
    public async Task RunAsync_ShouldReExecuteScripts_WhenHashChanges()
    {
        // Arrange
        var connectionParams = new ConnectionParameters("SQLite", "DataSource=:memory:", null, null, null, null, null);
        const string scriptFile = "script1.sql";
        const string oldHash = "hash1";
        const string newHash = "hash2";
        const string scriptContent = "SELECT 2";

        _migrationHistoryRepositoryMock.IsHistoryTableCreatedAsync().Returns(true);
        _migrationHistoryRepositoryMock.GetExecutedScriptsAsync().Returns(new List<ScriptExecutionHistory>
        {
            new() { ScriptFile = scriptFile, ScriptHash = oldHash }
        });

        var tempDir = Path.GetTempPath();
        var tempFile = Path.Combine(tempDir, scriptFile);
        await File.WriteAllTextAsync(tempFile, scriptContent);
        _sqlScriptsHelpersMock.ScanForSqlFiles(Arg.Any<string>()).Returns(new List<string> { tempFile });
        _sqlScriptsHelpersMock.CalculateHash(tempFile).Returns(newHash);
        _sqlScriptsHelpersMock.ExecuteScriptAsync(scriptContent, connectionParams).Returns(Task.CompletedTask);

        // Act
        await _migrationService.RunAsync(tempDir, connectionParams);

        // Assert
        await _sqlScriptsHelpersMock.Received(1).ExecuteScriptAsync(scriptContent, connectionParams);
        await _migrationHistoryRepositoryMock.Received(1).UpsertExecutedScriptAsync(Arg.Is<ScriptExecutionHistory>(s => s.ScriptFile == scriptFile && s.ScriptHash == newHash));

        File.Delete(tempFile);
    }

    [Fact]
    public async Task RunAsync_ShouldHandleErrorsGracefully()
    {
        // Arrange
        var connectionParams = new ConnectionParameters("SQLite", "DataSource=:memory:", null, null, null, null, null);
        const string scriptContent1 = "SELECT 1";
        const string scriptContent2 = "SELECT 2";
        const string hash1 = "hash1";
        const string hash2 = "hash2";

        _migrationHistoryRepositoryMock.IsHistoryTableCreatedAsync().Returns(true);
        _migrationHistoryRepositoryMock.GetExecutedScriptsAsync().Returns(new List<ScriptExecutionHistory>());

        var tempDir = Path.GetTempPath();
        var tempFile1 = Path.Combine(tempDir, "script1.sql");
        var tempFile2 = Path.Combine(tempDir, "script2.sql");
        await File.WriteAllTextAsync(tempFile1, scriptContent1);
        await File.WriteAllTextAsync(tempFile2, scriptContent2);

        _sqlScriptsHelpersMock.ScanForSqlFiles(Arg.Any<string>()).Returns(new List<string> { tempFile1, tempFile2 });
        _sqlScriptsHelpersMock.CalculateHash(tempFile1).Returns(hash1);
        _sqlScriptsHelpersMock.CalculateHash(tempFile2).Returns(hash2);
        _sqlScriptsHelpersMock.ExecuteScriptAsync(scriptContent1, connectionParams).ThrowsAsync(new Exception("Test Exception"));
        _sqlScriptsHelpersMock.ExecuteScriptAsync(scriptContent2, connectionParams).Returns(Task.CompletedTask);

        // Act
        await _migrationService.RunAsync(tempDir, connectionParams);

        // Assert
        await _sqlScriptsHelpersMock.Received(1).ExecuteScriptAsync(scriptContent1, connectionParams);
        await _sqlScriptsHelpersMock.Received(1).ExecuteScriptAsync(scriptContent2, connectionParams);
        await _migrationHistoryRepositoryMock.DidNotReceive().UpsertExecutedScriptAsync(Arg.Is<ScriptExecutionHistory>(s => s.ScriptFile == "script1.sql" && s.ScriptHash == hash1));
        await _migrationHistoryRepositoryMock.Received(1).UpsertExecutedScriptAsync(Arg.Is<ScriptExecutionHistory>(s => s.ScriptFile == "script2.sql" && s.ScriptHash == hash2));

        File.Delete(tempFile1);
        File.Delete(tempFile2);
    }

    [Fact]
    public async Task RunAsync_ShouldPerformDryRunCorrectly()
    {
        // Arrange
        _migrationService = new MigrationService(
            _sqlScriptsHelpersMock,
            _migrationHistoryRepositoryFactoryMock,
            _loggerMock,
            isDryRun: true);

        var connectionParams = new ConnectionParameters("SQLite", "DataSource=:memory:", null, null, null, null, null);
        const string scriptFile = "script1.sql";
        const string scriptHash = "hash1";

        _migrationHistoryRepositoryMock.IsHistoryTableCreatedAsync().Returns(true);
        _migrationHistoryRepositoryMock.GetExecutedScriptsAsync().Returns(new List<ScriptExecutionHistory>());
        _sqlScriptsHelpersMock.ScanForSqlFiles(Arg.Any<string>()).Returns(new List<string> { scriptFile });
        _sqlScriptsHelpersMock.CalculateHash(scriptFile).Returns(scriptHash);

        // Act
        await _migrationService.RunAsync("scripts", connectionParams);

        // Assert
        await _sqlScriptsHelpersMock.DidNotReceive().ExecuteScriptAsync(Arg.Any<string>(), Arg.Any<ConnectionParameters>());
        await _migrationHistoryRepositoryMock.DidNotReceive().UpsertExecutedScriptAsync(Arg.Any<ScriptExecutionHistory>());
    }
}
