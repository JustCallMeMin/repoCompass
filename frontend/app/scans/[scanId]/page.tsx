import { ScanClient } from "./ScanClient";

export default async function ScanPage({ params }: { params: Promise<{ scanId: string }> }) {
  const { scanId } = await params;
  return <ScanClient scanId={scanId} />;
}
