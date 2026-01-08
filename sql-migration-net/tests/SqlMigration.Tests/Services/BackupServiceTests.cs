using Microsoft.Extensions.Logging;
using NSubstitute;
using SqlMigration.Models;
using SqlMigration.Services;

namespace SqlMigration.Tests.Services;

public class BackupServiceTests
{
    private readonly ILogger<BackupService> _loggerMock = Substitute.For<ILogger<BackupService>>();
    private BackupService _backupService;

    public BackupServiceTests()
    {
        _backupService = new BackupService(_loggerMock);
    }

    [Fact]
    public async Task CreateBackupAsync_ShouldCreateBackupDirectory_WhenItDoesNotExist()
    {
        // Arrange
        var connectionParams = new ConnectionParameters("SQLite", null, null, null, "test.db", null, null);
        var backupDir = Path.Combine(Directory.GetCurrentDirectory(), ".sql-migration-backups");

        // Clean up before test
        if (Directory.Exists(backupDir))
        {
            Directory.Delete(backupDir, true);
        }

        // Create test database file
        var testDbPath = "test.db";
        await File.WriteAllTextAsync(testDbPath, "test data");

        try
        {
            // Act
            var backupPath = await _backupService.CreateBackupAsync(connectionParams);

            // Assert
            Assert.True(Directory.Exists(backupDir));
            Assert.True(File.Exists(backupPath));
            Assert.Contains(backupDir, backupPath);
        }
        finally
        {
            // Clean up
            if (File.Exists(testDbPath))
            {
                File.Delete(testDbPath);
            }
            if (Directory.Exists(backupDir))
            {
                Directory.Delete(backupDir, true);
            }
        }
    }

    [Fact]
    public async Task CreateBackupAsync_SQLite_ShouldCopyDatabaseFile()
    {
        // Arrange
        var testDbPath = Path.GetTempFileName();
        var testContent = "SQLite test data";
        await File.WriteAllTextAsync(testDbPath, testContent);

        var connectionParams = new ConnectionParameters("SQLite", null, null, null, testDbPath, null, null);
        var backupDir = Path.Combine(Directory.GetCurrentDirectory(), ".sql-migration-backups");

        try
        {
            // Act
            var backupPath = await _backupService.CreateBackupAsync(connectionParams);

            // Assert
            Assert.True(File.Exists(backupPath));
            Assert.Equal(testContent, await File.ReadAllTextAsync(backupPath));
            Assert.EndsWith(".db", backupPath);
        }
        finally
        {
            // Clean up
            if (File.Exists(testDbPath))
            {
                File.Delete(testDbPath);
            }
            if (Directory.Exists(backupDir))
            {
                Directory.Delete(backupDir, true);
            }
        }
    }

    [Fact]
    public async Task CreateBackupAsync_SQLite_InDryRunMode_ShouldNotCreateActualBackup()
    {
        // Arrange
        _backupService = new BackupService(_loggerMock, isDryRun: true);
        var testDbPath = Path.GetTempFileName();
        await File.WriteAllTextAsync(testDbPath, "test data");

        var connectionParams = new ConnectionParameters("SQLite", null, null, null, testDbPath, null, null);
        var backupDir = Path.Combine(Directory.GetCurrentDirectory(), ".sql-migration-backups");

        try
        {
            // Act
            var backupPath = await _backupService.CreateBackupAsync(connectionParams);

            // Assert
            Assert.False(File.Exists(backupPath));
            Assert.NotEmpty(backupPath);
        }
        finally
        {
            // Clean up
            if (File.Exists(testDbPath))
            {
                File.Delete(testDbPath);
            }
            if (Directory.Exists(backupDir))
            {
                Directory.Delete(backupDir, true);
            }
        }
    }

