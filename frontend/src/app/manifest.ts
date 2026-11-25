import type { MetadataRoute } from "next";

// Web App Manifest (PRD F7): makes Monetra installable as a standalone app.
export default function manifest(): MetadataRoute.Manifest {
  return {
    name: "Monetra",
    short_name: "Monetra",
    description: "Kelola keuangan pribadi dan keluarga",
    lang: "id",
    start_url: "/dashboard",
    scope: "/",
    display: "standalone",
    background_color: "#f7f2eb",
    theme_color: "#f7f2eb",
    icons: [
      { src: "/icon/192", sizes: "192x192", type: "image/png", purpose: "any" },
      { src: "/icon/512", sizes: "512x512", type: "image/png", purpose: "any" },
      { src: "/icon/512", sizes: "512x512", type: "image/png", purpose: "maskable" },
    ],
  };
}
