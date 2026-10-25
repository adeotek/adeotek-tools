using System.Net;
using BnrRatesExtractor;
using Xunit;

namespace BnrRatesExtractor.Tests;

public class BnrRatesParserTests
{
    // Trimmed excerpt of https://curs.bnr.ro/files/xml/years/nbrfxrates2024.xml (real values).
    private const string SampleXml = """
        <?xml version="1.0" encoding="utf-8"?>
        <DataSet xmlns="http://www.bnr.ro/xsd" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.bnr.ro/xsd nbrfxrates.xsd">
        	<Header>
        		<Publisher>National Bank of Romania</Publisher>
        		<PublishingDate>2024-12-31</PublishingDate>
        		<MessageType>DR</MessageType>
        	</Header>
        	<Body>
        		<Subject>Reference rates</Subject>
        		<OrigCurrency>RON</OrigCurrency>
        		<Cube date="2024-01-04">
        			<Rate currency="EUR">4.9730</Rate>
        			<Rate currency="HUF" multiplier="100">1.3105</Rate>
        			<Rate currency="USD">4.5395</Rate>
        			<Rate currency="XAU">298.7471</Rate>
        		</Cube>
        		<Cube date="2024-01-03">
        			<Rate currency="EUR">4.9731</Rate>
        			<Rate currency="HUF" multiplier="100">1.3054</Rate>
        			<Rate currency="USD">4.5504</Rate>
        			<Rate currency="XAU">300.6318</Rate>
        		</Cube>
        	</Body>
        </DataSet>
        """;

    [Fact]
    public void Parse_WithCurrency_ReturnsOnlyThatCurrencySortedByDate()
    {
        // Act
        var rates = BnrRatesParser.Parse(SampleXml, "USD", 2024);

        // Assert
        Assert.Equal(2, rates.Count);
        Assert.Equal(new DateOnly(2024, 1, 3), rates[0].Date);
        Assert.Equal(4.5504m, rates[0].Rate);
        Assert.Equal(new DateOnly(2024, 1, 4), rates[1].Date);
        Assert.Equal(4.5395m, rates[1].Rate);
        Assert.All(rates, r => Assert.Equal("USD", r.Currency));
    }

    [Fact]
    public void Parse_WithLowerCaseCurrency_MatchesCaseInsensitivelyAndNormalizes()
    {
        // Act
        var rates = BnrRatesParser.Parse(SampleXml, "eur", 2024);

        // Assert
        Assert.Equal(2, rates.Count);
        Assert.All(rates, r => Assert.Equal("EUR", r.Currency));
    }

    [Fact]
    public void Parse_WithMultiplier_PreservesMultiplier()
    {
        // Act
        var rates = BnrRatesParser.Parse(SampleXml, "HUF", 2024);

        // Assert
        Assert.Equal(2, rates.Count);
        Assert.All(rates, r => Assert.Equal(100, r.Multiplier));
        Assert.Equal(1.3054m, rates[0].Rate);
    }

    [Fact]
    public void Parse_WithRatesWithoutMultiplier_DefaultsToOne()
    {
        // Act
        var rates = BnrRatesParser.Parse(SampleXml, "XAU", 2024);

        // Assert
        Assert.All(rates, r => Assert.Equal(1, r.Multiplier));
    }

    [Fact]
    public void Parse_WithoutNamespace_UsesNoNamespace()
    {
        // Arrange
        const string noNamespace = """
            <DataSet>
              <Body>
                <Cube date="2024-01-03"><Rate currency="USD">4.5504</Rate></Cube>
              </Body>
            </DataSet>
            """;

        // Act
        var rates = BnrRatesParser.Parse(noNamespace, "USD", 2024);

        // Assert
        Assert.Single(rates);
        Assert.Equal(4.5504m, rates[0].Rate);
    }

    [Fact]
    public void Parse_WithCubeFromAnotherYear_IgnoresIt()
    {
        // Arrange
        const string mixedYears = """
            <DataSet xmlns="http://www.bnr.ro/xsd">
              <Body>
                <Cube date="2023-12-29"><Rate currency="USD">4.5000</Rate></Cube>
                <Cube date="2024-01-03"><Rate currency="USD">4.5504</Rate></Cube>
              </Body>
            </DataSet>
            """;

        // Act
        var rates = BnrRatesParser.Parse(mixedYears, "USD", 2024);

        // Assert
        Assert.Single(rates);
        Assert.Equal(new DateOnly(2024, 1, 3), rates[0].Date);
    }

    [Fact]
    public void Parse_WithDuplicateDate_KeepsSingleEntry()
    {
        // Arrange
        const string duplicated = """
            <DataSet xmlns="http://www.bnr.ro/xsd">
              <Body>
                <Cube date="2024-01-03"><Rate currency="USD">4.5504</Rate></Cube>
                <Cube date="2024-01-03"><Rate currency="USD">4.9999</Rate></Cube>
              </Body>
            </DataSet>
            """;

        // Act
        var rates = BnrRatesParser.Parse(duplicated, "USD", 2024);

        // Assert
        Assert.Single(rates);
        Assert.Equal(4.9999m, rates[0].Rate);
    }

