using System.Diagnostics;
using Microsoft.Extensions.Logging;
using SqlMigration.Models;

namespace SqlMigration.Services;

public class BackupService(ILogger<BackupService> logger, bool isDryRun = false) : IBackupService
{
    private const string BackupDirectoryName = ".sql-migration-backups";

    public async Task<string> CreateBackupAsync(ConnectionParameters connectionParameters, CancellationToken ct = default)
    {
        var backupDir = GetBackupDirectory(connectionParameters);
        if (!isDryRun && !Directory.Exists(backupDir))
        {
            Directory.CreateDirectory(backupDir);
            logger.LogDebug("Created backup directory: {BackupDir}", backupDir);
        }

        var timestamp = DateTime.UtcNow.ToString("yyyyMMdd_HHmmss");
        var backupFileName = $"{GetDatabaseName(connectionParameters)}_backup_{timestamp}";

        return connectionParameters.DbProvider switch
        {
            DatabaseProvider.PostgreSql => await CreatePostgresBackupAsync(connectionParameters, backupDir, backupFileName, ct),
            DatabaseProvider.SqLite => await CreateSqliteBackupAsync(connectionParameters, backupDir, backupFileName, ct),
            _ => throw new NotSupportedException($"Backup not supported for provider: {connectionParameters.DbProvider}")
        };
    }

    public async Task RestoreLastBackupAsync(ConnectionParameters connectionParameters, CancellationToken ct = default)
    {
        var lastBackupPath = GetLastBackupPath(connectionParameters);
        if (string.IsNullOrEmpty(lastBackupPath))
        {
            throw new InvalidOperationException("No backup found to restore.");
        }

        logger.LogInformation("Restoring backup from: {BackupPath}", lastBackupPath);

        if (isDryRun)
        {
            logger.LogInformation("Dry run: would restore backup from {BackupPath}", lastBackupPath);
            return;
        }

        switch (connectionParameters.DbProvider)
        {
            case DatabaseProvider.PostgreSql:
                await RestorePostgresBackupAsync(connectionParameters, lastBackupPath, ct);
                break;
            case DatabaseProvider.SqLite:
                await RestoreSqliteBackupAsync(connectionParameters, lastBackupPath, ct);
                break;
            default:
                throw new NotSupportedException($"Restore not supported for provider: {connectionParameters.DbProvider}");
        }

        logger.LogInformation("Backup restored successfully.");
    }

    public string? GetLastBackupPath(ConnectionParameters connectionParameters)
    {
        var backupDir = GetBackupDirectory(connectionParameters);
        if (!Directory.Exists(backupDir))
        {
            return null;
        }

        var extension = connectionParameters.DbProvider == DatabaseProvider.PostgreSql ? ".sql" : ".db";
        var backupFiles = Directory.GetFiles(backupDir, $"*{extension}")
            .OrderByDescending(f => File.GetCreationTimeUtc(f))
            .ToArray();

        return backupFiles.Length > 0 ? backupFiles[0] : null;
    }

    private async Task<string> CreatePostgresBackupAsync(
        ConnectionParameters connectionParameters,
        string backupDir,
        string backupFileName,
        CancellationToken ct)
    {
        var backupPath = Path.Combine(backupDir, $"{backupFileName}.sql");
        logger.LogInformation("Creating PostgreSQL backup: {BackupPath}", backupPath);

        if (isDryRun)
        {
            logger.LogInformation("Dry run: would create PostgreSQL backup at {BackupPath}", backupPath);
            return backupPath;
        }

        var pgDumpArgs = $"--host={connectionParameters.Host} " +
                         $"--port={connectionParameters.Port} " +
                         $"--username={connectionParameters.User} " +
                         $"--dbname={connectionParameters.DatabaseName} " +
                         $"--file=\"{backupPath}\" " +
                         "--format=plain --no-owner --no-acl --clean --if-exists";

        var startInfo = new ProcessStartInfo
        {
            FileName = "pg_dump",
            Arguments = pgDumpArgs,
            RedirectStandardOutput = true,
            RedirectStandardError = true,
            UseShellExecute = false,
            CreateNoWindow = true
        };

        if (!string.IsNullOrEmpty(connectionParameters.Password))
        {
            startInfo.Environment["PGPASSWORD"] = connectionParameters.Password;
        }

        using var process = Process.Start(startInfo);
        if (process == null)
        {
            throw new InvalidOperationException("Failed to start pg_dump process.");
        }

        await process.WaitForExitAsync(ct);

        if (process.ExitCode != 0)
        {
            var error = await process.StandardError.ReadToEndAsync(ct);
            throw new InvalidOperationException($"pg_dump failed with exit code {process.ExitCode}: {error}");
        }

        logger.LogInformation("PostgreSQL backup created successfully: {BackupPath}", backupPath);
        return backupPath;
    }