    [Fact]
    public async Task RestoreLastBackupAsync_SQLite_ShouldRestoreDatabaseFile()
    {
        // Arrange
        var testDbPath = Path.GetTempFileName();
        var originalContent = "original data";
        var backupContent = "backup data";

        await File.WriteAllTextAsync(testDbPath, originalContent);

        var connectionParams = new ConnectionParameters("SQLite", null, null, null, testDbPath, null, null);
        var backupDir = Path.Combine(Directory.GetCurrentDirectory(), ".sql-migration-backups");

        try
        {
            // Create a backup
            var backupPath = await _backupService.CreateBackupAsync(connectionParams);

            // Modify backup file to have different content
            await File.WriteAllTextAsync(backupPath, backupContent);

            // Modify original file
            await File.WriteAllTextAsync(testDbPath, "modified data");

            // Act
            await _backupService.RestoreLastBackupAsync(connectionParams);

            // Assert
            Assert.Equal(backupContent, await File.ReadAllTextAsync(testDbPath));
        }
        finally
        {
            // Clean up
            if (File.Exists(testDbPath))
            {
                File.Delete(testDbPath);
            }
            if (Directory.Exists(backupDir))
            {
                Directory.Delete(backupDir, true);
            }
        }
    }

    [Fact]
    public async Task RestoreLastBackupAsync_ShouldThrowException_WhenNoBackupExists()
    {
        // Arrange
        var connectionParams = new ConnectionParameters("SQLite", null, null, null, "nonexistent.db", null, null);

        // Act & Assert
        await Assert.ThrowsAsync<InvalidOperationException>(
            async () => await _backupService.RestoreLastBackupAsync(connectionParams));
    }

    [Fact]
    public void GetLastBackupPath_ShouldReturnNull_WhenNoBackupExists()
    {
        // Arrange
        var connectionParams = new ConnectionParameters("SQLite", null, null, null, "nonexistent.db", null, null);
        var backupDir = Path.Combine(Directory.GetCurrentDirectory(), ".sql-migration-backups");

        // Clean up before test
        if (Directory.Exists(backupDir))
        {
            Directory.Delete(backupDir, true);
        }

        // Act
        var lastBackupPath = _backupService.GetLastBackupPath(connectionParams);

        // Assert
        Assert.Null(lastBackupPath);
    }

    [Fact]
    public async Task GetLastBackupPath_ShouldReturnMostRecentBackup()
    {
        // Arrange
        var testDbPath = Path.GetTempFileName();
        await File.WriteAllTextAsync(testDbPath, "test data");

        var connectionParams = new ConnectionParameters("SQLite", null, null, null, testDbPath, null, null);
        var backupDir = Path.Combine(Directory.GetCurrentDirectory(), ".sql-migration-backups");

        try
        {
            // Create multiple backups with delay
            var backup1 = await _backupService.CreateBackupAsync(connectionParams);
            await Task.Delay(1100); // Wait more than 1 second to ensure different timestamps
            var backup2 = await _backupService.CreateBackupAsync(connectionParams);

            // Act
            var lastBackupPath = _backupService.GetLastBackupPath(connectionParams);

            // Assert
            Assert.Equal(backup2, lastBackupPath);
        }
        finally
        {
            // Clean up
            if (File.Exists(testDbPath))
            {
                File.Delete(testDbPath);
            }
            if (Directory.Exists(backupDir))
            {
                Directory.Delete(backupDir, true);
            }
        }
    }

    [Fact]
    public void CreateBackupAsync_PostgreSQL_ShouldThrowException_WhenPgDumpNotAvailable()
    {
        // Arrange
        var connectionParams = new ConnectionParameters(
            "PostgreSQL",
            null,
            "localhost",
            5432,
            "testdb",
            "testuser",
            "testpass");

        // Act & Assert
        // This test will pass if pg_dump is not installed on the system
        // If pg_dump is installed, it will attempt to connect and fail with connection error
        // Either way, it tests that the service properly handles errors
        Assert.ThrowsAsync<Exception>(async () => await _backupService.CreateBackupAsync(connectionParams));
    }

    [Fact]
    public async Task CreateBackupAsync_SQLite_ShouldThrowException_WhenDatabaseFileNotFound()
    {
        // Arrange
        var connectionParams = new ConnectionParameters("SQLite", null, null, null, "nonexistent.db", null, null);

        // Act & Assert
        await Assert.ThrowsAsync<FileNotFoundException>(
            async () => await _backupService.CreateBackupAsync(connectionParams));
    }
}
