using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;
using SqlMigration.Factories;
using SqlMigration.Services;

var configuration = new ConfigurationBuilder()
    .AddEnvironmentVariables(prefix: "SQL_MIGRATE_")
    .AddCommandLine(args)
    .Build();

var loggerFactory = LoggerFactory.Create(builder =>
{
    builder.AddConsole();
});

var logger = loggerFactory.CreateLogger<MigrationService>();
var fileScanner = new FileScanner();
var hashCalculator = new HashCalculator();
var scriptExecutor = new ScriptExecutor();
var migrationHistoryRepositoryFactory = new MigrationHistoryRepositoryFactory();
var migrationService = new MigrationService(fileScanner, hashCalculator, scriptExecutor, migrationHistoryRepositoryFactory, logger);

var scriptsPath = configuration["scripts-path"];
if (string.IsNullOrEmpty(scriptsPath))
{
    logger.LogError("The --scripts-path option is required.");
    return 1;
}

return await migrationService.RunAsync(scriptsPath, GetConnectionString(configuration));

static string GetConnectionString(IConfiguration configuration)
{
    var connectionString = configuration["connection-string"];
    if (!string.IsNullOrEmpty(connectionString))
    {
        return connectionString;
    }

    var host = configuration["host"];
    var port = configuration["port"];
    var database = configuration["database"];
    var user = configuration["user"];
    var password = configuration["password"];

    return $"Server={host},{port};Database={database};User Id={user};Password={password};";
}

public partial class Program { }