    private async Task<string> CreateSqliteBackupAsync(
        ConnectionParameters connectionParameters,
        string backupDir,
        string backupFileName,
        CancellationToken ct)
    {
        var sourcePath = connectionParameters.DatabaseName!;
        var backupPath = Path.Combine(backupDir, $"{backupFileName}.db");
        logger.LogInformation("Creating SQLite backup: {BackupPath}", backupPath);

        if (isDryRun)
        {
            logger.LogInformation("Dry run: would create SQLite backup at {BackupPath}", backupPath);
            return backupPath;
        }

        if (!File.Exists(sourcePath))
        {
            throw new FileNotFoundException($"SQLite database file not found: {sourcePath}");
        }

        await Task.Run(() => File.Copy(sourcePath, backupPath, overwrite: true), ct);

        logger.LogInformation("SQLite backup created successfully: {BackupPath}", backupPath);
        return backupPath;
    }

    private async Task RestorePostgresBackupAsync(
        ConnectionParameters connectionParameters,
        string backupPath,
        CancellationToken ct)
    {
        logger.LogInformation("Restoring PostgreSQL backup from: {BackupPath}", backupPath);

        var psqlArgs = $"--host={connectionParameters.Host} " +
                       $"--port={connectionParameters.Port} " +
                       $"--username={connectionParameters.User} " +
                       $"--dbname={connectionParameters.DatabaseName} " +
                       $"--file=\"{backupPath}\"";

        var startInfo = new ProcessStartInfo
        {
            FileName = "psql",
            Arguments = psqlArgs,
            RedirectStandardOutput = true,
            RedirectStandardError = true,
            UseShellExecute = false,
            CreateNoWindow = true
        };

        if (!string.IsNullOrEmpty(connectionParameters.Password))
        {
            startInfo.Environment["PGPASSWORD"] = connectionParameters.Password;
        }

        using var process = Process.Start(startInfo);
        if (process == null)
        {
            throw new InvalidOperationException("Failed to start psql process.");
        }

        await process.WaitForExitAsync(ct);

        if (process.ExitCode != 0)
        {
            var error = await process.StandardError.ReadToEndAsync(ct);
            throw new InvalidOperationException($"psql restore failed with exit code {process.ExitCode}: {error}");
        }

        logger.LogInformation("PostgreSQL backup restored successfully.");
    }

    private async Task RestoreSqliteBackupAsync(
        ConnectionParameters connectionParameters,
        string backupPath,
        CancellationToken ct)
    {
        var targetPath = connectionParameters.DatabaseName!;
        logger.LogInformation("Restoring SQLite backup from {BackupPath} to {TargetPath}", backupPath, targetPath);

        if (!File.Exists(backupPath))
        {
            throw new FileNotFoundException($"Backup file not found: {backupPath}");
        }

        await Task.Run(() => File.Copy(backupPath, targetPath, overwrite: true), ct);

        logger.LogInformation("SQLite backup restored successfully.");
    }

    private string GetBackupDirectory(ConnectionParameters connectionParameters)
    {
        var currentDir = Directory.GetCurrentDirectory();
        return Path.Combine(currentDir, BackupDirectoryName);
    }

    private string GetDatabaseName(ConnectionParameters connectionParameters)
    {
        var dbName = connectionParameters.DatabaseName ?? "unknown";
        // Remove invalid file name characters
        return string.Join("_", dbName.Split(Path.GetInvalidFileNameChars()));
    }
}
