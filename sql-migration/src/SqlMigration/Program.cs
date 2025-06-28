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

var logLevel = configuration.GetValue<string>(Constants.VerboseArgumentName, null) == null
    ? LogLevel.Information
    : LogLevel.Debug;
var loggerFactory = LoggerFactory.Create(builder =>
{
    builder
        .AddConsole(options => options.FormatterName = "SqlMigrationFormatter")
        .AddConsoleFormatter<CustomConsoleFormatter, CustomConsoleFormatterOptions>()
        .SetMinimumLevel(logLevel);
});

var logger = loggerFactory.CreateLogger<MigrationService>();
var sqlScriptsHelpers = new SqlScriptsHelpers();
var scriptExecutor = new ScriptExecutor();
var migrationHistoryRepositoryFactory = new MigrationHistoryRepositoryFactory();
var migrationService = new MigrationService(sqlScriptsHelpers, scriptExecutor, migrationHistoryRepositoryFactory, logger);

var targetPath = configuration[Constants.ScriptsPathArgumentName];
if (string.IsNullOrEmpty(targetPath))
{
    logger.LogError($"The {Constants.ScriptsPathArgumentName} option is required.");
    return 1;
}
logger.LogInformation("Starting SQL migration with target path: {TargetPath}", targetPath);

return await migrationService.RunAsync(targetPath, configuration.GetConnectionString());
