using System.Globalization;
using System.Text;
using System.Xml;
using System.Xml.Linq;

namespace BnrRatesExtractor;

/// <summary>
/// A single published reference rate.
/// </summary>
/// <param name="Currency">ISO 4217 currency code, uppercase (e.g. USD).</param>
/// <param name="Date">The date the rate was published for.</param>
/// <param name="Rate">The quoted rate, expressed in RON per <paramref name="Multiplier"/> units of <paramref name="Currency"/>.</param>
/// <param name="Multiplier">Number of currency units the quoted rate applies to (1 unless BNR publishes e.g. 100 for HUF/JPY).</param>
public readonly record struct RateData(string Currency, DateOnly Date, decimal Rate, int Multiplier);

public static class BnrRatesParser
{
    /// <summary>First year for which BNR publishes an XML file.</summary>
    public const int FirstSupportedYear = 2005;

    private const string BaseUrl = "https://curs.bnr.ro";

    public static string BuildYearUrl(int year) => $"{BaseUrl}/files/xml/years/nbrfxrates{year}.xml";

    public static bool IsValidCurrencyCode(string? currencyCode) =>
        currencyCode is { Length: 3 } && currencyCode.All(char.IsAsciiLetter);

    /// <summary>
    /// Downloads the yearly XML file for <paramref name="year"/>.
    /// </summary>
    /// <exception cref="HttpRequestException">The server did not return a success status code.</exception>
    public static async Task<string> FetchYearXmlAsync(HttpClient httpClient, int year, CancellationToken cancellationToken = default)
    {
        ArgumentNullException.ThrowIfNull(httpClient);

        var url = BuildYearUrl(year);
        using var response = await httpClient.GetAsync(url, cancellationToken).ConfigureAwait(false);
        if (!response.IsSuccessStatusCode)
        {
            throw new HttpRequestException(
                $"Failed to download {url}: HTTP {(int)response.StatusCode} {response.ReasonPhrase}".TrimEnd());
        }

        return await response.Content.ReadAsStringAsync(cancellationToken).ConfigureAwait(false);
    }

    /// <summary>
    /// Extracts the rates for <paramref name="currencyCode"/> published in <paramref name="year"/>,
    /// sorted ascending by date.
    /// </summary>
    /// <exception cref="InvalidDataException">The content is not a BNR exchange-rate document.</exception>
    public static IReadOnlyList<RateData> Parse(string xmlContent, string currencyCode, int year)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(xmlContent);
        ArgumentException.ThrowIfNullOrWhiteSpace(currencyCode);

        var doc = LoadDocument(xmlContent);
        var ns = doc.Root!.Name.Namespace;
        var currency = currencyCode.Trim().ToUpperInvariant();

        var rates = new Dictionary<DateOnly, RateData>();

        foreach (var cube in doc.Descendants(ns + "Cube"))
        {
            if (!TryParseDate(cube.Attribute("date")?.Value, out var date) || date.Year != year)
            {
                continue;
            }

            foreach (var rateElement in cube.Elements(ns + "Rate"))
            {
                if (!string.Equals(rateElement.Attribute("currency")?.Value, currency, StringComparison.OrdinalIgnoreCase))
                {
                    continue;
                }

                if (!decimal.TryParse(rateElement.Value, NumberStyles.Float, CultureInfo.InvariantCulture, out var rate)
                    || rate <= 0)
                {
                    continue;
                }

                rates[date] = new RateData(currency, date, rate, ParseMultiplier(rateElement.Attribute("multiplier")?.Value));
            }
        }

        return rates.Values.OrderBy(r => r.Date).ToArray();
    }

    /// <summary>
    /// Renders <paramref name="rates"/> as CSV (header plus one row per rate, 'yyyy-MM-dd' dates,
    /// invariant 4-decimal rates).
    /// </summary>
    public static string ToCsv(IEnumerable<RateData> rates)
    {
        ArgumentNullException.ThrowIfNull(rates);

        var csv = new StringBuilder();
        csv.Append("Currency,Date,Rate,Multiplier\n");

        foreach (var rate in rates)
        {
            csv.Append(rate.Currency)
                .Append(',')
                .Append(rate.Date.ToString("yyyy-MM-dd", CultureInfo.InvariantCulture))
                .Append(',')
                .Append(rate.Rate.ToString("F4", CultureInfo.InvariantCulture))
                .Append(',')
                .Append(rate.Multiplier.ToString(CultureInfo.InvariantCulture))
                .Append('\n');
        }

        return csv.ToString();
    }

    private static XDocument LoadDocument(string xmlContent)
    {
        XDocument doc;
        try
        {
            doc = XDocument.Parse(xmlContent);
        }
        catch (XmlException ex)
        {
            throw new InvalidDataException(
                $"Response is not valid XML: {ex.Message} The BNR XML location may have changed.", ex);
        }

        if (doc.Root is null || doc.Root.Name.LocalName != "DataSet")
        {
            throw new InvalidDataException(
                $"Response is not a BNR exchange-rate document (unexpected root element '{doc.Root?.Name.LocalName}'). "
                + "The BNR XML location may have changed.");
        }

        return doc;
    }

    private static bool TryParseDate(string? value, out DateOnly date)
    {
        date = default;
        return !string.IsNullOrWhiteSpace(value)
            && DateOnly.TryParseExact(value, "yyyy-MM-dd", CultureInfo.InvariantCulture, DateTimeStyles.None, out date);
    }

    private static int ParseMultiplier(string? value) =>
        int.TryParse(value, NumberStyles.Integer, CultureInfo.InvariantCulture, out var multiplier) && multiplier > 0
            ? multiplier
            : 1;
}
