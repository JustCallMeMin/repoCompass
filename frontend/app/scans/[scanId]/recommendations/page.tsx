import { ScanClient } from "../ScanClient";

export default async function ScanRecommendationsPage({ params }: { params: Promise<{ scanId: string }> }) {
  const { scanId } = await params;
  return <ScanClient scanId={scanId} view="recommendations" />;
}
