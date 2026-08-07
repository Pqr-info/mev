/**
 * Sovereign-27 Tor Exit Node & 5D Secure Storage Fetcher
 * 
 * This Cloudflare Worker acts as the public-facing gateway for the Mesh.
 * It accepts requests from users for Dark Web (.onion) resources, routes them 
 * to the Sovereign-27 Tor Proxy (darkweb_proxy.go), and fetches the securely 
 * encrypted payload from the Hetzner Z-Drive via the mesh API.
 */

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);

    // Only accept POST requests for submitting darkweb scraping jobs
    if (url.pathname === "/api/darkweb/fetch" && request.method === "POST") {
      try {
        const reqBody = await request.json();
        
        // 1. Verify the 5D address format
        if (!reqBody.client_5d || !reqBody.target_onion) {
          return new Response(JSON.stringify({ error: "Missing client_5d or target_onion" }), {
            status: 400,
            headers: { "Content-Type": "application/json" }
          });
        }

        // 2. Forward the request to the Sovereign-27 mesh (e.g. your external IP/API Gateway)
        // Note: env.MESH_API_ENDPOINT must be configured in Cloudflare secrets/vars.
        const meshEndpoint = `${env.MESH_API_ENDPOINT}/mesh/darkweb/queue`;
        
        const meshReq = await fetch(meshEndpoint, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${env.MESH_WORKER_SECRET}`
          },
          body: JSON.stringify({
            client_5d: reqBody.client_5d,
            target_onion: reqBody.target_onion
          })
        });

        if (!meshReq.ok) {
          return new Response(JSON.stringify({ error: "Mesh rejected the proxy request" }), {
            status: 502,
            headers: { "Content-Type": "application/json" }
          });
        }

        const meshRes = await meshReq.json();
        
        // Return the Job ID to the user. The file is being encrypted into the Z:\ drive.
        return new Response(JSON.stringify({
          status: "queued",
          job_id: meshRes.job_id,
          message: "Request accepted. File is being encrypted and deposited to Z-Drive.",
          download_url: `${url.origin}/api/darkweb/download/${meshRes.job_id}`
        }), {
          status: 202,
          headers: { "Content-Type": "application/json" }
        });

      } catch (err) {
        return new Response(JSON.stringify({ error: err.message }), { status: 500 });
      }
    }

    // Download endpoint for retrieving the encrypted file from the Z-Drive
    if (url.pathname.startsWith("/api/darkweb/download/") && request.method === "GET") {
      const jobId = url.pathname.split("/").pop();
      if (!jobId) return new Response("Missing Job ID", { status: 400 });

      // Fetch the encrypted file from the Sovereign-27 mesh node which reads it from Z:\
      const fileEndpoint = `${env.MESH_API_ENDPOINT}/mesh/darkweb/retrieve/${jobId}`;
      const fileReq = await fetch(fileEndpoint, {
        method: "GET",
        headers: {
          "Authorization": `Bearer ${env.MESH_WORKER_SECRET}`
        }
      });

      if (!fileReq.ok) {
        return new Response("File not found or still processing", { status: fileReq.status });
      }

      // Stream the encrypted payload back to the user
      return new Response(fileReq.body, {
        status: 200,
        headers: {
          "Content-Type": "application/octet-stream",
          "Content-Disposition": `attachment; filename="${jobId}.enc"`,
          "X-Sovereign-Liability": "Zero-Knowledge" // Absolves the mesh of DMCA
        }
      });
    }

    return new Response("Sovereign-27 Tor Exit Gateway Active", { status: 200 });
  }
};
