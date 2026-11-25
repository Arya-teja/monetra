import { ImageResponse } from "next/og";

// PWA icons (PRD F7). Full-bleed square so Android can mask it to any shape;
// the "M" stays inside the central 80% "safe zone".
const BRAND = "#8b9a6e"; // PRD §9 primary
const SIZES = [192, 512];

export const contentType = "image/png";

export function generateImageMetadata() {
  return SIZES.map((size) => ({ id: String(size), size: { width: size, height: size }, contentType }));
}

export default async function Icon({ id }: { id: Promise<string> }) {
  const size = Number(await id);
  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          background: BRAND,
          color: "#ffffff",
          fontSize: size * 0.5,
          fontWeight: 700,
        }}
      >
        M
      </div>
    ),
    { width: size, height: size },
  );
}
