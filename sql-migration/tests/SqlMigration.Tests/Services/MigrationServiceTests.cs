using Microsoft.Extensions.Logging;
using NSubstitute;
using SqlMigration.Contracts;
using SqlMigration.Factories;
using SqlMigration.Repositories;
using SqlMigration.Services;

namespace SqlMigration.Tests.Services;

public class MigrationServiceTests
{
    private readonly IFileScanner _fileScanner;
    private readonly IHashCalculator _hashCalculator;
    private readonly IScriptExecutor _scriptExecutor;
    private readonly IMigrationHistoryRepositoryFactory _migrationHistoryRepositoryFactory;
    private readonly IMigrationHistoryRepository _migrationHistoryRepository;
    private readonly ILogger<MigrationService> _logger;
    private readonly MigrationService _migrationService;

    public MigrationServiceTests()
    {
        _fileScanner = Substitute.For<IFileScanner>();
        _hashCalculator = Substitute.For<IHashCalculator>();
        _scriptExecutor = Substitute.For<IScriptExecutor>();
        _migrationHistoryRepositoryFactory = Substitute.For<IMigrationHistoryRepositoryFactory>();
        _migrationHistoryRepository = Substitute.For<IMigrationHistoryRepository>();
        _logger = Substitute.For<ILogger<MigrationService>>();

        _migrationHistoryRepositoryFactory.Create(Arg.Any<string>()).Returns(_migrationHistoryRepository);

        _migrationService = new MigrationService(
            _fileScanner,
            _hashCalculator,
            _scriptExecutor,
            _migrationHistoryRepositoryFactory,
            _logger);
    }

    [Fact]
    public async Task RunAsync_ShouldCreateHistoryTable_WhenItDoesNotExist()
    {
        // Arrange
        _migrationHistoryRepository.IsHistoryTableCreated().Returns(false);

        // Act
        await _migrationService.RunAsync("scripts", "connectionString");

        // Assert
        await _migrationHistoryRepository.Received(1).CreateHistoryTable();
    }

    [Fact]
    public async Task RunAsync_ShouldNotCreateHistoryTable_WhenItAlreadyExists()
    {
        // Arrange
        _migrationHistoryRepository.IsHistoryTableCreated().Returns(true);

        // Act
        await _migrationService.RunAsync("scripts", "connectionString");

        // Assert
        await _migrationHistoryRepository.DidNotReceive().CreateHistoryTable();
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
        _fileScanner.ScanForSqlFiles(tempDir).Returns(scriptFiles);
        _hashCalculator.CalculateHash(scriptFile).Returns("hash1");
        _migrationHistoryRepository.IsHistoryTableCreated().Returns(true);
        _migrationHistoryRepository.GetExecutedScripts().Returns(new List<MigrationHistory>());

        // Act
        await _migrationService.RunAsync(tempDir, "connectionString");

        // Assert
        await _scriptExecutor.Received(1).ExecuteScript("connectionString", "SELECT 1");
        await _migrationHistoryRepository.Received(1).AddExecutedScript(Arg.Is<MigrationHistory>(m => m.ScriptName == "script1.sql" && m.Hash == "hash1"));

        // Cleanup
        Directory.Delete(tempDir, true);
    }

    [Fact]
    public async Task RunAsync_ShouldSkipExecutedScripts_WithSameHash()
    {
        // Arrange
        var scriptFiles = new[] { "scripts/script1.sql" };
        _fileScanner.ScanForSqlFiles("scripts").Returns(scriptFiles);
        _hashCalculator.CalculateHash("scripts/script1.sql").Returns("hash1");
        _migrationHistoryRepository.IsHistoryTableCreated().Returns(true);
        _migrationHistoryRepository.GetExecutedScripts().Returns(new List<MigrationHistory>
        {
            new() { ScriptName = "script1.sql", Hash = "hash1" }
        });

        // Act
        await _migrationService.RunAsync("scripts", "connectionString");

        // Assert
        await _scriptExecutor.DidNotReceive().ExecuteScript(Arg.Any<string>(), Arg.Any<string>());
        await _migrationHistoryRepository.DidNotReceive().AddExecutedScript(Arg.Any<MigrationHistory>());
    }

    [Fact]
    public async Task RunAsync_ShouldFail_WhenExecutedScriptHashChanges()
    {
        // Arrange
        var scriptFiles = new[] { "scripts/script1.sql" };
        _fileScanner.ScanForSqlFiles("scripts").Returns(scriptFiles);
        _hashCalculator.CalculateHash("scripts/script1.sql").Returns("new_hash");
        _migrationHistoryRepository.IsHistoryTableCreated().Returns(true);
        _migrationHistoryRepository.GetExecutedScripts().Returns(new List<MigrationHistory>
        {
            new() { ScriptName = "script1.sql", Hash = "old_hash" }
        });

        // Act
        var result = await _migrationService.RunAsync("scripts", "connectionString");

        // Assert
        Assert.Equal(1, result);
        await _scriptExecutor.DidNotReceive().ExecuteScript(Arg.Any<string>(), Arg.Any<string>());
        await _migrationHistoryRepository.DidNotReceive().AddExecutedScript(Arg.Any<MigrationHistory>());
    }
}