    [Fact]
    public void Parse_WithInvalidRateValue_SkipsRate()
    {
        // Arrange
        const string badValues = """
            <DataSet xmlns="http://www.bnr.ro/xsd">
              <Body>
                <Cube date="2024-01-03">
                  <Rate currency="USD">not-a-number</Rate>
                  <Rate currency="USD">0</Rate>
                  <Rate currency="USD">4.5504</Rate>
                </Cube>
              </Body>
            </DataSet>
            """;

        // Act
        var rates = BnrRatesParser.Parse(badValues, "USD", 2024);

        // Assert
        Assert.Single(rates);
        Assert.Equal(4.5504m, rates[0].Rate);
    }

    [Fact]
    public void Parse_WithUnknownCurrency_ReturnsEmpty()
    {
        // Act
        var rates = BnrRatesParser.Parse(SampleXml, "ZZZ", 2024);

        // Assert
        Assert.Empty(rates);
    }

    [Fact]
    public void Parse_WithHtmlRedirectBody_ThrowsInvalidDataException()
    {
        // Arrange
        const string html = "<!doctype html><html><head><title>BNR</title></head><body></body></html>";

        // Act
        var ex = Assert.Throws<InvalidDataException>(() => BnrRatesParser.Parse(html, "USD", 2024));

        // Assert
        Assert.Contains("not valid XML", ex.Message);
    }

    [Fact]
    public void Parse_WithUnexpectedRoot_ThrowsInvalidDataException()
    {
        // Arrange
        const string otherXml = "<rss><channel /></rss>";

        // Act
        var ex = Assert.Throws<InvalidDataException>(() => BnrRatesParser.Parse(otherXml, "USD", 2024));

        // Assert
        Assert.Contains("not a BNR exchange-rate document", ex.Message);
    }

    [Fact]
    public void ToCsv_WithRates_WritesHeaderAndInvariantRows()
    {
        // Arrange
        RateData[] rates =
        [
            new("USD", new DateOnly(2024, 1, 3), 4.5504m, 1),
            new("HUF", new DateOnly(2024, 1, 4), 1.3m, 100),
        ];

        // Act
        var csv = BnrRatesParser.ToCsv(rates);

        // Assert
        Assert.Equal(
            "Currency,Date,Rate,Multiplier\nUSD,2024-01-03,4.5504,1\nHUF,2024-01-04,1.3000,100\n",
            csv);
    }

    [Fact]
    public void ToCsv_WithNoRates_WritesHeaderOnly()
    {
        // Act
        var csv = BnrRatesParser.ToCsv([]);

        // Assert
        Assert.Equal("Currency,Date,Rate,Multiplier\n", csv);
    }

    [Theory]
    [InlineData(2024, "https://curs.bnr.ro/files/xml/years/nbrfxrates2024.xml")]
    [InlineData(2005, "https://curs.bnr.ro/files/xml/years/nbrfxrates2005.xml")]
    public void BuildYearUrl_WithYear_ReturnsCursHostUrl(int year, string expected)
    {
        // Act
        var url = BnrRatesParser.BuildYearUrl(year);

        // Assert
        Assert.Equal(expected, url);
    }

    [Theory]
    [InlineData("USD", true)]
    [InlineData("usd", true)]
    [InlineData("XAU", true)]
    [InlineData("US", false)]
    [InlineData("USDD", false)]
    [InlineData("US1", false)]
    [InlineData("", false)]
    [InlineData(null, false)]
    public void IsValidCurrencyCode_WithCode_ValidatesThreeAsciiLetters(string? code, bool expected)
    {
        // Act
        var isValid = BnrRatesParser.IsValidCurrencyCode(code);

        // Assert
        Assert.Equal(expected, isValid);
    }

    [Fact]
    public async Task FetchYearXmlAsync_WithSuccessResponse_ReturnsBody()
    {
        // Arrange
        var handler = new StubHttpMessageHandler(HttpStatusCode.OK, SampleXml);
        using var httpClient = new HttpClient(handler);

        // Act
        var xml = await BnrRatesParser.FetchYearXmlAsync(
            httpClient, 2024, TestContext.Current.CancellationToken);

        // Assert
        Assert.Equal(SampleXml, xml);
        Assert.Equal("https://curs.bnr.ro/files/xml/years/nbrfxrates2024.xml", handler.LastRequestUri?.ToString());
    }

    [Fact]
    public async Task FetchYearXmlAsync_WithErrorResponse_ThrowsHttpRequestException()
    {
        // Arrange
        var handler = new StubHttpMessageHandler(HttpStatusCode.ServiceUnavailable, string.Empty);
        using var httpClient = new HttpClient(handler);

        // Act
        var ex = await Assert.ThrowsAsync<HttpRequestException>(
            () => BnrRatesParser.FetchYearXmlAsync(httpClient, 2024, TestContext.Current.CancellationToken));

        // Assert
        Assert.Contains("HTTP 503", ex.Message);
        Assert.Contains("nbrfxrates2024.xml", ex.Message);
    }
}
