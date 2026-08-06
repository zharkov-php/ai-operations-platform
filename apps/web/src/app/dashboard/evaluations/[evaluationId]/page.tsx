import { EvaluationDetail } from "@/components/dashboard/evaluation-detail";

export default async function EvaluationPage({ params }: { params: Promise<{ evaluationId: string }> }) {
  return <EvaluationDetail id={(await params).evaluationId} />;
}
