
using SqlMigration.Models;
using SqlMigration.Services;
using Npgsql;
using System.Data.SQLite;

namespace SqlMigration.Tests.Services;

public class DbConnectionFactoryTests
{
    [Fact]
    public void Create_ShouldReturnNpgsqlConnection_ForPostgreSql()
    {
        // Arrange
        var connectionParams = new ConnectionParameters("PostgreSql", null, "localhost", 5432, "test", "user", "password");
        var factory = new DbConnectionFactory(connectionParams);

        // Act
        var connection = factory.Create();

        // Assert
        Assert.IsType<NpgsqlConnection>(connection);
    }

    [Fact]
    public void Create_ShouldReturnSqliteConnection_ForSqLite()
    {
        // Arrange
        var connectionParams = new ConnectionParameters("SQLite", "DataSource=:memory:", null, null, null, null, null);
        var factory = new DbConnectionFactory(connectionParams);

        // Act
        var connection = factory.Create();

        // Assert
        Assert.IsType<SQLiteConnection>(connection);
    }

    [Fact]
    public void Create_ShouldThrowException_ForUnsupportedProvider()
    {
        // Arrange
        var connectionParams = new ConnectionParameters("Unknown", null, null, null, null, null, null);
        var factory = new DbConnectionFactory(connectionParams);

        // Act & Assert
        Assert.Throws<ArgumentException>(() => factory.Create());
    }

    [Fact]
    public void GetConnectionString_ShouldReturnCorrectString_ForPostgreSql()
    {
        // Arrange
        var connectionParams = new ConnectionParameters("PostgreSql", null, "localhost", 5432, "test", "user", "password");
        var factory = new DbConnectionFactory(connectionParams);

        // Act
        var connectionString = factory.ConnectionString;

        // Assert
        Assert.Equal("Host=localhost;Port=5432;Database=test;Username=user;Password=password;", connectionString);
    }
}
