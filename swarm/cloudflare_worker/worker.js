/**
 * Cloudflare Worker: 5D Dark Web Private Exit Node (Store-and-Forward)
 *
 * This worker acts as a public entry point to queue Dark Web fetches.
 * It passes requests to the 5D mesh backend over a Cloudflare tunnel.
 *
 * Usage:
 * GET /darkweb?onion=http://...&address=<5d_address>
 */

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);

    if (url.pathname === "/darkweb") {
      const onion = url.searchParams.get("onion");
      const address = url.searchParams.get("address");

      if (!onion || !address) {
        return new Response("Missing 'onion' or 'address' query parameters.", { status: 400 });
      }

      // Backend mesh node routing URL via Cloudflare Tunnel
      // Update this to point to your actual backend API
      const backendUrl = `https://your-tunnel-id.cfargotunnel.com/antigravity/mesh/darkweb?onion=${encodeURIComponent(onion)}&address=${encodeURIComponent(address)}`;

      try {
        // Forward the request to the Go Backend to queue the scrape
        const backendResp = await fetch(backendUrl, { method: "POST" });

        if (!backendResp.ok) {
          return new Response(`Backend Error: ${backendResp.statusText}`, { status: backendResp.status });
        }

        const data = await backendResp.json();
        
        return new Response(JSON.stringify({
          message: "Request queued securely. The payload will be AES-256 encrypted with your 5D address and stored.",
          job_id: data.job_id,
          status: data.status
        }), {
          status: 202,
          headers: { "Content-Type": "application/json" }
        });

      } catch (err) {
        return new Response(`Tunnel Error: ${err.message}`, { status: 502 });
      }
    }

    return new Response("Not Found", { status: 404 });
  }
};
