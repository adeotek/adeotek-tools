using System.Net;

namespace BnrRatesExtractor.Tests;

/// <summary>
/// Minimal <see cref="HttpMessageHandler"/> stub returning a canned response.
/// </summary>
internal sealed class StubHttpMessageHandler(HttpStatusCode statusCode, string content) : HttpMessageHandler
{
    public Uri? LastRequestUri { get; private set; }

    protected override Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken)
    {
        LastRequestUri = request.RequestUri;

        return Task.FromResult(new HttpResponseMessage(statusCode)
        {
            Content = new StringContent(content),
            RequestMessage = request,
        });
    }
}
