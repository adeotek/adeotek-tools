using Microsoft.Extensions.Logging;
using SqlMigration.Models;
using SqlMigration.Repositories;

namespace SqlMigration.Services;

public class MigrationService : IMigrationService
{
    private readonly ISqlScriptsHelpers _sqlScriptsHelpers;
    private readonly IScriptExecutor _scriptExecutor;
    private readonly IMigrationHistoryRepositoryFactory _migrationHistoryRepositoryFactory;
    private readonly ILogger<MigrationService> _logger;

    public MigrationService(
        ISqlScriptsHelpers sqlScriptsHelpers,
        IScriptExecutor scriptExecutor,
        IMigrationHistoryRepositoryFactory migrationHistoryRepositoryFactory,
        ILogger<MigrationService> logger)
    {
        _sqlScriptsHelpers = sqlScriptsHelpers;
        _scriptExecutor = scriptExecutor;
        _migrationHistoryRepositoryFactory = migrationHistoryRepositoryFactory;
        _logger = logger;
    }

    public async Task<int> RunAsync(string scriptsPath, string connectionString)
    {
        var migrationRepository = _migrationHistoryRepositoryFactory.Create(connectionString);

        if (!await migrationRepository.IsHistoryTableCreated())
        {
            _logger.LogInformation("History table not found. Creating...");
            await migrationRepository.CreateHistoryTable();
            _logger.LogInformation("History table created.");
        }

        var executedScripts = (await migrationRepository.GetExecutedScripts()).ToList();
        var scriptFiles = _sqlScriptsHelpers.ScanForSqlFiles(scriptsPath).ToList();

        _logger.LogInformation($"Found {scriptFiles.Count} script files.");

        foreach (var scriptFile in scriptFiles)
        {
            var scriptName = Path.GetFileName(scriptFile);
            var hash = _sqlScriptsHelpers.CalculateHash(scriptFile);

            var executedScript = executedScripts.FirstOrDefault(s => s.ScriptName == scriptName);

            if (executedScript != null)
            {
                if (executedScript.Hash == hash)
                {
                    _logger.LogInformation($"Skipping script {scriptName} (already executed and hash is unchanged).");
                    continue;
                }

                _logger.LogError($"Script {scriptName} has changed since it was last executed. Halting execution.");
                return 1;
            }

            try
            {
                _logger.LogInformation($"Executing script {scriptName}...");
                var scriptContent = await File.ReadAllTextAsync(scriptFile);
                await _scriptExecutor.ExecuteScript(connectionString, scriptContent);
                _logger.LogInformation($"Script {scriptName} executed successfully.");

                await migrationRepository.AddExecutedScript(new MigrationHistory
                {
                    ScriptName = scriptName,
                    Hash = hash,
                    ExecutedAt = DateTime.UtcNow
                });
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, $"Error executing script {scriptName}. Halting execution.");
                return 1;
            }
        }

        _logger.LogInformation("All scripts executed successfully.");
        return 0;
    }
}
