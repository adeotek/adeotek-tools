using Microsoft.Extensions.Logging;
using SqlMigration.Models;
using SqlMigration.Repositories;

namespace SqlMigration.Services;

public class MigrationService(
    ISqlScriptsHelpers sqlScriptsHelpers,
    IMigrationHistoryRepositoryFactory migrationHistoryRepositoryFactory,
    ILogger<MigrationService> logger,
    bool isDryRun = false)
    : IMigrationService
{
    public async Task<int> RunAsync(string scriptsPath, ConnectionParameters connectionParameters, CancellationToken ct = default)
    {
        var migrationHistoryRepository = migrationHistoryRepositoryFactory.Create(connectionParameters);
        List<ScriptExecutionHistory> executedScripts;
        if (!await migrationHistoryRepository.IsHistoryTableCreatedAsync(ct))
        {
            if (!isDryRun)
            {
                logger.LogDebug("History table not found. Creating...");
                await migrationHistoryRepository.CreateHistoryTableAsync();
                logger.LogDebug("History table created.");
            }
            executedScripts = [];
        }
        else
        {
            executedScripts = (await migrationHistoryRepository.GetExecutedScriptsAsync(ct)).ToList();
        }

        ct.ThrowIfCancellationRequested();
        var targetDir = Path.GetFullPath(scriptsPath);
        var scriptFiles = sqlScriptsHelpers.ScanForSqlFiles(targetDir);
        logger.LogDebug("Found {FilesCount} script files in directory {TargetDir}", scriptFiles.Count, targetDir);

        var successCount = 0;
        var errorsCount = 0;
        var skipCount = 0;
        foreach (var scriptName in scriptFiles)
        {
            ct.ThrowIfCancellationRequested();
            var scriptFile = Path.Combine(targetDir, scriptName);
            var hash = sqlScriptsHelpers.CalculateHash(scriptFile);

            var executedScript = executedScripts.FirstOrDefault(s => s.ScriptFile == scriptName);
            if (executedScript != null)
            {
                if (executedScript.ScriptHash == hash)
                {
                    skipCount++;
                    logger.LogInformation("Skipping script {ScriptName} [{ScriptHash}]", scriptName, hash);
                    continue;
                }
            }

            try
            {
                if (isDryRun)
                {
                    logger.LogInformation("Dry run: would execute script {ScriptName} [{ScriptHash}]", scriptName, hash);
                    successCount++;
                    continue;
                }

                var scriptContent = await File.ReadAllTextAsync(scriptFile, ct);
                await sqlScriptsHelpers.ExecuteScriptAsync(scriptContent, connectionParameters);
                logger.LogInformation("Script {ScriptName} executed successfully [{ScriptHash}]", scriptName, hash);
                await migrationHistoryRepository.UpsertExecutedScriptAsync(new ScriptExecutionHistory
                {
                    ScriptFile = scriptName,
                    ScriptHash = hash,
                    ExecutedAt = DateTime.UtcNow
                });
            }
            catch (Exception ex)
            {
                errorsCount++;
                logger.LogError(ex, "Error executing script {ScriptName} [{ScriptHash}]", scriptName, hash);
            }
        }

        logger.LogInformation("Total scripts: {TotalCount} | Success: {SuccessCount} | Skipped: {SkipCount} | Errors: {ErrorsCount}",
            scriptFiles.Count, successCount, skipCount, errorsCount);
        return 0;
    }
}
