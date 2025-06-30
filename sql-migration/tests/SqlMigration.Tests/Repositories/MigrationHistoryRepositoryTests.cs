using NSubstitute;
using SqlMigration.Models;
using SqlMigration.Repositories;

namespace SqlMigration.Tests.Repositories;

public class MigrationHistoryRepositoryTests
{
    private readonly ISqlRepository _sqlRepositoryMock = Substitute.For<ISqlRepository>();
    private readonly MigrationHistoryRepository _repository;

    public MigrationHistoryRepositoryTests()
    {
        _repository = new MigrationHistoryRepository(_sqlRepositoryMock);
    }

    [Fact]
    public async Task IsHistoryTableCreatedAsync_ShouldReturnFalse_WhenTableDoesNotExist()
    {
        // Arrange
        _sqlRepositoryMock.ExecuteScalarAsync<int>(Arg.Any<string>())
            .Returns(0);

        // Act
        var result = await _repository.IsHistoryTableCreatedAsync();

        // Assert
        Assert.False(result);
    }

    [Fact]
    public async Task IsHistoryTableCreatedAsync_ShouldReturnTrue_WhenTableExists()
    {
        // Arrange
        _sqlRepositoryMock.ExecuteScalarAsync<int>(Arg.Any<string>()).Returns(1);

        // Act
        var result = await _repository.IsHistoryTableCreatedAsync();

        // Assert
        Assert.True(result);
    }

    [Fact]
    public async Task GetExecutedScriptsAsync_ShouldReturnEmptyList_WhenNoScriptsExecuted()
    {
        // Arrange
        _sqlRepositoryMock.QueryAsync<ScriptExecutionHistory>(Arg.Any<string>())
            .Returns(Task.FromResult(Enumerable.Empty<ScriptExecutionHistory>()));

        // Act
        var result = await _repository.GetExecutedScriptsAsync();

        // Assert
        Assert.Empty(result);
    }

    [Fact]
    public async Task GetExecutedScriptsAsync_ShouldReturnExecutedScripts()
    {
        // Arrange
        var script = new ScriptExecutionHistory
        {
            ScriptFile = "script1.sql",
            ScriptHash = "hash1",
            ExecutedAt = DateTime.UtcNow
        };
        _sqlRepositoryMock.QueryAsync<ScriptExecutionHistory>(Arg.Any<string>())
            .Returns(Task.FromResult(new[] { script }.AsEnumerable()));

        // Act
        var result = (await _repository.GetExecutedScriptsAsync()).ToList();

        // Assert
        Assert.Single(result);
        Assert.Equal(script.ScriptFile, result[0].ScriptFile);
        Assert.Equal(script.ScriptHash, result[0].ScriptHash);
    }

    [Fact]
    public async Task UpsertExecutedScriptAsync_ShouldInsertNewScript()
    {
        // Arrange
        var script = new ScriptExecutionHistory
        {
            ScriptFile = "script1.sql",
            ScriptHash = "hash1",
            ExecutedAt = DateTime.UtcNow
        };
        _sqlRepositoryMock.QueryFirstOrDefaultAsync<ScriptExecutionHistory>(Arg.Any<string>(), Arg.Any<object>())
            .Returns(Task.FromResult<ScriptExecutionHistory?>(null));

        // Act
        await _repository.UpsertExecutedScriptAsync(script);

        // Assert
        await _sqlRepositoryMock.Received(1).ExecuteAsync(Arg.Is<string>(s => s.Contains("INSERT")), Arg.Any<object>());
    }

    [Fact]
    public async Task UpsertExecutedScriptAsync_ShouldUpdateExistingScript()
    {
        // Arrange
        var script = new ScriptExecutionHistory
        {
            ScriptFile = "script1.sql",
            ScriptHash = "hash1",
            ExecutedAt = DateTime.UtcNow
        };
        _sqlRepositoryMock.QueryFirstOrDefaultAsync<ScriptExecutionHistory>(Arg.Any<string>(), Arg.Any<object>())
            .Returns(Task.FromResult<ScriptExecutionHistory?>(script));

        // Act
        await _repository.UpsertExecutedScriptAsync(script);

        // Assert
        await _sqlRepositoryMock.Received(1).ExecuteAsync(Arg.Is<string>(s => s.Contains("UPDATE")), Arg.Any<object>());
    }
}
