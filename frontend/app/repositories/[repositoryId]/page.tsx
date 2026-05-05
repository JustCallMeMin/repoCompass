import { RepositoryClient } from "./RepositoryClient";

export default async function RepositoryPage({ params }: { params: Promise<{ repositoryId: string }> }) {
  const { repositoryId } = await params;
  return <RepositoryClient repositoryId={repositoryId} />;
}
