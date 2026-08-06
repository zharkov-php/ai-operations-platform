import { RecommendationDetail } from "@/components/dashboard/recommendation-detail";
export default async function RecommendationPage({ params }: { params: Promise<{ recommendationId: string }> }) { return <RecommendationDetail id={(await params).recommendationId} />; }
