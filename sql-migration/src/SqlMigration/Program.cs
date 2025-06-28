using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;
using SqlMigration.Extensions;
using SqlMigration.Models;
using SqlMigration.Repositories;
using SqlMigration.Services;

var configuration = new ConfigurationBuilder()
    .AddEnvironmentVariables(prefix: Constants.EnvironmentVariablesPrefix)
    .AddCommandLine(args)
    .Build();

var loggerFactory = LoggerFactory.Create(builder =>
{
    builder.AddConsole();
});

var logger = loggerFactory.CreateLogger<MigrationService>();
var sqlScriptsHelpers = new SqlScriptsHelpers();
var scriptExecutor = new ScriptExecutor();
var migrationHistoryRepositoryFactory = new MigrationHistoryRepositoryFactory();
var migrationService = new MigrationService(sqlScriptsHelpers, scriptExecutor, migrationHistoryRepositoryFactory, logger);

var scriptsPath = configuration[Constants.ScriptsPathArgumentName];
if (string.IsNullOrEmpty(scriptsPath))
{
    logger.LogError($"The {Constants.ScriptsPathArgumentName} option is required.");
    return 1;
}

return await migrationService.RunAsync(scriptsPath, configuration.GetConnectionString());
